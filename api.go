package go_eip

type Client interface {
	Read()
	Write()
	MultiRead()
	GetPLCTime()
	SetPLCTime()
	GetTagList()
	Discover()
	Close()
}

