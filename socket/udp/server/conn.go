/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-02-12 22:27
**/

package server

import "C"
import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
)

type Conn interface {
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	SendClose() error
	SendOpen() error
	Close() error
	Push(data []byte) error
	JsonEmit(event string, data any) error
	ProtoBufEmit(event string, data proto.Message) error
	Emit(event string, data []byte) error
	Write(msg []byte) (int, error)
	WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error)
	FD() int64
	SetFD(fd int64)
	LastPing() time.Time
	SetLastPing(t time.Time)
	CloseChan() chan struct{}
	AcceptChan() chan []byte
	Name() string
	Server() *Server
	Conn() *net.UDPAddr
	SetReadDeadline(t time.Time) error
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	name         string
	fd           int64
	conn         *net.UDPAddr
	server       *Server
	lastPing     time.Time
	mux          sync.RWMutex
	timeoutTimer *time.Timer
	accept       chan []byte
	close        chan struct{}
}

func (c *conn) SetReadDeadline(t time.Time) error {
	c.timeoutTimer.Reset(t.Sub(time.Now()))
	return nil
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) Server() *Server {
	return c.server
}

func (c *conn) Conn() *net.UDPAddr {
	return c.conn
}

func (c *conn) Host() string {
	return c.conn.String()
}

func (c *conn) CloseChan() chan struct{} {
	return c.close
}

func (c *conn) AcceptChan() chan []byte {
	return c.accept
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

func (c *conn) ClientIP() string {
	if ip, _, err := net.SplitHostPort(c.conn.String()); err == nil {
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

func (c *conn) SendClose() error {
	_, err := c.Write(c.server.Protocol.SendClose())
	return err
}

func (c *conn) SendOpen() error {
	_, err := c.Write(c.server.Protocol.SendOpen())
	return err
}

func (c *conn) Push(msg []byte) error {
	_, err := c.Write(msg)
	return err
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

func (c *conn) Close() error {
	return c.SendClose()
}

func (c *conn) Write(msg []byte) (int, error) {
	if len(msg) > c.server.ReadBufferSize+c.server.Protocol.HeadLen() {
		return 0, errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.server.ReadBufferSize))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.server.netListen.WriteToUDP(msg, c.conn)
}

func (c *conn) WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error) {
	if len(msg) > c.server.ReadBufferSize+c.server.Protocol.HeadLen() {
		return 0, errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.server.ReadBufferSize))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.server.netListen.WriteToUDP(msg, addr)
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	var msg = c.server.Protocol.Encode(messageType, 0, route, body)

	if len(msg) > c.server.ReadBufferSize+c.server.Protocol.HeadLen() {
		return errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.server.ReadBufferSize))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.server.netListen.WriteToUDP(msg, c.conn)
	return err
}
