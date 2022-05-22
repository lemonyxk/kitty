/**
* @program: kitty
*
* @description:
*
* @author: lemon
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

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/kitty"

	"github.com/lemonyxk/kitty/v2/socket"
)

type Conn interface {
	Name() string
	FD() int64
	SetFD(int64)
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	Push(msg []byte) error
	Emit(pack socket.Pack) error
	JsonEmit(msg socket.JsonPack) error
	ProtoBufEmit(msg socket.ProtoBufPack) error
	Write(messageType int, msg []byte) (int, error)
	Close() error
	SetLastPing(time.Time)
	LastPing() time.Time
	Conn() *websocket.Conn
	Server() *Server
	protocol(messageType byte, route []byte, body []byte) error
}

type conn struct {
	name     string
	fd       int64
	conn     *websocket.Conn
	lastPing time.Time
	server   *Server
	response http.ResponseWriter
	request  *http.Request
	mux      sync.Mutex
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) Server() *Server {
	return c.server
}

func (c *conn) Conn() *websocket.Conn {
	return c.conn
}

func (c *conn) SetLastPing(t time.Time) {
	c.lastPing = t
}

func (c *conn) LastPing() time.Time {
	return c.lastPing
}

func (c *conn) FD() int64 {
	return c.fd
}

func (c *conn) SetFD(fd int64) {
	c.fd = fd
}

func (c *conn) Host() string {
	if host := c.request.Header.Get(kitty.Host); host != "" {
		return host
	}
	return c.request.Host
}

func (c *conn) ClientIP() string {

	if ip := strings.Split(c.request.Header.Get(kitty.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := c.request.Header.Get(kitty.XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(c.request.RemoteAddr); err == nil {
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
	_, err := c.Write(int(socket.Bin), msg)
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
	return c.conn.Close()
}

func (c *conn) Write(messageType int, msg []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	return len(msg), c.conn.WriteMessage(messageType, msg)
}

func (c *conn) protocol(messageType byte, route []byte, body []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	var msg = c.server.Protocol.Encode(messageType, 0, route, body)
	err := c.conn.WriteMessage(int(socket.Bin), msg)
	return err
}
