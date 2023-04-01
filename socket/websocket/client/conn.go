/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-10-29 18:57
**/

package client

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

type Conn interface {
	Name() string
	SetName(name string)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
	Write(messageType int, data []byte) (int, error)
	Read() (int, []byte, error)
	LastPong() time.Time
	SetLastPong(t time.Time)
	Client() *Client
	Ping() error
	Pong() error
	Conn() *websocket.Conn
	SubProtocols() []string
	SetReadDeadline(t time.Time) error
	socket.Emitter
	protocol.Protocol
}

type conn struct {
	name         string
	conn         *websocket.Conn
	client       *Client
	lastPong     time.Time
	mux          sync.RWMutex
	messageID    int64
	subProtocols []string
	protocol.Protocol
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
}

func (c *conn) Conn() *websocket.Conn {
	return c.conn
}

func (c *conn) Emit(event string, data []byte) error {
	return c.Pack(protocol.Bin, atomic.AddInt64(&c.messageID, 1), []byte(event), data)
}

func (c *conn) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return c.Pack(protocol.Bin, atomic.AddInt64(&c.messageID, 1), []byte(event), msg)
}

func (c *conn) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return c.Pack(protocol.Bin, atomic.AddInt64(&c.messageID, 1), []byte(event), msg)
}

func (c *conn) Client() *Client {
	return c.client
}

func (c *conn) SubProtocols() []string {
	return c.subProtocols
}

func (c *conn) LastPong() time.Time {
	return c.lastPong
}

func (c *conn) SetLastPong(t time.Time) {
	c.lastPong = t
}

func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *conn) Ping() error {
	err := c.Push(c.PackPing())
	return err
}

func (c *conn) Pong() error {
	err := c.Push(c.PackPong())
	return err
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Push(message []byte) error {
	_, err := c.Write(int(protocol.Bin), message)
	return err
}

func (c *conn) Read() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *conn) Write(messageType int, message []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return len(message), c.conn.WriteMessage(messageType, message)
}

func (c *conn) Pack(messageType byte, messageID int64, route []byte, body []byte) error {
	var message = c.Encode(messageType, messageID, route, body)
	_, err := c.Write(int(protocol.Bin), message)
	return err
}
