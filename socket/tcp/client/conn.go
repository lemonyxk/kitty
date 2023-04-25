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
	"time"

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
	SetDeadline(t time.Time) error
	socket.Packer
	protocol.Protocol
}

type conn struct {
	name     string
	conn     net.Conn
	client   *Client
	lastPong time.Time
	mux      sync.RWMutex
	protocol.Protocol
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
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

func (c *conn) Write(message []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.Write(message)
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *conn) Pack(async byte, messageType byte, code uint32, messageID uint64, route []byte, body []byte) error {
	var message = c.Encode(async, messageType, code, messageID, route, body)
	_, err := c.Write(message)
	return err
}

func (c *conn) Push(message []byte) error {
	_, err := c.Write(message)
	return err
}
