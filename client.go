package go_eip

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type CIPType struct {
	ByteCount uint8
	TypeName  string
}

type Option struct {
	VendorID               uint16
	SessionHandle          uint32
	ProcessorSlot          uint8
	ContextPointer         uint8
	SerialNumber           uint16
	OriginatorSerialNumber uint32
	OTNetworkConnectionID  uint32
	SequenceCounter        uint16
	Offset                 uint32
}

var GlobalOption = Option{
	VendorID:               1,
	ProcessorSlot:          0,
	SessionHandle:          0x0000,
	ContextPointer:         0,
	SerialNumber:           0,
	OriginatorSerialNumber: 42,
	SequenceCounter:        1,
	Offset:                 0,
}

var cipTypeMap = map[uint8]CIPType{
	160: {0, "STRING"},
	193: {1, "BOOL"},
	194: {1, "SINT"},
	195: {2, "INT"},
	196: {4, "DINT"},
	197: {8, "LINT"},
	198: {1, "USINT"},
	199: {2, "UINT"},
	200: {4, "UDINT"},
	201: {8, "LWORD"},
	202: {4, "REAL"},
	203: {8, "LREAL"},
	//211: {4, "DWORD"},
}
var contextMap = map[uint8]uint64{
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
var knownTags = make(map[string]uint8)
var ProgramNames = make(map[string]string)

type ClientHandler interface {
	Packager
	Transporter
}

type client struct {
	packager    Packager
	transporter Transporter
}

type Tag struct {
	TagName  string
	Offset   uint16
	DataType uint8
}

func NewClient(handler ClientHandler, slot int) Client {
	c := &client{packager: handler, transporter: handler}
	GlobalOption.ProcessorSlot = uint8(slot)

	resp, err := c.transporter.Send(c.BuildRegisterSessionRequest())
	if err != nil {
		c.transporter.Close()
		return nil
	}
	var sessionHandle, networkCID uint32
	binary.Read(bytes.NewBuffer(resp[4:8]), binary.LittleEndian, &sessionHandle)
	GlobalOption.SessionHandle = sessionHandle

	resp, err = c.transporter.Send(c.BuildForwardOpenRequest())
	if err != nil {
		c.transporter.Close()
		return nil
	}
	binary.Read(bytes.NewBuffer(resp[44:48]), binary.LittleEndian, &networkCID)
	GlobalOption.OTNetworkConnectionID = networkCID

	return c
}

func (c *client) Read(tag string) (interface{}, error) {
	dataType, e := c.getDataType(tag)
	if e != nil {
		return nil, e
	}

	tagSplit := strings.Split(tag, ".")
	var requestData []byte
	if pos, err := strconv.Atoi(tagSplit[len(tagSplit)-1]); err == nil {
		words := (pos + 1) / int(c.getByteCount(dataType).ByteCount*8)
		if (pos + 1) > 32 {
			words += 1
		}
		requestData = c.BuildReadIOIRequest(tag, false, words)
	} else {
		requestData = c.BuildReadIOIRequest(tag, false, 1)
	}

	response, err := c.send(NewProtocolDataUnit(requestData))
	if err != nil {
		return nil, err
	}
	status := c.getStatus(response.Data)
	if status != 0 && status != 6 {
		return nil, errors.New(ErrorText(int(status)))
	}

	return c.ParseOutput(tag, response.Data[50:])
}
func (c *client) Write(tag string, value interface{}) error {
	request := c.BuildWriteIOIRequest(tag, value)
	response, err := c.send(NewProtocolDataUnit(request))
	if err != nil {
		log.Println(err)
		return err
	}
	status := c.getStatus(response.Data)
	if status == 0 {
		return nil
	}

	return errors.New(ErrorText(int(status)))
}
func (c *client) MultiRead(tags ...string) (map[string]interface{}, error) {
	reply := make(map[string]interface{})

	req := c.BuildMultiReadRequest(tags...)
	response, err := c.send(NewProtocolDataUnit(req))
	if err != nil {
		return reply, err
	}
	status := c.getStatus(response.Data)
	if status != 0 {
		return reply, errors.New(ErrorText(int(status)))
	}

	stripped := response.Data[52:]
	for i, tag := range tags {
		var offset uint16
		var extend uint8
		binary.Read(bytes.NewBuffer(stripped[i*2:i*2+2]), binary.LittleEndian, &offset)
		binary.Read(bytes.NewBuffer(stripped[offset:offset+1]), binary.LittleEndian, &status)
		binary.Read(bytes.NewBuffer(stripped[offset+1:offset+2]), binary.LittleEndian, &extend)

		if status != 0 || extend != 0 {
			return reply, errors.New(ErrorText(int(status)))
		}

		v, _ := c.ParseOutput(tag, stripped[offset+2:])
		reply[tag] = v
	}

	return reply, nil
}
func (c *client) GetPLCTime() (time.Time, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		AttributeService      uint8
		AttributeSize         uint8
		AttributeClassType    uint8
		AttributeClass        uint8
		AttributeInstanceType uint8
		AttributeInstance     uint8
		AttributeCount        uint16
		TimeAttribute         uint16
	}{
		0x03,
		0x02,
		0x20,
		0x8B,
		0x24,
		0x01,
		0x01,
		0x0B,
	})

	eipHeader := c.BuildEIPHeader(buf.Bytes())
	response, err := c.send(NewProtocolDataUnit(eipHeader))
	if err != nil {
		return time.Time{}, err
	}
	var plcTimeMSOffset uint64
	binary.Read(bytes.NewBuffer(response.Data[56:64]), binary.LittleEndian, &plcTimeMSOffset)
	originTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	return originTime.Add(time.Microsecond * time.Duration(plcTimeMSOffset)), nil
}
func (c *client) SetPLCTime(t time.Time) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		AttributeService      uint8
		AttributeSize         uint8
		AttributeClassType    uint8
		AttributeClass        uint8
		AttributeInstanceType uint8
		AttributeInstance     uint8
		AttributeCount        uint16
		TimeAttribute         uint16
		Time                  uint64
	}{
		0x04,
		0x02,
		0x20,
		0x8B,
		0x24,
		0x01,
		0x01,
		0x06,
		uint64(time.Now().UnixNano()) / 1e3,
	})
	eipHeader := c.BuildEIPHeader(buf.Bytes())
	_, err := c.send(NewProtocolDataUnit(eipHeader))
	if err != nil {
		return err
	}
	return nil
}
func (c *client) GetTagList() ([]Tag, error) {
	tagList := make([]Tag, 0)

	tList, e := c._getTagList("")
	if e != nil {
		log.Println(e)
		return tagList, e
	}
	tagList = append(tagList, tList...)

	for p := range ProgramNames {
		tList, e := c._getTagList(p)
		if e != nil {
			log.Println(e)
			return tagList, e
		}
		tagList = append(tagList, tList...)
	}

	return tagList, nil
}
func (c *client) Discover() {}
func (c *client) Stop() {
	c.transporter.Send(c.BuildForwardCloseRequest())
	c.transporter.Send(c.BuildUnregisterSessionRequest())
	c.transporter.Close()
}

func (c *client) getByteCount(s uint8) CIPType {
	if cip, ok := cipTypeMap[s]; ok {
		return cip
	}
	return CIPType{}
}
func (c *client) getStatus(data []byte) (status uint8) {
	if len(data) < 48 {
		return 0xFF
	}
	if e := binary.Read(bytes.NewBuffer(data[48:49]), binary.LittleEndian, &status); e != nil {
		return 0xFF
	}
	return
}
func (c *client) TagNameParser(tag string, offset int) (string, string, int) {
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
func (c *client) BuildEIPHeader(tagIOI []byte) []byte {
	buf := new(bytes.Buffer)

	if GlobalOption.ContextPointer == 155 {
		GlobalOption.ContextPointer = 0
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
		GlobalOption.SessionHandle,
		0,
		contextMap[GlobalOption.ContextPointer],
		0,
		0,
		0,
		2,
		0xA1,
		4,
		GlobalOption.OTNetworkConnectionID,
		0xB1,
		uint16(len(tagIOI)) + 2,
		GlobalOption.SequenceCounter,
	})
	GlobalOption.SequenceCounter += 1
	GlobalOption.SequenceCounter = GlobalOption.SequenceCounter % 10000

	buf.Write(tagIOI)

	return buf.Bytes()
}
func (c *client) BuildRegisterSessionRequest() []byte {
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
func (c *client) BuildUnregisterSessionRequest() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct {
		EIPCommand       uint16
		EIPLength        uint16
		EIPSessionHandle uint32
		EIPStatus        uint32
		EIPContext       uint64
		EIPOptions       uint32
	}{
		0x66, 0x00, GlobalOption.SessionHandle,
		0x00, contextMap[GlobalOption.ContextPointer], 0x00,
	})
	return buf.Bytes()
}
func (c *client) BuildForwardOpenRequest() []byte {
	rand.Seed(time.Now().UnixNano())
	forwardOpenBuf := new(bytes.Buffer)
	GlobalOption.SerialNumber = uint16(rand.Intn(65000))
	binary.Write(forwardOpenBuf, binary.LittleEndian, struct {
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
		GlobalOption.SerialNumber,
		GlobalOption.VendorID,
		GlobalOption.OriginatorSerialNumber,
		0x03,
		0x00201234,
		0x43f4,
		0x00204001,
		0x43f4,
		0xA3,
	})
	connectionPath := []byte{0x01, GlobalOption.ProcessorSlot, 0x20, 0x02, 0x24, 0x01}
	forwardOpenBuf.Write(append([]byte{uint8(len(connectionPath) / 2)}, connectionPath...))

	dH := c.buildEIPSendRRDataHeader(len(forwardOpenBuf.Bytes()))
	buf := new(bytes.Buffer)
	buf.Write(append(dH, forwardOpenBuf.Bytes()...))
	return buf.Bytes()
}
func (c *client) BuildForwardCloseRequest() []byte {
	rand.Seed(time.Now().UnixNano())
	forwardCloseBuf := new(bytes.Buffer)
	GlobalOption.SerialNumber = uint16(rand.Intn(65000))
	binary.Write(forwardCloseBuf, binary.LittleEndian, struct {
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
		GlobalOption.SerialNumber,
		GlobalOption.VendorID,
		GlobalOption.OriginatorSerialNumber,
	})
	connectionPath := []byte{0x01, GlobalOption.ProcessorSlot, 0x20, 0x02, 0x24, 0x01}
	forwardCloseBuf.Write(append([]byte{uint8(len(connectionPath) / 2)}, connectionPath...))

	dH := c.buildEIPSendRRDataHeader(len(forwardCloseBuf.Bytes()))
	buf := new(bytes.Buffer)
	buf.Write(append(dH, forwardCloseBuf.Bytes()...))
	return buf.Bytes()
}
func (c *client) BuildTagListRequest(programName string) []byte {
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

	if GlobalOption.Offset < 256 {
		binary.Write(pathSegment, binary.LittleEndian, struct{ H, L uint8 }{
			0x24, uint8(GlobalOption.Offset),
		})
	} else {
		binary.Write(pathSegment, binary.LittleEndian, struct{ H, L uint16 }{
			0x25, uint16(GlobalOption.Offset),
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

	return c.BuildEIPHeader(buf.Bytes())
}
func (c *client) BuildPartialReadRequest(tag string) []byte {
	tagIOI := c.buildTagIOI(tag, false)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x52, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)
	binary.Write(buf, binary.LittleEndian, struct{ a, b, c uint16 }{uint16(1), uint16(GlobalOption.Offset), uint16(0)})

	return c.BuildEIPHeader(buf.Bytes())
}
func (c *client) BuildReadIOIRequest(tag string, isBoolArray bool, elements int) []byte {
	tagIOI := c.buildTagIOI(tag, isBoolArray)
	return c.BuildEIPHeader(c.buildReadIOI(tagIOI, elements))
}
func (c *client) BuildMultiReadRequest(tags ...string) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, []uint8{0x0a, 0x02, 0x20, 0x02, 0x24, 0x01})
	offset := len(buf.Bytes())
	if len(tags) > 2 {
		offset += (len(tags) - 2) * 2
	}

	binary.Write(buf, binary.LittleEndian, uint16(len(tags)))
	binary.Write(buf, binary.LittleEndian, uint16(offset))

	segments := make([][]byte, 0)
	for _, tag := range tags {
		tI := c.buildReadIOI(c.buildTagIOI(tag, false), 1)
		segments = append(segments, tI)
	}

	for i := range tags[:len(tags)-1] {
		offset += len(segments[i])
		binary.Write(buf, binary.LittleEndian, uint16(offset))
	}

	for _, ti := range segments {
		binary.Write(buf, binary.LittleEndian, ti)
	}

	return c.BuildEIPHeader(buf.Bytes())
}
func (c *client) BuildWriteIOIRequest(tag string, value interface{}) []byte {
	buf := new(bytes.Buffer)
	tagData := c.buildTagIOI(tag, false)
	tagSplit := strings.Split(tag, ".")
	if _, e := strconv.ParseInt(tagSplit[len(tagSplit)-1], 10, 8); e == nil {
		if v, ok := value.(bool); ok {
			binary.Write(buf, binary.LittleEndian, c.buildWriteBitIOT(tag, tagData, v, knownTags[tag]))
		}
	} else {
		binary.Write(buf, binary.LittleEndian, c.buildWriteIOT(tagData, value, knownTags[tag]))
	}

	return c.BuildEIPHeader(buf.Bytes())
}
func (c *client) buildWriteIOT(tagIOI []byte, value interface{}, dataType uint8) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x4d, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)

	if dataType == 160 {
		binary.Write(buf, binary.LittleEndian, struct {
			a, b uint8
			c, d uint16
		}{dataType, 0x02, 0x0FCE, 1})
	} else {
		binary.Write(buf, binary.LittleEndian, struct {
			a, b uint8
			c    uint16
		}{dataType, 0x0, 1})
	}

	switch dataType {
	case 160:
		binary.Write(buf, binary.LittleEndian, struct{ uint32 }{uint32(len(value.(string)))})
		binary.Write(buf, binary.LittleEndian, []byte(value.(string)))
		if len(value.(string)) < 84 {
			binary.Write(buf, binary.LittleEndian, make([]byte, 84-len(value.(string))))
		}

	case 193:
		v, e := strconv.ParseBool(fmt.Sprintf("%v", value))
		if e != nil {
			log.Println(e)
		}
		binary.Write(buf, binary.LittleEndian, struct{ a bool }{v})
	case 194, 195, 196, 197, 198, 199, 200, 201, 211:
		bitCount := cipTypeMap[dataType].ByteCount * 8
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
		bitCount := cipTypeMap[dataType].ByteCount * 8
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
func (c *client) buildWriteBitIOT(tag string, tagIOI []byte, value bool, dataType uint8) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x4e, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)

	var bit int
	bitCount := cipTypeMap[dataType].ByteCount
	tagSplit := strings.Split(tag, ".")
	if dataType == 211 {
		tag, _, bit = c.TagNameParser(tagSplit[len(tagSplit)-1], 0)
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
func (c *client) buildTagIOI(tagName string, isBoolArray bool) []byte {
	buf := new(bytes.Buffer)
	tagSplit := strings.Split(tagName, ".")
	for i, ts := range tagSplit {
		if strings.HasSuffix(ts, "]") {
			_, baseTag, index := c.TagNameParser(ts, 0)
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
func (c *client) buildReadIOI(tagIOI []byte, elements int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, struct{ a, b uint8 }{0x4c, uint8(len(tagIOI) / 2)})
	binary.Write(buf, binary.LittleEndian, tagIOI)
	binary.Write(buf, binary.LittleEndian, struct{ a uint16 }{uint16(elements)})
	return buf.Bytes()
}
func (c *client) buildEIPSendRRDataHeader(frameLen int) []byte {
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
		GlobalOption.SessionHandle,
		0x00,
		contextMap[GlobalOption.ContextPointer],
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

func (c *client) parseTag(packet []byte, programName string) (Tag, error) {
	tag := Tag{}
	var length uint16
	if e := binary.Read(bytes.NewBuffer(packet[8:10]), binary.LittleEndian, &length); e != nil {
		return tag, e
	}

	tag.TagName = string(packet[10 : length+10])
	if programName != "" {
		tag.TagName = programName + "." + tag.TagName
	}
	if e := binary.Read(bytes.NewBuffer(packet[:2]), binary.LittleEndian, &tag.Offset); e != nil {
		return tag, e
	}
	if e := binary.Read(bytes.NewBuffer(packet[4:5]), binary.LittleEndian, &tag.DataType); e != nil {
		log.Println(e)
	}

	return tag, nil
}
func (c *client) ExtractTagPacket(data []byte, programName string) ([]Tag, error) {
	var packetStart uint16 = 50
	var tagLen uint16
	tagList := make([]Tag, 0)

	for int(packetStart) < len(data) {
		if e := binary.Read(bytes.NewBuffer(data[packetStart+8:packetStart+10]), binary.LittleEndian, &tagLen); e != nil {
			return tagList, e
		}
		packet := data[packetStart : packetStart+tagLen+10]
		var offset uint16
		if e := binary.Read(bytes.NewBuffer(packet[:2]), binary.LittleEndian, &offset); e != nil {
			return tagList, e
		}
		GlobalOption.Offset = uint32(offset)
		tag, _ := c.parseTag(packet, programName)
		tagList = append(tagList, tag)
		if programName == "" {
			if strings.Contains(tag.TagName, "Program:") {
				ProgramNames[tag.TagName] = tag.TagName
			}
		}
		packetStart += tagLen + 10
	}
	for _, t := range tagList {
		knownTags[t.TagName] = t.DataType
	}
	return tagList, nil
}

func (c *client) ParseOutput(tag string, data []byte) (interface{}, error) {
	var dataType uint8
	binary.Read(bytes.NewBuffer(data[:1]), binary.LittleEndian, &dataType)

	tagSplit := strings.Split(tag, ".")

	switch dataType {
	case 193:
		var v bool
		binary.Read(bytes.NewBuffer(data[2:2+1]), binary.LittleEndian, &v)
		return v, nil
	case 194, 195, 196, 197, 198, 199, 200, 201, 211:
		byteCount := c.getByteCount(dataType).ByteCount
		getBool := false
		pos, err := strconv.Atoi(tagSplit[len(tagSplit)-1])
		if err == nil {
			getBool = true
		}

		switch byteCount {
		case 1:
			var v uint8
			binary.Read(bytes.NewBuffer(data[2:2+byteCount]), binary.LittleEndian, &v)
			if getBool && pos < 8 {
				return (v>>uint(pos))&1 == 1, nil
			}
			return v, nil
		case 2:
			var v uint16
			binary.Read(bytes.NewBuffer(data[2:2+byteCount]), binary.LittleEndian, &v)
			if getBool && pos < 16 {
				return (v>>uint(pos))&1 == 1, nil
			}
			return v, nil
		case 4:
			var v uint32
			binary.Read(bytes.NewBuffer(data[2:2+byteCount]), binary.LittleEndian, &v)
			if getBool && pos < 32 {
				return (v>>uint(pos))&1 == 1, nil
			}
			return v, nil
		case 8:
			var v uint64
			binary.Read(bytes.NewBuffer(data[2:2+byteCount]), binary.LittleEndian, &v)
			if getBool && pos < 64 {
				return (v>>uint(pos))&1 == 1, nil
			}
			return v, nil
		}
	case 202:
		var v float32
		binary.Read(bytes.NewBuffer(data[2:2+4*8]), binary.LittleEndian, &v)
		return v, nil
	case 203:
		var v float64
		binary.Read(bytes.NewBuffer(data[2:2+8*8]), binary.LittleEndian, &v)
		return v, nil
	case 160:
		var strLen uint32
		binary.Read(bytes.NewBuffer(data[4:8]), binary.LittleEndian, &strLen)
		return string(data[8 : 8+uint8(strLen)]), nil
	default:
		return nil, errors.New("unknown dataType")
	}
	return nil, errors.New("unknown error")
}

func (c *client) getDataType(tag string) (uint8, error) {
	var dataType uint8
	if _, ok := knownTags[tag]; !ok {
		r := c.BuildPartialReadRequest(tag)
		response, err := c.send(NewProtocolDataUnit(r))
		if err != nil {
			return dataType, err
		}
		if s := c.getStatus(response.Data); s != 0 {
			return dataType, errors.New(ErrorText(int(s)))
		}
		if e := binary.Read(bytes.NewBuffer(response.Data[50:51]), binary.LittleEndian, &dataType); e != nil {
			return 0, e
		}
		knownTags[tag] = dataType
	}
	return dataType, nil
}

func (c *client) _getTagList(p string) ([]Tag, error) {
	tagList := make([]Tag, 0)

	tagListRequest := c.BuildTagListRequest(p)
	response, err := c.send(NewProtocolDataUnit(tagListRequest))
	if err != nil {
		return tagList, err
	}
	tList, e := c.ExtractTagPacket(response.Data, p)
	if e != nil {
		return tagList, e
	}
	tagList = append(tagList, tList...)

	status := c.getStatus(response.Data)
	for status == 6 {
		tagListRequest := c.BuildTagListRequest(p)
		response, err := c.send(NewProtocolDataUnit(tagListRequest))
		if err != nil {
			return tagList, err
		}
		status = c.getStatus(response.Data)
		tList, e := c.ExtractTagPacket(response.Data, p)
		if e != nil {
			return tagList, e
		}
		tagList = append(tagList, tList...)
		time.Sleep(250 * time.Millisecond)
	}

	return tagList, nil
}

func (c *client) send(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	dataResponse, err := c.transporter.Send(request.Data)

	if err = c.packager.Verify(request.Data, dataResponse); err != nil {
		return
	}
	if dataResponse == nil || len(dataResponse) == 0 {
		err = fmt.Errorf("eip: response data is empty")
		return
	}
	response = &ProtocolDataUnit{Data: dataResponse}
	err = responseError(response)
	return response, err
}

func responseError(response *ProtocolDataUnit) error {
	return nil
}
