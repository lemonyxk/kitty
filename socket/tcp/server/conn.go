/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-02-11 16:22
**/

package server

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
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	Close() error
	Write(msg []byte) (int, error)
	FD() int64
	SetFD(fd int64)
	LastPing() time.Time
	SetLastPing(t time.Time)
	Name() string
	SetName(name string)
	Server() *Server
	Conn() net.Conn
	SetReadDeadline(t time.Time) error
	socket.Emitter
	protocol.Protocol
}

type conn struct {
	name      string
	fd        int64
	conn      net.Conn
	server    *Server
	lastPing  time.Time
	mux       sync.RWMutex
	messageID int64
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

func (c *conn) Server() *Server {
	return c.server
}

func (c *conn) Conn() net.Conn {
	return c.conn
}

func (c *conn) FD() int64 {
	return c.fd
}

func (c *conn) SetFD(fd int64) {
	c.fd = fd
}

func (c *conn) LastPing() time.Time {
	return c.lastPing
}

func (c *conn) SetLastPing(t time.Time) {
	c.lastPing = t
}

func (c *conn) Host() string {
	return c.conn.RemoteAddr().String()
}

func (c *conn) ClientIP() string {
	if ip, _, err := net.SplitHostPort(c.conn.RemoteAddr().String()); err == nil {
		return ip
	}
	return ""
}

func (c *conn) Ping() error {
	_, err := c.Write(c.PackPing())
	return err
}

func (c *conn) Pong() error {
	_, err := c.Write(c.PackPong())
	return err
}

func (c *conn) Push(msg []byte) error {
	_, err := c.Write(msg)
	return err
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

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Write(msg []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.Write(msg)
}

func (c *conn) Pack(messageType byte, messageID int64, route []byte, body []byte) error {
	var data = c.Encode(messageType, messageID, route, body)
	_, err := c.Write(data)
	return err
}
