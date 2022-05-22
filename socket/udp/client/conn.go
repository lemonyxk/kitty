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

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/udp"
)

type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Read(b []byte) (n int, addr *net.UDPAddr, err error)
	Write(b []byte) (err error)
	WriteToAddr(b []byte, addr *net.UDPAddr) (err error)
	Close() error
	Conn() *net.UDPConn
	LastPong() time.Time
	SetLastPong(t time.Time)
	Client() *Client
	Ping() error
	Pong() error
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	client   *Client
	lastPong time.Time
	mux      sync.RWMutex
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

func (c *conn) Ping() error {
	return c.protocol(socket.Ping, nil, nil)
}

func (c *conn) Close() error {
	_ = c.Write(udp.CloseMessage)
	return c.conn.Close()
}

func (c *conn) Read(b []byte) (n int, addr *net.UDPAddr, err error) {
	return c.conn.ReadFromUDP(b)
}

func (c *conn) Pong() error {
	return c.protocol(socket.Pong, nil, nil)
}

func (c *conn) Write(msg []byte) error {
	if len(msg) > c.client.ReadBufferSize+udp.HeadLen {
		return errors.New("max length is " + strconv.Itoa(c.client.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.conn.WriteToUDP(msg, c.addr)
	return err
}

func (c *conn) WriteToAddr(msg []byte, addr *net.UDPAddr) error {
	if len(msg) > c.client.ReadBufferSize+udp.HeadLen {
		return errors.New("max length is " + strconv.Itoa(c.client.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.conn.WriteToUDP(msg, addr)
	return err
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	var msg = c.client.Protocol.Encode(messageType, 0, route, body)
	if len(msg) > c.client.ReadBufferSize+udp.HeadLen {
		return errors.New("max length is " + strconv.Itoa(c.client.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.conn.WriteToUDP(msg, c.addr)
	return err
}
