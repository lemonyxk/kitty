package lemo

var globalClientPath = ""

func (c *Client) InitRouter() {
	c.WebSocketRouter = make(map[string]WebSocketClientFunction)
}

func (c *Client) Group(path string, fn func()) {
	globalClientPath = path
	fn()
	globalClientPath = ""
}

func (c *Client) SetRouter(path string, f WebSocketClientFunction) {
	path = globalClientPath + path
	c.WebSocketRouter[path] = f
}

func (c *Client) GetRouter(path string) WebSocketClientFunction {
	if f, ok := c.WebSocketRouter[path]; ok {
		return f
	}
	return nil
}

func (c *Client) router(client *Client, msg *MessagePackage) {

	switch c.TsProto {
	case Json:
		c.jsonRouter(c, msg)
	case ProtoBuf:
		c.protoBufRouter(c, msg)
	}

}

func (c *Client) jsonRouter(client *Client, msg *MessagePackage) {

	var f = c.GetRouter(msg.Event)

	if f == nil {
		return
	}

	f(c, msg)
}

func (c *Client) protoBufRouter(client *Client, msg *MessagePackage) {}
