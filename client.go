package go_eip

type ClientHandler interface {
	Packager
	Transporter
}

type client struct {
	packager    Packager
	transporter Transporter
}

func NewClient(handler ClientHandler) Client {
	return &client{packager: handler, transporter: handler}
}

func (c *client) Read() {
	panic("implement me")
}

func (c *client) Write() {
	panic("implement me")
}

func (c *client) MultiRead() {
	panic("implement me")
}

func (c *client) GetPLCTime() {
	panic("implement me")
}

func (c *client) SetPLCTime() {
	panic("implement me")
}

func (c *client) GetTagList() {
	panic("implement me")
}

func (c *client) Discover() {
	panic("implement me")
}

func (c *client) Close() {
	panic("implement me")
}
