package ws

import "github.com/tidwall/gjson"

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

func (c *Client) router(client *Client, messageType int, message []byte) {

	switch c.TsProto {
	case Json:
		c.jsonRouter(c, messageType, message)
	case ProtoBuf:
		c.protoBufRouter(c, messageType, message)
	}

}

func (c *Client) jsonRouter(client *Client, messageType int, message []byte) {
	var event = gjson.GetBytes(message, "event").Str

	var f = c.GetRouter(event)

	if f == nil {
		return
	}

	f(c, messageType, message)
}

func (c *Client) protoBufRouter(client *Client, messageType int, message []byte) {}
