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
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
)

type Conn interface {
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	Push(msg []byte) error
	Emit(pack socket.Pack) error
	JsonEmit(msg socket.JsonPack) error
	ProtoBufEmit(msg socket.ProtoBufPack) error
	Close() error
	Write(msg []byte) (int, error)
	FD() int64
	SetFD(fd int64)
	LastPing() time.Time
	SetLastPing(t time.Time)
	Name() string
	Server() *Server
	Conn() net.Conn
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	name     string
	fd       int64
	conn     net.Conn
	server   *Server
	lastPing time.Time
	mux      sync.RWMutex
}

func (c *conn) Name() string {
	return c.name
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
	_, err := c.Write(c.server.Protocol.Ping())
	return err
}

func (c *conn) Pong() error {
	_, err := c.Write(c.server.Protocol.Pong())
	return err
}

func (c *conn) Push(msg []byte) error {
	_, err := c.Write(msg)
	return err
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

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Write(msg []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.Write(msg)
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	var data = c.server.Protocol.Encode(messageType, 0, route, body)
	_, err := c.conn.Write(data)
	return err
}
