/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-11 16:19
**/

package server

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lemoyxk/kitty/kitty"

	"github.com/lemoyxk/kitty/socket"
)

type Conn struct {
	Name     string
	FD       int64
	Conn     *websocket.Conn
	LastPing time.Time
	Server   *Server
	Response http.ResponseWriter
	Request  *http.Request
	mux      sync.Mutex
}

func (c *Conn) Host() string {
	if host := c.Request.Header.Get(kitty.Host); host != "" {
		return host
	}
	return c.Request.Host
}

func (c *Conn) ClientIP() string {

	if ip := strings.Split(c.Request.Header.Get(kitty.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := c.Request.Header.Get(kitty.XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
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

func (c *Conn) Write(messageType int, msg []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	return len(msg), c.Conn.WriteMessage(messageType, msg)
}
