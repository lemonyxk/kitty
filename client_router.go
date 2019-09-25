package lemo

import (
	"strings"
)

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

func (c *Client) router(client *Client, fte *Fte, message []byte) {

	switch c.TsProto {
	case Json:
		c.jsonRouter(c, fte, message)
	case ProtoBuf:
		c.protoBufRouter(c, fte, message)
	}

}

func (c *Client) jsonRouter(client *Client, fte *Fte, msg []byte) {

	if len(msg) < 12 {
		return
	}

	var event, data = ParseMessage(msg)

	event = strings.Replace(event, "\\", "", -1)

	var f = c.GetRouter(event)

	if f == nil {
		return
	}

	fte.Event = event

	f(c, fte, data)
}

func (c *Client) protoBufRouter(client *Client, fte *Fte, message []byte) {}
