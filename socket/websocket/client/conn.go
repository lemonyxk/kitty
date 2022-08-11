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

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
)

type Conn interface {
	Name() string
	SetName(name string)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
	Write(messageType int, data []byte) (int, error)
	Read() (int, []byte, error)
	Push(msg []byte) error
	LastPong() time.Time
	SetLastPong(t time.Time)
	Client() *Client
	Ping() error
	Pong() error
	Conn() *websocket.Conn
	JsonEmit(event string, data any) error
	ProtoBufEmit(event string, data proto.Message) error
	Emit(event string, data []byte) error
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	name     string
	conn     *websocket.Conn
	client   *Client
	lastPong time.Time
	mux      sync.RWMutex
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
	return c.protocol(protocol.Bin, []byte(event), data)
}

func (c *conn) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return c.protocol(protocol.Bin, []byte(event), msg)
}

func (c *conn) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return c.protocol(protocol.Bin, []byte(event), msg)
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
	err := c.Push(c.client.Protocol.Ping())
	return err
}

func (c *conn) Pong() error {
	err := c.Push(c.client.Protocol.Pong())
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

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	var message = c.client.Protocol.Encode(messageType, 0, route, body)
	err := c.conn.WriteMessage(int(protocol.Bin), message)
	return err
}
