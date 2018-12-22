package go_eip

type Packager interface {
	Verify(request []byte, response []byte) (err error)
}
type Transporter interface {
	Send(request []byte, response []byte) (err error)
}
