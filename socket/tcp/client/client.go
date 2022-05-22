package client

import (
	"fmt"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"

	"github.com/lemonyxk/kitty/v2/socket/tcp"
)

type Client struct {
	Name string
	Addr string

	Conn Conn

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReconnectInterval time.Duration
	HeartBeat         func(c Conn) error

	ReadBufferSize  int
	WriteBufferSize int
	DailTimeout     time.Duration

	OnOpen         func(conn Conn)
	OnClose        func(conn Conn)
	OnMessage      func(conn Conn, msg []byte)
	OnError        func(err error)
	OnSuccess      func()
	OnReconnecting func()
	OnUnknown      func(conn Conn, message []byte, next Middle)

	PingHandler func(client Conn) func(appData string) error
	PongHandler func(client Conn) func(appData string) error

	Protocol tcp.Protocol

	router                *router.Router[*socket.Stream[Conn]]
	middle                []func(Middle) Middle
	isStop                bool
	stopCh                chan struct{}
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
	return c.protocol(socket.Bin, []byte(pack.Event), pack.Data)
}

func (c *Client) JsonEmit(pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.protocol(socket.Bin, []byte(pack.Event), data)
}

func (c *Client) ProtoBufEmit(pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.protocol(socket.Bin, []byte(pack.Event), data)
}

func (c *Client) Push(message []byte) error {
	return c.Conn.Write(message)
}

func (c *Client) protocol(messageType byte, route []byte, body []byte) error {
	return c.Conn.protocol(messageType, route, body)
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

func (c *Client) Connect() {

	if c.Addr == "" {
		panic("Addr must set")
	}

	if c.OnOpen == nil {
		c.OnOpen = func(client Conn) {
			fmt.Println("tcp client: connect success")
		}
	}

	if c.OnClose == nil {
		c.OnClose = func(client Conn) {
			fmt.Println("tcp client: connection close")
		}
	}

	if c.OnError == nil {
		c.OnError = func(err error) {
			fmt.Println("tcp client:", err)
		}
	}

	// 握手
	if c.DailTimeout == 0 {
		c.DailTimeout = 2 * time.Second
	}

	// 读出BUF大小
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 1024
	}

	// 读出BUF大小
	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 1024
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
		c.Protocol = &tcp.DefaultProtocol{}
	}

	// 连接服务器
	handler, err := net.DialTimeout("tcp", c.Addr, c.DailTimeout)
	if err != nil {
		fmt.Println(err)
		c.OnError(err)
		c.reconnecting()
		return
	}

	err = handler.(*net.TCPConn).SetReadBuffer(c.ReadBufferSize)
	if err != nil {
		panic(err)
	}

	err = handler.(*net.TCPConn).SetWriteBuffer(c.WriteBufferSize)
	if err != nil {
		panic(err)
	}

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
		c.HeartBeat = func(client Conn) error {
			return client.Ping()
		}
	}

	// no answer
	if c.PingHandler == nil {
		c.PingHandler = func(client Conn) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if c.PongHandler == nil {
		c.PongHandler = func(connection Conn) func(appData string) error {
			return func(appData string) error {
				c.Conn.SetLastPong(time.Now())
				if c.HeartBeatTimeout != 0 {
					c.pongTimer.Reset(c.HeartBeatTimeout)
				}
				return nil
			}
		}
	}

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

	var buffer = make([]byte, c.ReadBufferSize)

	go func() {
		for {
			n, err := c.Conn.Read(buffer)
			// close error
			if err != nil {
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
				break
			}

			err = reader(n, buffer, func(bytes []byte) {
				err = c.decodeMessage(bytes)
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

func (c *Client) decodeMessage(message []byte) error {
	// unpack
	messageType, id, route, body := c.Protocol.Decode(message)
	_ = id

	if c.OnMessage != nil {
		c.OnMessage(c.Conn, message)
	}

	if messageType == socket.Unknown {
		if c.OnUnknown != nil {
			c.OnUnknown(c.Conn, message, c.middleware)
		}
		return nil
	}

	// Ping
	if messageType == socket.Ping {
		return c.PingHandler(c.Conn)("")
	}

	// Pong
	if messageType == socket.Pong {
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
			c.OnError(errors.New(stream.Event + " " + "404 not found"))
		}
		return
	}

	var n, formatPath = c.router.GetRoute(stream.Event)
	if n == nil {
		if c.OnError != nil {
			c.OnError(errors.New(stream.Event + " " + "404 not found"))
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

func (c *Client) SetRouter(router *router.Router[*socket.Stream[Conn]]) *Client {
	c.router = router
	return c
}

func (c *Client) GetRouter() *router.Router[*socket.Stream[Conn]] {
	return c.router
}
