/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-10-29 18:44
**/

package client

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
)

type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Read(b []byte) (n int, addr *net.UDPAddr, err error)
	Write(b []byte) (int, error)
	WriteToAddr(b []byte, addr *net.UDPAddr) (int, error)
	Push(msg []byte) error
	Close() error
	Conn() *net.UDPConn
	LastPong() time.Time
	SetLastPong(t time.Time)
	Client() *Client
	Ping() error
	Pong() error
	SendClose() error
	SendOpen() error
	JsonEmit(pack socket.JsonPack) error
	ProtoBufEmit(pack socket.ProtoBufPack) error
	Emit(pack socket.Pack) error
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
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

func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *conn) Conn() *net.UDPConn {
	return c.conn
}

func (c *conn) LastPong() time.Time {
	return c.lastPong
}

func (c *conn) SetLastPong(t time.Time) {
	c.lastPong = t
}

func (c *conn) Close() error {
	_ = c.SendClose()
	return c.conn.Close()
}

func (c *conn) Push(message []byte) error {
	_, err := c.Write(message)
	return err
}

func (c *conn) Read(b []byte) (n int, addr *net.UDPAddr, err error) {
	return c.conn.ReadFromUDP(b)
}

func (c *conn) Ping() error {
	_, err := c.Write(c.client.Protocol.Ping())
	return err
}

func (c *conn) Pong() error {
	_, err := c.Write(c.client.Protocol.Pong())
	return err
}

func (c *conn) SendClose() error {
	_, err := c.Write(c.client.Protocol.SendClose())
	return err
}

func (c *conn) SendOpen() error {
	_, err := c.Write(c.client.Protocol.SendOpen())
	return err
}

func (c *conn) Write(msg []byte) (int, error) {
	if len(msg) > c.client.ReadBufferSize+c.client.Protocol.HeadLen() {
		return 0, errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.client.ReadBufferSize))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.conn.WriteToUDP(msg, c.addr)
}

func (c *conn) WriteToAddr(msg []byte, addr *net.UDPAddr) (int, error) {
	if len(msg) > c.client.ReadBufferSize+c.client.Protocol.HeadLen() {
		return 0, errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.client.ReadBufferSize))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.conn.WriteToUDP(msg, addr)
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	var msg = c.client.Protocol.Encode(messageType, 0, route, body)
	if len(msg) > c.client.ReadBufferSize+c.client.Protocol.HeadLen() {
		return errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.client.ReadBufferSize))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.conn.WriteToUDP(msg, c.addr)
	return err
}
