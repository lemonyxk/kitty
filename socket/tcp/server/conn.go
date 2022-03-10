/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-11 16:22
**/

package server

import (
	"net"
	"sync"
	"time"

	"github.com/lemoyxk/kitty/v2/socket"
)

type Conn struct {
	Name     string
	FD       int64
	Conn     net.Conn
	Server   *Server
	LastPing time.Time
	mux      sync.RWMutex
}

func (c *Conn) Host() string {
	return c.Conn.RemoteAddr().String()
}

func (c *Conn) ClientIP() string {
	if ip, _, err := net.SplitHostPort(c.Conn.RemoteAddr().String()); err == nil {
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
	return c.Conn.Close()
}

func (c *Conn) Write(msg []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.Conn.Write(msg)
}
