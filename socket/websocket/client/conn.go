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
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
)

type Conn interface {
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
	JsonEmit(pack socket.JsonPack) error
	ProtoBufEmit(pack socket.ProtoBufPack) error
	Emit(pack socket.Pack) error
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	conn     *websocket.Conn
	client   *Client
	lastPong time.Time
	mux      sync.RWMutex
}

func (c *conn) Emit(pack socket.Pack) error {
	return c.protocol(protocol.Bin, []byte(pack.Event), pack.Data)
}

func (c *conn) JsonEmit(pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.protocol(protocol.Bin, []byte(pack.Event), data)
}

func (c *conn) ProtoBufEmit(pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.protocol(protocol.Bin, []byte(pack.Event), data)
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
