package ws

func (c *Client) InitRouter() {
	c.WebSocketRouter = make(map[string]WebSocketClientFunction)
}

func (c *Client) SetRouter(route string, f WebSocketClientFunction) {
	c.WebSocketRouter[route] = f
}

func (c *Client) GetRouter(route string) WebSocketClientFunction {
	if f, ok := c.WebSocketRouter[route]; ok {
		return f
	}
	return nil
}

func (c *Client) router(client *Client, fte *Fte, message []byte) {

	switch c.TsProto {
	case Json:
		c.jsonRouter(c, fte, message)
	case ProtoBuf:
		c.protoBufRouter(c, fte, message)
	}

}

func (c *Client) jsonRouter(client *Client, fte *Fte, msg []byte) {

	if len(msg) < 22 {
		return
	}

	var event, data = ParseMessage(msg)

	var f = c.GetRouter(event)

	if f == nil {
		return
	}

	f(c, fte, data)
}

func (c *Client) protoBufRouter(client *Client, fte *Fte, message []byte) {}
