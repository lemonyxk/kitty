/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-12 22:27
**/

package server

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/udp"
)

type Conn struct {
	FD     int64
	Conn   *net.UDPAddr
	Server *Server
	mux    sync.RWMutex
	tick   *time.Timer
	accept chan []byte
	close  chan struct{}
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
	_, err := c.Server.netListen.WriteToUDP(udp.CloseMessage, c.Conn)
	return err
}

func (c *Conn) Write(msg []byte) (int, error) {
	if len(msg) > c.Server.ReadBufferSize+udp.HeadLen {
		return 0, errors.New("max length is " + strconv.Itoa(c.Server.ReadBufferSize) + "but now is " + strconv.Itoa(len(msg)))
	}
	return c.Server.netListen.WriteToUDP(msg, c.Conn)
}
