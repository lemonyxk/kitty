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
	"time"

	"github.com/fasthttp/websocket"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
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
	Ping() error
	Pong() error
	Conn() *websocket.Conn
	SubProtocols() []string
	SetDeadline(t time.Time) error
	socket.Packer
}

type conn struct {
	name         string
	conn         *websocket.Conn
	lastPong     time.Time
	mux          sync.RWMutex
	subProtocols []string
	protocol.Protocol
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.NetConn().SetDeadline(t)
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

func (c *conn) Pack(order uint32, messageType byte, code uint32, messageID uint64, route []byte, body []byte) error {
	var message = c.Encode(order, messageType, code, messageID, route, body)
	_, err := c.Write(int(protocol.Bin), message)
	return err
}

func (c *conn) UnPack(message []byte) (uint32, byte, uint32, uint64, []byte, []byte) {
	var order, messageType, code, id, route, body = c.Decode(message)
	return order, messageType, code, id, route, body
}
