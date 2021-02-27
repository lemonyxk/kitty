/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-13 17:40
**/

package client

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/udp"
)

type Client struct {
	Name string
	Host string

	Conn              *net.UDPConn
	AutoHeartBeat     bool
	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HeartBeat         func(c *Client) error
	Reconnect         bool
	ReconnectInterval time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	DailTimeout       time.Duration

	OnOpen    func(client *Client)
	OnClose   func(client *Client)
	OnMessage func(client *Client, msg []byte)
	OnError   func(err error)
	OnSuccess func()
	OnUnknown func(client *Client, message []byte, next Middle)

	PingHandler func(client *Client) func(appData string) error
	PongHandler func(client *Client) func(appData string) error

	Protocol udp.Protocol

	router *Router
	middle []func(Middle) Middle
	mux    sync.RWMutex
	addr   *net.UDPAddr
}

type Middle func(client *Client, stream *socket.Stream)

func (c *Client) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Client) RemoteAddr() net.Addr {
	return c.addr
}

func (c *Client) Use(middle ...func(Middle) Middle) {
	c.middle = append(c.middle, middle...)
}

func (c *Client) Emit(pack socket.Pack) error {
	return c.Push(c.Protocol.Encode(socket.BinData, pack.ID, []byte(pack.Event), pack.Data))
}

func (c *Client) JsonEmit(pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.Push(c.Protocol.Encode(socket.BinData, pack.ID, []byte(pack.Event), data))
}

func (c *Client) ProtoBufEmit(pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.Push(c.Protocol.Encode(socket.BinData, pack.ID, []byte(pack.Event), data))
}

func (c *Client) Push(message []byte) error {
	if len(message) > c.ReadBufferSize+udp.HeadLen {
		return errors.New("max length is " + strconv.Itoa(c.ReadBufferSize) + "but now is " + strconv.Itoa(len(message)))
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	_, err := c.Conn.WriteToUDP(message, c.addr)
	return err
}

func (c *Client) Close() error {
	_, _ = c.Conn.WriteToUDP(udp.CloseMessage, c.addr)
	return c.Conn.Close()
}

func (c *Client) reconnecting() {
	if c.Reconnect == true {
		time.Sleep(c.ReconnectInterval)
		c.Connect()
	}
}

func (c *Client) Connect() {

	if c.Host == "" {
		panic("Host must set")
	}

	if c.OnOpen == nil {
		panic("OnOpen must set")
	}

	if c.OnClose == nil {
		panic("OnClose must set")
	}

	if c.OnError == nil {
		panic("OnError must set")
	}

	if c.DailTimeout == 0 {
		c.DailTimeout = 2 * time.Second
	}

	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 512
	}

	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 512
	}

	// 定时心跳间隔
	if c.HeartBeatInterval == 0 {
		c.HeartBeatInterval = 15 * time.Second
	}

	if c.HeartBeatTimeout == 0 {
		c.HeartBeatTimeout = 30 * time.Second
	}

	// 自动重连间隔
	if c.ReconnectInterval == 0 {
		c.ReconnectInterval = time.Second
	}

	if c.Protocol == nil {
		c.Protocol = &udp.DefaultProtocol{}
	}

	// heartbeat function
	if c.HeartBeat == nil {
		c.HeartBeat = func(client *Client) error {
			return client.Push(client.Protocol.Encode(socket.PingData, 0, nil, nil))
		}
	}

	if c.PingHandler == nil {
		c.PingHandler = func(connection *Client) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if c.PongHandler == nil {
		c.PongHandler = func(connection *Client) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	// 连接服务器
	addr, err := net.ResolveUDPAddr("udp", c.Host)
	if err != nil {
		panic(err)
	}

	c.addr = addr

	// more useful
	handler, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		c.OnError(err)
		c.reconnecting()
		return
	}

	// handler, err := net.DialUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0}, addr)
	// if err != nil {
	// 	c.OnError(err)
	// 	c.reconnecting()
	// 	return
	// }

	c.Conn = handler

	// send open message
	_, err = c.Conn.WriteToUDP(udp.OpenMessage, c.addr)
	if err != nil {
		c.OnError(err)
		c.reconnecting()
		return
	}

	var tick = time.AfterFunc(c.DailTimeout, func() {
		// handler.ReadFromUDP will read failed
		// then trigger reconnecting
		_ = c.Close()
	})

	var msg = make([]byte, c.ReadBufferSize+udp.HeadLen)
	_, _, err = c.Conn.ReadFromUDP(msg)
	if err != nil {
		c.OnError(err)
		c.reconnecting()
		return
	}

	if msg[2] != socket.OpenData {
		c.OnError(err)
		c.reconnecting()
		return
	}

	tick.Stop()

	// start success
	if c.OnSuccess != nil {
		c.OnSuccess()
	}

	// 连接成功
	c.OnOpen(c)

	// 定时器 心跳
	ticker := time.NewTicker(c.HeartBeatInterval)

	// 如果有心跳设置
	if c.AutoHeartBeat != true {
		ticker.Stop()
	}

	go func() {
		for range ticker.C {
			if err := c.HeartBeat(c); err != nil {
				c.OnError(err)
				_ = c.Close()
				break
			}
		}
	}()

	// var reader = c.Protocol.Reader()

	var buffer = make([]byte, c.ReadBufferSize+udp.HeadLen)

	for {
		n, err := c.Conn.Read(buffer)
		// close error
		if err != nil {
			break
		}

		err = c.process(buffer[:n])

		if err != nil {
			if err.Error() != "close" {
				c.OnError(err)
			}
			break
		}
	}

	// 关闭定时器
	ticker.Stop()
	// 关闭连接
	_ = c.Close()
	// 触发回调
	c.OnClose(c)
	// 触发重连设置
	c.reconnecting()
}

func (c *Client) process(message []byte) error {

	switch message[2] {
	case socket.BinData, socket.PingData, socket.PongData:
		return c.decodeMessage(message)
	case socket.OpenData:
		return nil
	case socket.CloseData:
		return errors.New("close")
	default:
		return nil
	}
}

func (c *Client) decodeMessage(message []byte) error {
	// unpack
	messageType, id, route, body := c.Protocol.Decode(message)

	if c.OnMessage != nil {
		c.OnMessage(c, message)
	}

	if messageType == socket.Unknown {
		if c.OnUnknown != nil {
			c.OnUnknown(c, message, c.middleware)
		}
		return nil
	}

	// Ping
	if messageType == socket.PingData {
		return c.PingHandler(c)("")
	}

	// Pong
	if messageType == socket.PongData {
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

	var nodeData = n.Data.(*node)

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
