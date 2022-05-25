package client

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/protocol"

	"github.com/gorilla/websocket"
)

type Client struct {
	Name string
	Addr string

	Conn     Conn
	Response *http.Response

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReconnectInterval time.Duration
	HeartBeat         func(conn Conn) error

	WriteBufferSize int
	ReadBufferSize  int
	DailTimeout     time.Duration

	OnOpen         func(conn Conn)
	OnClose        func(conn Conn)
	OnMessage      func(conn Conn, messageType int, msg []byte)
	OnError        func(err error)
	OnSuccess      func()
	OnReconnecting func()
	OnUnknown      func(conn Conn, message []byte, next Middle)

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error

	Protocol protocol.Protocol

	router                *router.Router[*socket.Stream[Conn]]
	middle                []func(Middle) Middle
	stopCh                chan struct{}
	isStop                bool
	heartbeatTicker       *time.Ticker
	cancelHeartbeatTicker chan struct{}
	pongTimer             *time.Timer
	cancelPongTimer       chan struct{}
}

type Middle router.Middle[*socket.Stream[Conn]]

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
	return c.Conn.Emit(pack)
}

func (c *Client) JsonEmit(pack socket.JsonPack) error {
	return c.Conn.JsonEmit(pack)
}

func (c *Client) ProtoBufEmit(pack socket.ProtoBufPack) error {
	return c.Conn.ProtoBufEmit(pack)
}

func (c *Client) Push(message []byte) error {
	return c.Conn.Push(message)
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
		c.OnOpen = func(conn Conn) {
			fmt.Println("webSocket client: connect success")
		}
	}

	if c.OnClose == nil {
		c.OnClose = func(conn Conn) {
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
		c.DailTimeout = 3 * time.Second
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
		c.Protocol = &protocol.DefaultWsProtocol{}
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

	c.Conn = &conn{conn: handler, client: c, lastPong: time.Now()}

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
		c.HeartBeat = func(conn Conn) error {
			return conn.Ping()
		}
	}

	// no answer
	if c.PingHandler == nil {
		c.PingHandler = func(conn Conn) func(data string) error {
			return func(data string) error {
				return nil
			}
		}
	}

	if c.PongHandler == nil {
		c.PongHandler = func(connection Conn) func(data string) error {
			return func(data string) error {
				c.Conn.SetLastPong(time.Now())
				if c.HeartBeatTimeout != 0 {
					c.pongTimer.Reset(c.HeartBeatTimeout)
				}
				return nil
			}
		}
	}

	// 设置PING处理函数
	handler.SetPingHandler(c.PingHandler(c.Conn))

	// 设置PONG处理函数
	handler.SetPongHandler(c.PongHandler(c.Conn))

	// 如果有心跳设置
	if c.HeartBeatInterval == 0 {
		c.heartbeatTicker.Stop()
	}

	go func() {
		for {
			select {
			case <-c.heartbeatTicker.C:
				if err := c.HeartBeat(c.Conn); err != nil {
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
	c.OnOpen(c.Conn)

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
	c.OnClose(c.Conn)
	// 触发重连设置
	c.reconnecting()

}

func (c *Client) decodeMessage(messageFrame int, message []byte) error {
	// unpack
	messageType, id, route, body := c.Protocol.Decode(message)
	_ = id

	if c.OnMessage != nil {
		c.OnMessage(c.Conn, messageFrame, message)
	}

	if c.Protocol.IsUnknown(messageType) {
		if c.OnUnknown != nil {
			c.OnUnknown(c.Conn, message, c.middleware)
		}
		return nil
	}

	// Ping
	if c.Protocol.IsPing(messageType) {
		return c.PingHandler(c.Conn)("")
	}

	// Pong
	if c.Protocol.IsPong(messageType) {
		return c.PongHandler(c.Conn)("")
	}

	// on router
	c.middleware(&socket.Stream[Conn]{Conn: c.Conn, Pack: socket.Pack{Event: string(route), Data: body}})

	return nil
}

func (c *Client) middleware(stream *socket.Stream[Conn]) {
	var next Middle = c.handler
	for i := len(c.middle) - 1; i >= 0; i-- {
		next = c.middle[i](next)
	}
	next(stream)
}

func (c *Client) handler(stream *socket.Stream[Conn]) {

	if c.router == nil {
		if c.OnError != nil {
			c.OnError(errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var n, formatPath = c.router.GetRoute(stream.Event)
	if n == nil {
		if c.OnError != nil {
			c.OnError(errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var nodeData = n.Data

	stream.Params = kitty.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if c.OnError != nil {
				c.OnError(err)
			}
			return
		}
	}

	err := nodeData.Function(stream)
	if err != nil {
		if c.OnError != nil {
			c.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](stream); err != nil {
			if c.OnError != nil {
				c.OnError(err)
			}
			return
		}
	}
}

func (c *Client) GetDailTimeout() time.Duration {
	return c.DailTimeout
}

func (c *Client) SetRouter(router *router.Router[*socket.Stream[Conn]]) *Client {
	c.router = router
	return c
}

func (c *Client) GetRouter() *router.Router[*socket.Stream[Conn]] {
	return c.router
}
