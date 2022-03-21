package client

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/kitty"

	"github.com/lemonyxk/kitty/v2/socket"

	websocket2 "github.com/lemonyxk/kitty/v2/socket/websocket"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Client struct {
	Name string
	Addr string

	Conn     *Conn
	Response *http.Response

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReconnectInterval time.Duration
	HeartBeat         func(c *Client) error

	WriteBufferSize int
	ReadBufferSize  int
	DailTimeout     time.Duration

	OnOpen         func(client *Client)
	OnClose        func(client *Client)
	OnMessage      func(client *Client, messageType int, msg []byte)
	OnError        func(err error)
	OnSuccess      func()
	OnReconnecting func()
	OnUnknown      func(conn *Client, message []byte, next Middle)

	PingHandler func(client *Client) func(appData string) error
	PongHandler func(client *Client) func(appData string) error

	Protocol websocket2.Protocol

	router                *Router
	middle                []func(Middle) Middle
	stopCh                chan struct{}
	isStop                bool
	heartbeatTicker       *time.Ticker
	cancelHeartbeatTicker chan struct{}
	pongTimer             *time.Timer
	cancelPongTimer       chan struct{}
}

type Middle func(client *Client, stream *socket.Stream)

func (c *Client) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Client) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Client) Use(middle ...func(Middle) Middle) {
	c.middle = append(c.middle, middle...)
}

func (c *Client) Emit(pack socket.Pack) error {
	return c.Push(c.Protocol.Encode(socket.Bin, pack.ID, []byte(pack.Event), pack.Data))
}

func (c *Client) JsonEmit(pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.Push(c.Protocol.Encode(socket.Bin, pack.ID, []byte(pack.Event), data))
}

func (c *Client) ProtoBufEmit(pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.Push(c.Protocol.Encode(socket.Bin, pack.ID, []byte(pack.Event), data))
}

func (c *Client) Ping() error {
	return c.Push(c.Protocol.Encode(socket.Ping, 0, nil, nil))
}

func (c *Client) Pong() error {
	return c.Push(c.Protocol.Encode(socket.Pong, 0, nil, nil))
}

func (c *Client) Push(message []byte) error {
	return c.Conn.Write(message)
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func (c *Client) reconnecting() {
	if c.ReconnectInterval != 0 {
		time.Sleep(c.ReconnectInterval)
		if c.OnReconnecting != nil {
			c.OnReconnecting()
		}
		c.Connect()
	}
}

// Connect 连接服务器
func (c *Client) Connect() {

	if c.Addr == "" {
		panic("Addr must set")
	}

	if c.OnOpen == nil {
		c.OnOpen = func(client *Client) {
			fmt.Println("webSocket client: connect success")
		}
	}

	if c.OnClose == nil {
		c.OnClose = func(client *Client) {
			fmt.Println("webSocket client: connection close")
		}
	}

	if c.OnError == nil {
		c.OnError = func(err error) {
			fmt.Println("webSocket client:", err)
		}
	}

	// 握手
	if c.DailTimeout == 0 {
		c.DailTimeout = 2 * time.Second
	}

	// 写入BUF大小
	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 1024
	}

	// 读出BUF大小
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 1024
	}

	// 定时心跳间隔
	if c.HeartBeatInterval == 0 {
		c.HeartBeatInterval = 3 * time.Second
	}

	// 服务器返回PONG超时
	if c.HeartBeatTimeout == 0 {
		c.HeartBeatTimeout = 6 * time.Second
	}

	// 自动重连间隔
	if c.ReconnectInterval == 0 {
		c.ReconnectInterval = time.Second
	}

	if c.Protocol == nil {
		c.Protocol = &websocket2.DefaultProtocol{}
	}

	var dialer = websocket.Dialer{
		HandshakeTimeout: c.DailTimeout,
		WriteBufferSize:  c.WriteBufferSize,
		ReadBufferSize:   c.ReadBufferSize,
	}

	// 连接服务器
	handler, response, err := dialer.Dial(c.Addr, nil)
	if err != nil {
		fmt.Println(err)
		c.OnError(err)
		c.reconnecting()
		return
	}

	c.Response = response

	c.Conn = &Conn{Conn: handler, LastPong: time.Now()}

	c.stopCh = make(chan struct{})
	c.isStop = false

	// 定时器 心跳
	c.heartbeatTicker = time.NewTicker(c.HeartBeatInterval)
	c.cancelHeartbeatTicker = make(chan struct{})

	// PONG
	c.pongTimer = time.NewTimer(c.HeartBeatTimeout)
	c.cancelPongTimer = make(chan struct{})

	// heartbeat function
	if c.HeartBeat == nil {
		c.HeartBeat = func(client *Client) error {
			return client.Ping()
		}
	}

	// no answer
	if c.PingHandler == nil {
		c.PingHandler = func(client *Client) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if c.PongHandler == nil {
		c.PongHandler = func(connection *Client) func(appData string) error {
			return func(appData string) error {
				c.Conn.LastPong = time.Now()
				if c.HeartBeatTimeout != 0 {
					c.pongTimer.Reset(c.HeartBeatTimeout)
				}
				return nil
			}
		}
	}

	// 设置PING处理函数
	handler.SetPingHandler(c.PingHandler(c))

	// 设置PONG处理函数
	handler.SetPongHandler(c.PongHandler(c))

	// 如果有心跳设置
	if c.HeartBeatInterval == 0 {
		c.heartbeatTicker.Stop()
	}

	go func() {
		for {
			select {
			case <-c.heartbeatTicker.C:
				if err := c.HeartBeat(c); err != nil {
					c.OnError(err)
				}
			case <-c.cancelHeartbeatTicker:
				return
			}
		}
	}()

	if c.HeartBeatTimeout == 0 {
		c.pongTimer.Stop()
	}

	go func() {
		for {
			select {
			case <-c.pongTimer.C:
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
			case <-c.cancelPongTimer:
				return
			}
		}
	}()

	// start success
	if c.OnSuccess != nil {
		c.OnSuccess()
	}

	// 连接成功
	c.OnOpen(c)

	var reader = c.Protocol.Reader()

	go func() {
		for {
			messageFrame, message, err := c.Conn.Read()
			// close error
			if err != nil {
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
				break
			}

			err = reader(len(message), message, func(bytes []byte) {
				err = c.decodeMessage(messageFrame, bytes)
			})

			if err != nil {
				c.OnError(err)
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
				break
			}
		}
	}()

	<-c.stopCh

	c.isStop = true
	c.cancelHeartbeatTicker <- struct{}{}
	c.cancelPongTimer <- struct{}{}

	// 关闭定时器
	c.heartbeatTicker.Stop()
	// 关闭连接
	_ = c.Close()
	// 触发回调
	c.OnClose(c)
	// 触发重连设置
	c.reconnecting()

}

func (c *Client) decodeMessage(messageFrame int, message []byte) error {
	// unpack
	messageType, id, route, body := c.Protocol.Decode(message)

	if c.OnMessage != nil {
		c.OnMessage(c, messageFrame, message)
	}

	if messageType == socket.Unknown {
		if c.OnUnknown != nil {
			c.OnUnknown(c, message, c.middleware)
		}
		return nil
	}

	// Ping
	if messageType == socket.Ping {
		return c.PingHandler(c)("")
	}

	// Pong
	if messageType == socket.Pong {
		return c.PongHandler(c)("")
	}

	// on router
	c.middleware(c, &socket.Stream{Pack: socket.Pack{Event: string(route), Data: body, ID: id}})

	return nil
}

func (c *Client) middleware(conn *Client, stream *socket.Stream) {
	var next Middle = c.handler
	for i := len(c.middle) - 1; i >= 0; i-- {
		next = c.middle[i](next)
	}
	next(conn, stream)
}

func (c *Client) handler(conn *Client, stream *socket.Stream) {

	if c.router == nil {
		if c.OnError != nil {
			c.OnError(errors.New(stream.Event + " " + "404 not found"))
		}
		return
	}

	var n, formatPath = c.router.getRoute(stream.Event)
	if n == nil {
		if c.OnError != nil {
			c.OnError(errors.New(stream.Event + " " + "404 not found"))
		}
		return
	}

	var nodeData = n.Data

	stream.Params = kitty.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](conn, stream); err != nil {
			if c.OnError != nil {
				c.OnError(err)
			}
			return
		}
	}

	err := nodeData.Function(conn, stream)
	if err != nil {
		if c.OnError != nil {
			c.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](conn, stream); err != nil {
			if c.OnError != nil {
				c.OnError(err)
			}
			return
		}
	}
}

func (c *Client) SetRouter(router *Router) *Client {
	c.router = router
	return c
}

func (c *Client) GetRouter() *Router {
	return c.router
}
