package go_eip

type ProtocolDataUnit struct {
	Data []byte
}

type Packager interface {
	//getByteCount(uint8) CIPType
	//getStatus([]byte) uint8
	//TagNameParser(tag string, offset int) (string, string, int)
	//BuildEIPHeader(tagIOI []byte) []byte
	//BuildTagListRequest(programName string) []byte
	//BuildRegisterSessionRequest() []byte
	//BuildUnregisterSessionRequest() []byte
	//BuildForwardOpenRequest() []byte
	//BuildForwardCloseRequest() []byte
	//BuildPartialReadRequest(tag string) []byte
	//
	//BuildReadIOIRequest(tag string, isBoolArray bool, elements int) []byte
	//BuildWriteIOIRequest(tag string, value interface{}) []byte
	//BuildMultiReadRequest(tags ...string)[]byte
	//ExtractTagPacket([]byte, string) ([]Tag, error)

	Verify(request []byte, response []byte) (err error)
}

type Transporter interface {
	Send(request []byte) (response []byte, err error)
	Connect() (err error)
	Close() error
}

func NewProtocolDataUnit(data []byte) *ProtocolDataUnit {
	return &ProtocolDataUnit{Data: data}
}
