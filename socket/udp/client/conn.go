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
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
)

type Conn interface {
	Name() string
	SetName(name string)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Read(b []byte) (n int, addr *net.UDPAddr, err error)
	Write(b []byte) (int, error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	Close() error
	Conn() *net.UDPConn
	LastPong() time.Time
	SetLastPong(t time.Time)
	Client() *Client
	Ping() error
	Pong() error
	SendClose() error
	SendOpen() error
	SetReadDeadline(t time.Time) error
	socket.Emitter
	protocol.UDPProtocol
}

type conn struct {
	name               string
	conn               *net.UDPConn
	addr               *net.UDPAddr
	client             *Client
	lastPong           time.Time
	timeoutTimer       *time.Timer
	cancelTimeoutTimer chan struct{}
	mux                sync.RWMutex
	messageID          int64
	protocol.UDPProtocol
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
}

func (c *conn) SetReadDeadline(t time.Time) error {
	c.timeoutTimer.Reset(t.Sub(time.Now()))
	return nil
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
	_, err := c.Write(c.PackPing())
	return err
}

func (c *conn) Pong() error {
	_, err := c.Write(c.PackPong())
	return err
}

func (c *conn) SendClose() error {
	_, err := c.Write(c.PackClose())
	return err
}

func (c *conn) SendOpen() error {
	_, err := c.Write(c.PackOpen())
	return err
}

func (c *conn) Write(msg []byte) (int, error) {
	return c.WriteToUDP(msg, c.addr)
}

func (c *conn) WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error) {
	if len(msg) > c.client.ReadBufferSize+c.HeadLen() {
		return 0, errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.client.ReadBufferSize))
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.WriteToUDP(msg, addr)
}

func (c *conn) Pack(messageType byte, messageID int64, route []byte, body []byte) error {
	var msg = c.Encode(messageType, messageID, route, body)
	_, err := c.WriteToUDP(msg, c.addr)
	return err
}
