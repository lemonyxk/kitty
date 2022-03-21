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

	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/udp"
)

type Conn struct {
	Name     string
	FD       int64
	Conn     *net.UDPAddr
	Server   *Server
	LastPing time.Time
	mux      sync.RWMutex
	tick     *time.Timer
	accept   chan []byte
	close    chan struct{}
}

func (c *Conn) Host() string {
	return c.Conn.String()
}

func (c *Conn) ClientIP() string {
	if ip, _, err := net.SplitHostPort(c.Conn.String()); err == nil {
		return ip
	}
	return ""
}

func (c *Conn) Ping() error {
	return c.Push(c.Server.Protocol.Encode(socket.Ping, 0, nil, nil))
}

func (c *Conn) Pong() error {
	return c.Push(c.Server.Protocol.Encode(socket.Pong, 0, nil, nil))
}

func (c *Conn) Push(msg []byte) error {
	return c.Server.Push(c.FD, msg)
}

func (c *Conn) Emit(pack socket.Pack) error {
	return c.Server.Emit(c.FD, pack)
}

func (c *Conn) JsonEmit(msg socket.JsonPack) error {
	return c.Server.JsonEmit(c.FD, msg)
}

func (c *Conn) ProtoBufEmit(msg socket.ProtoBufPack) error {
	return c.Server.ProtoBufEmit(c.FD, msg)
}

func (c *Conn) Close() error {
	_, err := c.Write(udp.CloseMessage)
	return err
}

func (c *Conn) Write(msg []byte) (int, error) {
	if len(msg) > c.Server.ReadBufferSize+udp.HeadLen {
		return 0, errors.New("max length is " + strconv.Itoa(c.Server.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.Server.netListen.WriteToUDP(msg, c.Conn)
}

func (c *Conn) WriteToUDP(msg []byte, addr *net.UDPAddr) (int, error) {
	if len(msg) > c.Server.ReadBufferSize+udp.HeadLen {
		return 0, errors.New("max length is " + strconv.Itoa(c.Server.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.Server.netListen.WriteToUDP(msg, addr)
}
