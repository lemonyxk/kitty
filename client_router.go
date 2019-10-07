package lemo

var globalClientPath = ""

func (c *Client) InitRouter() {
	c.Router = make(map[string]WebSocketClientFunction)
}

func (c *Client) Group(path string, fn func()) {
	globalClientPath = path
	fn()
	globalClientPath = ""
}

func (c *Client) SetRouter(path string, f WebSocketClientFunction) {
	path = globalClientPath + path
	c.Router[path] = f
}

func (c *Client) GetRouter(path string) WebSocketClientFunction {
	if f, ok := c.Router[path]; ok {
		return f
	}
	return nil
}

func (c *Client) router(client *Client, receive *ReceivePackage) {

	var f = c.GetRouter(receive.Event)

	if f == nil {
		return
	}

	err := f(c, receive)
	if err != nil {
		c.OnError(err)
	}

}
