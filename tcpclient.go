package go_eip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	tcpTimeout     = 10 * time.Second
	tcpIdleTimeout = 60 * time.Second
	tcpMaxLength   = 1024
	isoTCP         = 44818

	connectionTypeBasic = 3
)

type CIPType struct {
	BitCount uint8
	TypeName string
}

type tcpPackager struct {
	VendorID               uint16
	SessionHandle          uint32
	ProcessorSlot          uint8
	Context                uint64
	ContextPointer         uint8
	SerialNumber           uint16
	OriginatorSerialNumber uint32
	OTNetworkConnectionID  uint32
	SequenceCounter        uint16
	Offset                 uint32
	ProgramNames           []string
	contextMap             map[uint8]uint64
	cipTypeMap             map[uint8]CIPType
	knownTags              map[string]uint8
}

type tcpTransporter struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
	Logger      *log.Logger

	mu           sync.Mutex
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time
}

type TCPClientHandler struct {
	tcpPackager
	tcpTransporter
}

func NewTCPClientHandler(address string, slot int) *TCPClientHandler {
	h := &TCPClientHandler{}
	h.Address = address
	if !strings.Contains(address, ":") {
		h.Address = address + ":" + strconv.Itoa(isoTCP)
	}
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout

	h.VendorID = 1
	h.ProcessorSlot = uint8(slot)
	h.SessionHandle = 0x0000
	h.Context = 0x00
	h.ContextPointer = 0
	h.SerialNumber = 0
	h.OriginatorSerialNumber = 42
	h.SequenceCounter = 1
	h.Offset = 0
	h.ProgramNames = []string{""}
	h.knownTags = make(map[string]uint8)
	h.contextMap = map[uint8]uint64{
		0:   0x6572276557,
		1:   0x6f6e,
		2:   0x676e61727473,
		3:   0x737265,
		4:   0x6f74,
		5:   0x65766f6c,
		6:   0x756f59,
		7:   0x776f6e6b,
		8:   0x656874,
		9:   0x73656c7572,
		10:  0x646e61,
		11:  0x6f73,
		12:  0x6f64,
		13:  0x49,
		14:  0x41,
		15:  0x6c6c7566,
		16:  0x74696d6d6f63,
		17:  0x7327746e656d,
		18:  0x74616877,
		19:  0x6d2749,
		20:  0x6b6e696874,
		21:  0x676e69,
		22:  0x666f,
		23:  0x756f59,
		24:  0x746e646c756f77,
		25:  0x746567,
		26:  0x73696874,
		27:  0x6d6f7266,
		28:  0x796e61,
		29:  0x726568746f,
		30:  0x797567,
		31:  0x49,
		32:  0x7473756a,
		33:  0x616e6e6177,
		34:  0x6c6c6574,
		35:  0x756f79,
		36:  0x776f68,
		37:  0x6d2749,
		38:  0x676e696c656566,
		39:  0x6174746f47,
		40:  0x656b616d,
		41:  0x756f79,
		42:  0x7265646e75,
		43:  0x646e617473,
		44:  0x726576654e,
		45:  0x616e6e6f67,
		46:  0x65766967,
		47:  0x756f79,
		48:  0x7075,
		49:  0x726576654e,
		50:  0x616e6e6f67,
		51:  0x74656c,
		52:  0x756f79,
		53:  0x6e776f64,
		54:  0x726576654e,
		55:  0x616e6e6f67,
		56:  0x6e7572,
		57:  0x646e756f7261,
		58:  0x646e61,
		59:  0x747265736564,
		60:  0x756f79,
		61:  0x726576654e,
		62:  0x616e6e6f67,
		63:  0x656b616d,
		64:  0x756f79,
		65:  0x797263,
		66:  0x726576654e,
		67:  0x616e6e6f67,
		68:  0x796173,
		69:  0x657962646f6f67,
		70:  0x726576654e,
		71:  0x616e6e6f67,
		72:  0x6c6c6574,
		73:  0x61,
		74:  0x65696c,
		75:  0x646e61,
		76:  0x74727568,
		77:  0x756f79,
		78:  0x6576276557,
		79:  0x6e776f6e6b,
		80:  0x68636165,
		81:  0x726568746f,
		82:  0x726f66,
		83:  0x6f73,
		84:  0x676e6f6c,
		85:  0x72756f59,
		86:  0x73277472616568,
		87:  0x6e656562,
		88:  0x676e69686361,
		89:  0x747562,
		90:  0x657227756f59,
		91:  0x6f6f74,
		92:  0x796873,
		93:  0x6f74,
		94:  0x796173,
		95:  0x7469,
		96:  0x656469736e49,
		97:  0x6577,
		98:  0x68746f62,
		99:  0x776f6e6b,
		100: 0x732774616877,
		101: 0x6e656562,
		102: 0x676e696f67,
		103: 0x6e6f,
		104: 0x6557,
		105: 0x776f6e6b,
		106: 0x656874,
		107: 0x656d6167,
		108: 0x646e61,
		109: 0x6572276577,
		110: 0x616e6e6f67,
		111: 0x79616c70,
		112: 0x7469,
		113: 0x646e41,
		114: 0x6669,
		115: 0x756f79,
		116: 0x6b7361,
		117: 0x656d,
		118: 0x776f68,
		119: 0x6d2749,
		120: 0x676e696c656566,
		121: 0x74276e6f44,
		122: 0x6c6c6574,
		123: 0x656d,
		124: 0x657227756f79,
		125: 0x6f6f74,
		126: 0x646e696c62,
		127: 0x6f74,
		128: 0x656573,
		129: 0x726576654e,
		130: 0x616e6e6f67,
		131: 0x65766967,
		132: 0x756f79,
		133: 0x7075,
		134: 0x726576654e,
		135: 0x616e6e6f67,
		136: 0x74656c,
		137: 0x756f79,
		138: 0x6e776f64,
		139: 0x726576654e,
		140: 0x6e7572,
		141: 0x646e756f7261,
		142: 0x646e61,
		143: 0x747265736564,
		144: 0x756f79,
		145: 0x726576654e,
		146: 0x616e6e6f67,
		147: 0x656b616d,
		148: 0x756f79,
		149: 0x797263,
		150: 0x726576654e,
		151: 0x616e6e6f67,
		152: 0x796173,
		153: 0x657962646f6f67,
		154: 0x726576654e,
		155: 0xa680e2616e6e6f67,
	}
	h.cipTypeMap = map[uint8]CIPType{
		160: CIPType{88, "STRUCT"},
		193: CIPType{1, "BOOL"},
		194: CIPType{1, "SINT"},
		195: CIPType{2, "INT"},
		196: CIPType{4, "DINT"},
		197: CIPType{8, "LINT"},
		198: CIPType{1, "USINT"},
		199: CIPType{2, "UINT"},
		200: CIPType{4, "UDINT"},
		201: CIPType{8, "LWORD"},
		202: CIPType{4, "REAL"},
		203: CIPType{8, "LREAL"},
		211: CIPType{4, "DWORD"},
		218: CIPType{0, "STRING"},
	}
	return h
}

func (t *tcpPackager) GetBitCount(s uint8) CIPType {
	if cip, ok := t.cipTypeMap[s]; ok {
		return cip
	}
	return CIPType{}
}
func (t *tcpPackager) SetSessionHandle(s uint32) {
	t.SessionHandle = s
}
func (t *tcpPackager) SetOffset(s uint32) {
	t.Offset = s
}
func (t *tcpPackager) IncreaseOffset(s uint32) {
	t.Offset += s
}
func (t *tcpPackager) SetNetworkConnectionID(s uint32) {
	t.OTNetworkConnectionID = s
}
func (t *tcpPackager) GetProgramName() []string {
	return t.ProgramNames
}
func (t *tcpPackager) AddProgramName(s string) []string {
	for _, p := range t.ProgramNames {
		if p == s {
			return t.ProgramNames
		}
	}
	t.ProgramNames = append(t.ProgramNames, s)
	return t.ProgramNames
}
func (t *tcpPackager) FillKnownTags(s string, v uint8) {
	t.knownTags[s] = v
}

func (t *tcpPackager) Verify(request []byte, response []byte) (err error) {
	return
}

func (t *tcpPackager) TagNameParser(tag string, offset int) (string, string, int) {
	base := tag
	ind := 0
	if strings.HasSuffix(tag, "]") {
		pos := strings.LastIndex(tag, "[")
		base = tag[:pos]
		temp := tag[pos:]
		s := strings.Split(temp[1:len(temp)-1], ",")
		if len(s) == 1 {
			ind, _ = strconv.Atoi(temp[1 : len(temp)-1])
		} else {

		}
	} else {
		tagSplit := strings.Split(tag, ".")
		if v, e := strconv.ParseInt(tagSplit[len(tagSplit)-1], 10, 8); e == nil {
			return tag, strings.Join(tagSplit[:len(tagSplit)-1], "."), int(v)
		}

		return tag, base, ind
	}
	return tag, base, ind
}
func (t *tcpPackager) BuildEIPHeader(tagIOI []byte) []byte {
	buf := new(bytes.Buffer)

	if t.ContextPointer == 155 {
		t.ContextPointer = 0
	}
	binary.Write(buf, binary.LittleEndian, struct {
		EIPCommand         uint16
		EIPLength          uint16
		EIPSessionHandle   uint32
		EIPStatus          uint32
		EIPContext         uint64
		EIPOptions         uint32
		EIPInterfaceHandle uint32
		EIPTimeout         uint16
		EIPItemCount       uint16
		EIPItem1ID         uint16
		EIPItem1Length     uint16
		EIPItem1           uint32
		EIPItem2ID         uint16
		EIPItem2Length     uint16
		EIPSequence        uint16
	}{
		0x70,
		22 + uint16(len(tagIOI)),
		t.SessionHandle,
		0,
		t.contextMap[t.ContextPointer],
		0,
		0,
		0,
		2,
		0xA1,
		4,
		t.OTNetworkConnectionID,
		0xB1,
		uint16(len(tagIOI)) + 2,
		t.SequenceCounter,
	})
	t.SequenceCounter += 1
	t.SequenceCounter = t.SequenceCounter % 10000

	buf.Write(tagIOI)

	return buf.Bytes()
}

func (t *tcpPackager) BuildRegisterSessionRequest() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		EIPCommand         uint16
		EIPLength          uint16
		EIPSessionHandle   uint32
		EIPStatus          uint32
		EIPContext         uint64
		EIPOptions         uint32
		EIPProtocolVersion uint16
		EIPOptionFlag      uint16
	}{
		0x0065,
		0x0004,
		0,
		0,
		0,
		0,
		1,
		0,
	})
	return buf.Bytes()
}
func (t *tcpPackager) BuildUnregisterSessionRequest() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		EIPCommand       uint16
		EIPLength        uint16
		EIPSessionHandle uint32
		EIPStatus        uint32
		EIPContext       uint64
		EIPOptions       uint32
	}{
		0x66, 0x00, t.SessionHandle,
		0x00, t.Context, 0x00,
	})
	return buf.Bytes()
}
func (t *tcpPackager) BuildForwardOpenRequest() []byte {
	forwardOpen := t.buildCIPForwardOpen()
	dH := t.buildEIPSendRRDataHeader(len(forwardOpen))
	buf := new(bytes.Buffer)
	buf.Write(append(dH, forwardOpen...))
	return buf.Bytes()
}
func (t *tcpPackager) BuildForwardCloseRequest() []byte {
	forwardClose := t.buildCIPForwardClose()
	dH := t.buildEIPSendRRDataHeader(len(forwardClose))
	buf := new(bytes.Buffer)
	buf.Write(append(dH, forwardClose...))
	return buf.Bytes()
}
func (t *tcpPackager) BuildTagListRequest(programName string) []byte {
	buf := new(bytes.Buffer)
	pathSegment := new(bytes.Buffer)
	attributes := new(bytes.Buffer)

	if programName != "" {
		binary.Write(pathSegment, binary.LittleEndian, struct{ H, L uint8 }{0x91, uint8(len(programName))})
		pathSegment.Write([]byte(programName))
		if len(programName)%2 != 0 {
			pathSegment.Write([]byte{0x00})
		}
	}
	binary.Write(pathSegment, binary.LittleEndian, struct{ H uint16 }{0x6B20})

	if t.Offset < 256 {
		binary.Write(pathSegment, binary.LittleEndian, struct{ H, L uint8 }{
			0x24, uint8(t.Offset),
		})
	} else {
		binary.Write(pathSegment, binary.LittleEndian, struct{ H, L uint16 }{
			0x25, uint16(t.Offset),
		})
	}

	binary.Write(buf, binary.LittleEndian, struct {
		Service        uint8
		PathSegmentLen uint8
	}{
		0x55,
		uint8(pathSegment.Len() / 2),
	})
	binary.Write(attributes, binary.LittleEndian, struct {
		AttributeCount uint16
		SymbolType     uint16
		ByteCount      uint16
		SymbolName     uint16
	}{03, 02, 07, 01})
	buf.Write(pathSegment.Bytes())
	buf.Write(attributes.Bytes())

	return t.BuildEIPHeader(buf.Bytes())
}
func (t *tcpPackager) BuildPartialReadRequest(tag, baseTag string) []byte {
	return t.BuildEIPHeader(t.buildPartialReadIOI(t.buildTagIOI(tag, false), 1))
}
func (t *tcpPackager) BuildReadIOIRequest(tag string, isBoolArray bool, elements int) []byte {
	buf := new(bytes.Buffer)
	tagIOI := t.buildTagIOI(tag, isBoolArray)
	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x4c, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)
	binary.Write(buf, binary.LittleEndian, struct{ a uint16 }{uint16(elements)})

	return t.BuildEIPHeader(buf.Bytes())
}
func (t *tcpPackager) buildWriteIOT(tagIOI []byte, value interface{}, dataType uint8) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x4d, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)

	if dataType == 160 {
		binary.Write(buf, binary.LittleEndian, struct {
			a, b uint8
			c, d uint16
		}{dataType, 0x02, 0x0fCE, 1})
	} else {
		binary.Write(buf, binary.LittleEndian, struct {
			a, b uint8
			c    uint16
		}{dataType, 0x0, 1})
	}

	switch dataType {
	case 193:
		v, e := strconv.ParseBool(fmt.Sprintf("%v", value))
		if e != nil {
			log.Println(e)
		}
		binary.Write(buf, binary.LittleEndian, struct{ a bool }{v})
	case 194, 195, 196, 197, 198, 199, 200, 201, 211:
		bitCount := t.cipTypeMap[dataType].BitCount * 8
		v, e := strconv.ParseInt(fmt.Sprintf("%v", value), 10, int(bitCount))
		if e != nil {
			log.Println(e)
		}
		switch bitCount {
		case 8:
			binary.Write(buf, binary.LittleEndian, struct{ a uint8 }{uint8(v)})
		case 16:
			binary.Write(buf, binary.LittleEndian, struct{ a uint16 }{uint16(v)})
		case 32:
			binary.Write(buf, binary.LittleEndian, struct{ a uint32 }{uint32(v)})
		case 64:
			binary.Write(buf, binary.LittleEndian, struct{ a uint64 }{uint64(v)})
		}

	case 202, 203:
		bitCount := t.cipTypeMap[dataType].BitCount * 8
		v, e := strconv.ParseFloat(fmt.Sprintf("%v", value), int(bitCount))
		if e != nil {
			log.Println(e)
		}
		switch bitCount {
		case 32:
			binary.Write(buf, binary.LittleEndian, struct{ a float32 }{float32(v)})
		case 64:
			binary.Write(buf, binary.LittleEndian, struct{ a float64 }{v})
		}
	}

	return buf.Bytes()
}
func (t *tcpPackager) buildWriteBitIOT(tag string, tagIOI []byte, value bool, dataType uint8) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x4e, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)

	var bit int
	bitCount := t.cipTypeMap[dataType].BitCount
	tagSplit := strings.Split(tag, ".")
	if dataType == 211 {
		tag, _, bit = t.TagNameParser(tagSplit[len(tagSplit)-1], 0)
		bit %= 32
	} else {
		bit, _ = strconv.Atoi(tagSplit[len(tagSplit)-1])
	}
	binary.Write(buf, binary.LittleEndian, struct{ a int16 }{int16(bitCount)})

	b := math.Pow(float64(2), float64(bitCount*8)) - 1
	bits := math.Pow(float64(2), float64(bit))
	if !value {
		b -= bits
		bits = 0
	}
	switch bitCount {
	case 2:
		binary.Write(buf, binary.LittleEndian, struct{ a uint16 }{uint16(bits)})
		binary.Write(buf, binary.LittleEndian, struct{ a uint16 }{uint16(b)})
	case 4:
		binary.Write(buf, binary.LittleEndian, struct{ a uint32 }{uint32(bits)})
		binary.Write(buf, binary.LittleEndian, struct{ a uint32 }{uint32(b)})
	case 8:
		binary.Write(buf, binary.LittleEndian, struct{ a uint64 }{uint64(bits)})
		binary.Write(buf, binary.LittleEndian, struct{ a uint64 }{uint64(b)})
	}

	return buf.Bytes()
}
func (t *tcpPackager) BuildWriteIOIRequest(tag string, value interface{}) []byte {
	buf := new(bytes.Buffer)
	var tagData []byte
	_, baseTag, _ := t.TagNameParser(tag, 0)
	dataType := t.knownTags[baseTag]
	if dataType == 211 {
		tagData = t.buildTagIOI(tag, true)
	} else {
		tagData = t.buildTagIOI(tag, false)
	}

	tagSplit := strings.Split(tag, ".")
	if _, e := strconv.ParseInt(tagSplit[len(tagSplit)-1], 10, 8); e == nil {
		if v, ok := value.(bool); ok {
			binary.Write(buf, binary.LittleEndian, t.buildWriteBitIOT(tag, tagData, v, 196))
		}
	} else {
		binary.Write(buf, binary.LittleEndian, t.buildWriteIOT(tagData, value, dataType))
	}

	return t.BuildEIPHeader(buf.Bytes())
}

func (t *tcpPackager) buildTagIOI(tagName string, isBoolArray bool) []byte {
	buf := new(bytes.Buffer)
	tagSplit := strings.Split(tagName, ".")
	for i, ts := range tagSplit {
		if strings.HasSuffix(ts, "]") {
			_, baseTag, index := t.TagNameParser(ts, 0)
			baseTagLenBytes := len(baseTag)

			if isBoolArray && i == len(tagSplit)-1 {
				index = index / 32
			}
			binary.Write(buf, binary.LittleEndian, struct{ H, L uint8 }{0x91, uint8(baseTagLenBytes)})
			binary.Write(buf, binary.LittleEndian, []byte(baseTag))
			if baseTagLenBytes%2 != 0 {
				baseTagLenBytes += 1
				binary.Write(buf, binary.LittleEndian, []byte{0x0})
			}
			if i < len(tagSplit) {
				if index < 256 {
					binary.Write(buf, binary.LittleEndian, struct{ H, L uint8 }{0x28, uint8(index)})
				}
				if index > 255 && index < 65536 {
					binary.Write(buf, binary.LittleEndian, struct{ H, L uint16 }{0x29, uint16(index)})
				}
				if index > 65535 {
					binary.Write(buf, binary.LittleEndian, struct {
						H uint16
						L uint32
					}{0x2A, uint32(index)})
				}
			}
		} else {
			if _, err := strconv.ParseInt(ts, 10, 8); err != nil {
				baseTagLenBytes := len(ts)
				binary.Write(buf, binary.LittleEndian, struct{ H, L uint8 }{0x91, uint8(baseTagLenBytes)})
				binary.Write(buf, binary.LittleEndian, []byte(ts))
				if baseTagLenBytes%2 != 0 {
					baseTagLenBytes += 1
					binary.Write(buf, binary.LittleEndian, []byte{0x0})
				}
			}
		}
	}
	return buf.Bytes()
}
func (t *tcpPackager) buildCIPForwardOpen() []byte {
	rand.Seed(time.Now().UnixNano())
	buf := new(bytes.Buffer)
	t.SerialNumber = uint16(rand.Intn(65000))
	binary.Write(buf, binary.LittleEndian, struct {
		CIPService                       uint8
		CIPPathSize                      uint8
		CIPClassType                     uint8
		CIPClass                         uint8
		CIPInstanceType                  uint8
		CIPInstance                      uint8
		CIPPriority                      uint8
		CIPTimeoutTicks                  uint8
		CIPOTConnectionID                uint32
		CIPTOConnectionTD                uint32
		CIPConnectionSerialNumber        uint16
		CIPVendorID                      uint16
		CIPOriginatorSerialNumber        uint32
		CIPMultiplier                    uint32
		CIPOTRPI                         uint32
		CIPOTNetworkConnectionParameters int16
		CIPTORPI                         uint32
		CIPTONetworkConnectionParameters int16
		CIPTransportTrigger              uint8
	}{
		0x54,
		0x02,
		0x20,
		0x06,
		0x24,
		0x01,
		0x0A,
		0x0E,
		0x20000002,
		0x20000001,
		t.SerialNumber,
		t.VendorID,
		t.OriginatorSerialNumber,
		0x03,
		0x00201234,
		0x43f4,
		0x00204001,
		0x43f4,
		0xA3,
	})
	connectionPath := []byte{0x01, t.ProcessorSlot, 0x20, 0x02, 0x24, 0x01}
	buf.Write(append([]byte{uint8(len(connectionPath) / 2)}, connectionPath...))
	return buf.Bytes()
}
func (t *tcpPackager) buildCIPForwardClose() []byte {
	rand.Seed(time.Now().UnixNano())
	buf := new(bytes.Buffer)
	t.SerialNumber = uint16(rand.Intn(65000))

	binary.Write(buf, binary.LittleEndian, struct {
		CIPService                uint8
		CIPPathSize               uint8
		CIPClassType              uint8
		CIPClass                  uint8
		CIPInstanceType           uint8
		CIPInstance               uint8
		CIPPriority               uint8
		CIPTimeoutTicks           uint8
		CIPConnectionSerialNumber uint16
		CIPVendorID               uint16
		CIPOriginatorSerialNumber uint32
	}{
		0x4E,
		0x02,
		0x20,
		0x06,
		0x24,
		0x01,
		0x0A,
		0x0E,
		t.SerialNumber,
		t.VendorID,
		t.OriginatorSerialNumber,
	})
	connectionPath := []byte{0x01, t.ProcessorSlot, 0x20, 0x02, 0x24, 0x01}
	buf.Write(append([]byte{uint8(len(connectionPath) / 2)}, connectionPath...))
	return buf.Bytes()
}
func (t *tcpPackager) buildEIPSendRRDataHeader(frameLen int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		EIPCommand         uint16
		EIPLength          uint16
		EIPSessionHandle   uint32
		EIPStatus          uint32
		EIPContext         uint64
		EIPOptions         uint32
		EIPInterfaceHandle uint32
		EIPTimeout         uint16
		EIPItemCount       uint16
		EIPItem1Type       uint16
		EIPItem1Length     uint16
		EIPItem2Type       uint16
		EIPItem2Length     uint16
	}{
		0x6F,
		16 + uint16(frameLen),
		t.SessionHandle,
		0x00,
		t.Context,
		0x00,
		0x00,
		0x00,
		0x02,
		0x00,
		0x00,
		0xB2,
		uint16(frameLen),
	})
	return buf.Bytes()
}
func (t *tcpPackager) buildPartialReadIOI(tagIOI []byte, elements int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x52, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)
	binary.Write(buf, binary.LittleEndian, struct{ a, b, c uint16 }{uint16(elements), uint16(t.Offset), uint16(0)})

	return buf.Bytes()
}
func (t *tcpPackager) _getWordCount(start, length, bits int) int {
	total := start + length
	wc := total / bits
	if total%32 > 0 {
		wc += 1
	}
	return wc
}
func (t *tcpPackager) _wordsToBits(tag string, value []byte, count int) int {
	return 0
}

func (t *tcpPackager) _getReplyValues(response []byte, tag string, element int) {
	//tagName, baseTag, index := t.TagNameParser(tag, 0)
	//dataType := t.knownTags[baseTag]
}
func (t *tcpPackager) ParseReply(response []byte, tag string, element int) {
	//tagName, baseTag, index := t.TagNameParser(tag, 0)
	//tagSplit := strings.Split(tag, ".")
	//dataType, _ := t.knownTags[baseTag]
	//bitCount, _ := t.cipTypeMap[dataType]

	//if _, err := strconv.Atoi(tagSplit[len(tagSplit)-1]); err == nil {
	//	//wc := t._getWordCount(pos, element, bitCount)
	//	//t._wordsToBits(tag, wc, element)
	//} else {
	//	if dataType == 211 {
	//
	//	} else {
	//		for k := 0; k < element; k++ {
	//		}
	//	}
	//}

}

func (t *tcpTransporter) Send(request []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastActivity = time.Now()
	t.startCloseTimer()

	var timeout time.Time
	if t.Timeout > 0 {
		timeout = t.lastActivity.Add(t.Timeout)
	}
	if err := t.conn.SetDeadline(timeout); err != nil {
		return nil, err
	}

	//t.logf("logix: sending %02x", request)
	if _, err := t.conn.Write(request); err != nil {
		return nil, err
	}
	data := make([]byte, tcpMaxLength)

	l, e := t.conn.Read(data)
	if e != nil {
		return nil, e
	}
	return data[:l], nil
}
func (t *tcpTransporter) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.close()
}
func (t *tcpTransporter) logf(format string, v ...interface{}) {
	if t.Logger != nil {
		t.Logger.Printf(format, v...)
	}
}
func (t *tcpTransporter) tcpConnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		dialer := net.Dialer{Timeout: t.Timeout}
		conn, err := dialer.Dial("tcp", t.Address)
		if err != nil {
			t.logf("%v", err)
			return err
		}
		t.conn = conn
	}
	return nil
}
func (t *tcpTransporter) Connect() error {
	err := t.tcpConnect()
	if err != nil {
		return err
	}
	//
	//resp, err := t.Send(t.buildRegisterSession())
	//if err != nil {
	//	t.logf("%v", err)
	//	t.close()
	//}
	//binary.Read(bytes.NewBuffer(resp[4:8]), binary.LittleEndian, &t.SessionHandle)
	//
	//resp, err = t.Send(t.buildForwardOpenPacket())
	//if err != nil {
	//	t.logf("%v", err)
	//	t.close()
	//}
	//binary.Read(bytes.NewBuffer(resp[44:48]), binary.LittleEndian, &t.OTNetworkConnectionID)
	return nil
}
func (t *tcpTransporter) startCloseTimer() {
	if t.IdleTimeout <= 0 {
		return
	}

	if t.closeTimer == nil {
		t.closeTimer = time.AfterFunc(t.IdleTimeout, t.closeIdle)
	} else {
		t.closeTimer.Reset(t.IdleTimeout)
	}
}
func (t *tcpTransporter) close() (err error) {
	if t.conn != nil {
		err = t.conn.Close()
		t.conn = nil
	}
	return
}
func (t *tcpTransporter) closeIdle() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(t.lastActivity)
	if idle >= t.IdleTimeout {
		t.logf("logix: closing connection due to idle timeout: %v", idle)
		t.close()
	}
}
