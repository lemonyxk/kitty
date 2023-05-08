package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
	"github.com/lemonyxk/kitty/ssl"
)

type Client struct {
	Name string
	Addr string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

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
	OnError        func(stream *socket.Stream[Conn], err error)
	OnException    func(err error)
	OnSuccess      func()
	OnReconnecting func()
	OnUnknown      func(conn Conn, message []byte, next Middle)

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error

	Protocol protocol.Protocol

	conn                  Conn
	sender                socket.Emitter[Conn]
	router                *router.Router[*socket.Stream[Conn]]
	middle                []func(Middle) Middle
	isStop                bool
	stopCh                chan struct{}
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
			fmt.Println("tcp client: connect success")
		}
	}

	if c.OnClose == nil {
		c.OnClose = func(conn Conn) {
			fmt.Println("tcp client: connection close")
		}
	}

	if c.OnError == nil {
		c.OnError = func(stream *socket.Stream[Conn], err error) {
			fmt.Println("tcp client err:", err)
		}
	}

	if c.OnException == nil {
		c.OnException = func(err error) {
			fmt.Println("tcp client exception:", err)
		}
	}

	if c.DailTimeout == 0 {
		c.DailTimeout = 3 * time.Second
	}

	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 8192
	}

	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 8192
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
		c.Protocol = &protocol.DefaultTcpProtocol{}
	}

	var err error
	var handler net.Conn

	if c.CertFile != "" && c.KeyFile != "" {
		var config, err = ssl.NewTLSConfig(c.CertFile, c.KeyFile)
		if err != nil {
			panic(err)
		}
		handler, err = tls.DialWithDialer(&net.Dialer{Timeout: c.DailTimeout}, "tcp", c.Addr, config)

	} else {
		handler, err = net.DialTimeout("tcp", c.Addr, c.DailTimeout)
	}

	if err != nil {
		c.OnException(err)
		c.reconnecting()
		return
	}

	switch handler.(type) {
	case *tls.Conn:
		err = handler.(*tls.Conn).NetConn().(*net.TCPConn).SetReadBuffer(c.ReadBufferSize)
		if err != nil {
			panic(err)
		}

		err = handler.(*tls.Conn).NetConn().(*net.TCPConn).SetWriteBuffer(c.WriteBufferSize)
		if err != nil {
			panic(err)
		}
	case *net.TCPConn:
		err = handler.(*net.TCPConn).SetReadBuffer(c.ReadBufferSize)
		if err != nil {
			panic(err)
		}

		err = handler.(*net.TCPConn).SetWriteBuffer(c.WriteBufferSize)
		if err != nil {
			panic(err)
		}
	}

	var netConn = &conn{
		conn:     handler,
		client:   c,
		lastPong: time.Now(),
		Protocol: c.Protocol,
	}

	c.conn = netConn
	c.sender = socket.NewSender(c.conn)

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
				conn.SetLastPong(t)
				if c.HeartBeatTimeout != 0 {
					return conn.SetDeadline(t.Add(c.HeartBeatTimeout))
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
					c.OnException(err)
				}
			case <-c.cancelHeartbeatTicker:
				return
			}
		}
	}()

	if c.HeartBeatTimeout != 0 {
		err = c.conn.SetDeadline(time.Now().Add(c.HeartBeatTimeout))
		if err != nil {
			panic(err)
		}
	}

	// start success
	if c.OnSuccess != nil {
		c.OnSuccess()
	}

	c.OnOpen(c.conn)

	var reader = c.Protocol.Reader()

	var buffer = make([]byte, c.ReadBufferSize)

	go func() {
		for {
			n, err := c.conn.Read(buffer)
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
				c.OnException(err)
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

	_ = c.conn.Close()
	c.OnClose(c.conn)
	c.reconnecting()
}

func (c *Client) decodeMessage(message []byte) error {
	// unpack
	order, messageType, code, id, route, body := c.conn.UnPack(message)

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
	c.middleware(socket.NewStream(c.conn, order, messageType, code, id, route, body))

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
			c.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event()))
		}
		return
	}

	var n, formatPath = c.router.GetRoute(stream.Event())
	if n == nil {
		if c.OnError != nil {
			c.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event()))
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

func (c *Client) Conn() Conn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}
