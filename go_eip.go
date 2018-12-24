package go_eip

type ProtocolDataUnit struct {
	Data []byte
}

type Packager interface {
	SetSessionHandle(uint32)
	SetOffset(uint32)
	IncreaseOffset(uint32)
	SetNetworkConnectionID(uint32)
	GetProgramName() []string
	AddProgramName(string) []string
	GetBitCount(uint8) CIPType
	FillKnownTags(s string, v uint8)

	Verify(request []byte, response []byte) (err error)
	TagNameParser(tag string, offset int) (string, string, int)
	BuildEIPHeader(tagIOI []byte) []byte
	BuildTagListRequest(programName string) []byte
	//BuildPartialReadIOI(tagData []byte, elements int) []byte
	BuildRegisterSessionRequest() []byte
	BuildUnregisterSessionRequest() []byte
	//BuildCIPForwardOpen() []byte
	//BuildCIPForwardClose() []byte
	//BuildEIPSendRRDataHeader(frameLen int) []byte
	BuildForwardOpenRequest() []byte
	BuildForwardCloseRequest() []byte
	//buildTagIOI(tagName string, isBoolArray bool) []byte
	BuildPartialReadRequest(tag, baseTag string) []byte

	//BuildReadIOI(tag string, isBoolArray bool, elements int) []byte
	BuildReadIOIRequest(tag string, isBoolArray bool, elements int) []byte
	BuildWriteIOIRequest(tag string, value interface{}) []byte
}

type Transporter interface {
	//BuildEIPHeader(tagIOI []byte) []byte
	//BuildTagListRequest(programName string) []byte
	//Offset() uint32
	//SetOffset(uint32) uint32
	//GetProgramName() []string
	//AddProgramName(string) []string
	Send(request []byte) (response []byte, err error)
	Connect() (err error)
	Close() error

	//Read(string) (interface{}, error)
	//Write(string, interface{}) error
	//MultiRead()
	//GetPLCTime() (time.Time, error)
	//SetPLCTime(time.Time) error
	//GetTagList() ([]Tag, error)
	//Discover()
}

func NewProtocolDataUnit(data []byte) *ProtocolDataUnit {
	return &ProtocolDataUnit{Data: data}
}
