/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-02-13 17:40
**/

package client

import (
	"fmt"
	"net"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
)

type Client struct {
	Name string
	Addr string

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReconnectInterval time.Duration
	HeartBeat         func(conn Conn) error

	ReadBufferSize  int
	WriteBufferSize int
	DailTimeout     time.Duration

	OnOpen         func(conn Conn)
	OnClose        func(conn Conn)
	OnMessage      func(conn Conn, msg []byte)
	OnError        func(stream *socket.Stream[Conn], err error)
	OnSuccess      func()
	OnReconnecting func()
	OnUnknown      func(conn Conn, message []byte, next Middle)

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error

	Protocol protocol.UDPProtocol

	conn                  Conn
	sender                socket.Emitter[Conn]
	router                *router.Router[*socket.Stream[Conn]]
	middle                []func(Middle) Middle
	addr                  *net.UDPAddr
	stopCh                chan struct{}
	isStop                bool
	heartbeatTicker       *time.Ticker
	cancelHeartbeatTicker chan struct{}
}

type Middle router.Middle[*socket.Stream[Conn]]

func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) Use(middle ...func(Middle) Middle) {
	c.middle = append(c.middle, middle...)
}

func (c *Client) Sender() socket.Emitter[Conn] {
	return c.sender
}

func (c *Client) Push(message []byte) error {
	return c.conn.Push(message)
}

func (c *Client) PushUDP(message []byte, addr *net.UDPAddr) error {
	_, err := c.conn.WriteToUDP(message, addr)
	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Conn() Conn {
	return c.conn
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
		panic("addr can not be empty")
	}

	if c.OnOpen == nil {
		c.OnOpen = func(conn Conn) {
			fmt.Println("udp client: connect success")
		}
	}

	if c.OnClose == nil {
		c.OnClose = func(conn Conn) {
			fmt.Println("udp client: connection close")
		}
	}

	if c.OnError == nil {
		c.OnError = func(stream *socket.Stream[Conn], err error) {
			fmt.Println("udp client:", err)
		}
	}

	if c.DailTimeout == 0 {
		c.DailTimeout = 3 * time.Second
	}

	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 512
	}

	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 512
	}

	// if c.HeartBeatInterval == 0 {
	// 	c.HeartBeatInterval = 3 * time.Second
	// }
	//
	// if c.HeartBeatTimeout == 0 {
	// 	c.HeartBeatTimeout = 6 * time.Second
	// }
	//
	// if c.ReconnectInterval == 0 {
	// 	c.ReconnectInterval = time.Second
	// }

	if c.Protocol == nil {
		c.Protocol = &protocol.DefaultUdpProtocol{}
	}

	addr, err := net.ResolveUDPAddr("udp", c.Addr)
	if err != nil {
		panic(err)
	}

	c.addr = addr

	// more useful
	handler, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		fmt.Println(err)
		c.reconnecting()
		return
	}

	var heartBeatTimeout = c.HeartBeatTimeout
	if c.HeartBeatTimeout == 0 {
		heartBeatTimeout = time.Second
	}

	var netConn = &conn{
		addr:     addr,
		conn:     handler,
		client:   c,
		lastPong: time.Now(),
		// PONG
		timeoutTimer:       time.NewTimer(heartBeatTimeout),
		cancelTimeoutTimer: make(chan struct{}),
		UDPProtocol:        c.Protocol,
	}

	c.conn = netConn
	c.sender = socket.NewSender(c.conn)

	// send open message
	err = c.conn.SendOpen()
	if err != nil {
		fmt.Println(err)
		c.reconnecting()
		return
	}

	var tick = time.AfterFunc(c.DailTimeout, func() {
		// handler.ReadFromUDP will read failed
		// then trigger reconnecting
		_ = c.Close()
	})

	var msg = make([]byte, c.Protocol.HeadLen())
	_, _, err = c.conn.Read(msg)
	if err != nil {
		fmt.Println(err)
		c.reconnecting()
		return
	}

	messageType := c.Protocol.GetMessageType(msg)

	if !c.Protocol.IsOpen(messageType) {
		fmt.Println("open message type error: ", messageType)
		c.reconnecting()
		return
	}

	tick.Stop()

	c.stopCh = make(chan struct{})
	c.isStop = false

	var heartBeatInterval = c.HeartBeatInterval
	if c.HeartBeatInterval == 0 {
		heartBeatInterval = time.Second
	}

	c.heartbeatTicker = time.NewTicker(heartBeatInterval)
	c.cancelHeartbeatTicker = make(chan struct{})

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
		c.PongHandler = func(conn Conn) func(data string) error {
			return func(data string) error {
				var t = time.Now()
				c.conn.SetLastPong(t)
				if c.HeartBeatTimeout != 0 {
					return netConn.SetDeadline(t.Add(c.HeartBeatTimeout))
				}
				return nil
			}
		}
	}

	if c.HeartBeatInterval == 0 {
		c.heartbeatTicker.Stop()
	}

	go func() {
		for {
			select {
			case <-c.heartbeatTicker.C:
				if err := c.HeartBeat(c.conn); err != nil {
					fmt.Println(err)
				}
			case <-c.cancelHeartbeatTicker:
				return
			}
		}
	}()

	if c.HeartBeatTimeout == 0 {
		netConn.timeoutTimer.Stop()
	}

	go func() {
		for {
			select {
			case <-netConn.timeoutTimer.C:
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
			case <-netConn.cancelTimeoutTimer:
				return
			}
		}
	}()

	// start success
	if c.OnSuccess != nil {
		c.OnSuccess()
	}

	c.OnOpen(c.conn)

	var reader = c.Protocol.Reader()

	var buffer = make([]byte, c.ReadBufferSize+c.Protocol.HeadLen())

	go func() {
		for {
			n, _, err := c.conn.Read(buffer)
			// close error
			if err != nil {
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
				break
			}

			err = reader(n, buffer, func(bytes []byte) {
				err = c.process(bytes)
			})

			if err != nil {
				if errors.Is(err, errors.ServerClosed) {
					fmt.Println(err)
				}
				if !c.isStop {
					c.stopCh <- struct{}{}
				}
				break
			}
		}
	}()

	<-c.stopCh

	c.isStop = true
	c.heartbeatTicker.Stop()
	c.cancelHeartbeatTicker <- struct{}{}
	netConn.cancelTimeoutTimer <- struct{}{}

	_ = c.Close()
	c.OnClose(c.conn)
	c.reconnecting()
}

func (c *Client) process(message []byte) error {
	messageType := c.Protocol.GetMessageType(message)

	if c.Protocol.IsPing(messageType) || c.Protocol.IsPong(messageType) {
		return c.decodeMessage(message)
	} else if c.Protocol.IsOpen(messageType) {
		return nil
	} else if c.Protocol.IsClose(messageType) {
		return errors.ServerClosed
	} else {
		// bin message
		return c.decodeMessage(message)
	}
}

func (c *Client) decodeMessage(message []byte) error {
	// unpack
	messageType, code, id, route, body := c.Protocol.Decode(message)
	_ = id

	if c.OnMessage != nil {
		c.OnMessage(c.conn, message)
	}

	if c.Protocol.IsUnknown(messageType) {
		if c.OnUnknown != nil {
			c.OnUnknown(c.conn, message, c.middleware)
		}
		return nil
	}

	// Ping
	if c.Protocol.IsPing(messageType) {
		return c.PingHandler(c.conn)("")
	}

	// Pong
	if c.Protocol.IsPong(messageType) {
		return c.PongHandler(c.conn)("")
	}

	// on router

	c.middleware(socket.NewStream(c.conn, code, id, string(route), body))

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
			c.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var n, formatPath = c.router.GetRoute(stream.Event)
	if n == nil {
		if c.OnError != nil {
			c.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var nodeData = n.Data

	stream.Params = socket.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if c.OnError != nil {
				c.OnError(stream, err)
			}
			return
		}
	}

	err := nodeData.Function(stream)
	if err != nil {
		if c.OnError != nil {
			c.OnError(stream, err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](stream); err != nil {
			if c.OnError != nil {
				c.OnError(stream, err)
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
