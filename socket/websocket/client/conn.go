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

	"github.com/gorilla/websocket"
	"github.com/lemonyxk/kitty/v2/socket"
)

type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
	Write(data []byte) error
	Read() (int, []byte, error)
	LastPong() time.Time
	SetLastPong(t time.Time)
	Client() *Client
	Ping() error
	Pong() error
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	conn     *websocket.Conn
	client   *Client
	lastPong time.Time
	mux      sync.RWMutex
}

func (c *conn) Client() *Client {
	return c.client
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
	return c.protocol(socket.Ping, nil, nil)
}

func (c *conn) Pong() error {
	return c.protocol(socket.Pong, nil, nil)
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Read() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *conn) Write(message []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.WriteMessage(int(socket.Bin), message)
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	var message = c.client.Protocol.Encode(messageType, 0, route, body)
	err := c.conn.WriteMessage(int(socket.Bin), message)
	return err
}
