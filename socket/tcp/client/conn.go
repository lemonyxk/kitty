/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-10-29 18:38
**/

package client

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
)

type Conn interface {
	Name() string
	SetName(name string)
	Conn() net.Conn
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	LastPong() time.Time
	SetLastPong(time.Time)
	Client() *Client
	Ping() error
	Pong() error
	socket.Emitter
	protocol.Protocol
}

type conn struct {
	name      string
	conn      net.Conn
	client    *Client
	lastPong  time.Time
	mux       sync.RWMutex
	messageID int64
	protocol.Protocol
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
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

func (c *conn) Conn() net.Conn {
	return c.conn
}

func (c *conn) Client() *Client {
	return c.client
}

func (c *conn) Ping() error {
	_, err := c.Write(c.PackPing())
	return err
}

func (c *conn) Pong() error {
	_, err := c.Write(c.PackPong())
	return err
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

func (c *conn) Push(message []byte) error {
	_, err := c.Write(message)
	return err
}

func (c *conn) Write(message []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.Write(message)
}

func (c *conn) Pack(messageType byte, messageID int64, route []byte, body []byte) error {
	var message = c.Encode(messageType, messageID, route, body)
	_, err := c.Write(message)
	return err
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}
