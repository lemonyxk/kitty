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
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/udp"
)

type Conn interface {
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	Close() error
	Push(data []byte) error
	Emit(pack socket.Pack) error
	JsonEmit(msg socket.JsonPack) error
	ProtoBufEmit(msg socket.ProtoBufPack) error
	Write(msg []byte) (int, error)
	WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error)
	FD() int64
	SetFD(fd int64)
	LastPing() time.Time
	SetLastPing(t time.Time)
	Tick() *time.Timer
	CloseChan() chan struct{}
	AcceptChan() chan []byte
	Name() string
	Server() *Server
	Conn() *net.UDPAddr
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	name     string
	fd       int64
	conn     *net.UDPAddr
	server   *Server
	lastPing time.Time
	mux      sync.RWMutex
	tick     *time.Timer
	accept   chan []byte
	close    chan struct{}
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

func (c *conn) Tick() *time.Timer {
	return c.tick
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
	return c.protocol(socket.Ping, nil, nil)
}

func (c *conn) Pong() error {
	return c.protocol(socket.Pong, nil, nil)
}

func (c *conn) Push(msg []byte) error {
	_, err := c.Write(msg)
	return err
}

func (c *conn) Emit(pack socket.Pack) error {
	return c.protocol(socket.Bin, []byte(pack.Event), pack.Data)
}

func (c *conn) JsonEmit(pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.protocol(socket.Bin, []byte(pack.Event), data)
}

func (c *conn) ProtoBufEmit(pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return c.protocol(socket.Bin, []byte(pack.Event), data)
}

func (c *conn) Close() error {
	_, err := c.Write(udp.CloseMessage)
	return err
}

func (c *conn) Write(msg []byte) (int, error) {
	if len(msg) > c.server.ReadBufferSize+udp.HeadLen {
		return 0, errors.New("max length is " + strconv.Itoa(c.server.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.server.netListen.WriteToUDP(msg, c.conn)
}

func (c *conn) WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error) {
	if len(msg) > c.server.ReadBufferSize+udp.HeadLen {
		return 0, errors.New("max length is " + strconv.Itoa(c.server.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.server.netListen.WriteToUDP(msg, addr)
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	var msg = c.server.Protocol.Encode(messageType, 0, route, body)

	if len(msg) > c.server.ReadBufferSize+udp.HeadLen {
		return errors.New("max length is " + strconv.Itoa(c.server.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.server.netListen.WriteToUDP(msg, c.conn)
	return err
}
