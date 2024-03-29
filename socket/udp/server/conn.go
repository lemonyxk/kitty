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

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
)

type Conn interface {
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	SendClose() error
	SendOpen() error
	Close() error
	Write(msg []byte) (int, error)
	WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error)
	FD() int64
	SetFD(fd int64)
	LastPing() time.Time
	SetLastPing(t time.Time)
	CloseChan() chan struct{}
	AcceptChan() chan []byte
	Name() string
	SetName(name string)
	Conn() *net.UDPAddr
	SetDeadline(t time.Time) error
	socket.Packer
}

type conn struct {
	name         string
	fd           int64
	conn         *net.UDPAddr
	lastPing     time.Time
	mux          sync.RWMutex
	timeoutTimer *time.Timer
	accept       chan []byte
	close        chan struct{}
	mtu          int
	netListen    *net.UDPConn
	protocol.UDPProtocol
}

func (c *conn) SetDeadline(t time.Time) error {
	c.timeoutTimer.Reset(t.Sub(time.Now()))
	return nil
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
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

func (c *conn) Close() error {
	return c.SendClose()
}

func (c *conn) Write(msg []byte) (int, error) {
	return c.WriteToUDP(msg, c.conn)
}

func (c *conn) WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error) {
	if len(msg) > c.mtu+c.HeadLen() {
		return 0, errors.Wrap(errors.MaximumExceeded, strconv.Itoa(c.mtu))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.netListen.WriteToUDP(msg, addr)
}

func (c *conn) Push(msg []byte) error {
	_, err := c.Write(msg)
	return err
}

func (c *conn) Pack(order uint32, messageType byte, code uint32, messageID uint64, route []byte, body []byte) error {
	var msg = c.Encode(order, messageType, code, messageID, route, body)
	_, err := c.WriteToUDP(msg, c.conn)
	return err
}

func (c *conn) UnPack(message []byte) (uint32, byte, uint32, uint64, []byte, []byte) {
	var order, messageType, code, id, route, body = c.Decode(message)
	return order, messageType, code, id, route, body
}
