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
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/kitty/header"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

type Conn interface {
	Name() string
	SetName(name string)
	FD() int64
	SetFD(int64)
	Host() string
	ClientIP() string
	Ping() error
	Pong() error
	Write(messageType int, msg []byte) (int, error)
	Close() error
	SetLastPing(time.Time)
	LastPing() time.Time
	Conn() *websocket.Conn
	Server() *Server
	Response() http.ResponseWriter
	Request() *http.Request
	SubProtocols() []string
	SetReadDeadline(t time.Time) error
	socket.Emitter
	protocol.Protocol
}

type conn struct {
	name         string
	fd           int64
	conn         *websocket.Conn
	lastPing     time.Time
	server       *Server
	response     http.ResponseWriter
	request      *http.Request
	mux          sync.Mutex
	messageID    int64
	subProtocols []string
	protocol.Protocol
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *conn) Response() http.ResponseWriter {
	return c.response
}

func (c *conn) Request() *http.Request {
	return c.request
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
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
	if host := c.request.Header.Get(header.Host); host != "" {
		return host
	}
	return c.request.Host
}

func (c *conn) ClientIP() string {

	if ip := strings.Split(c.request.Header.Get(header.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := c.request.Header.Get(header.XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(c.request.RemoteAddr); err == nil {
		return ip
	}

	return ""
}

func (c *conn) SubProtocols() []string {
	return c.subProtocols
}

func (c *conn) Ping() error {
	err := c.Push(c.PackPing())
	return err
}

func (c *conn) Pong() error {
	err := c.Push(c.PackPong())
	return err
}

func (c *conn) Push(msg []byte) error {
	_, err := c.Write(int(protocol.Bin), msg)
	return err
}

func (c *conn) Emit(event string, data []byte) error {
	return c.Pack(protocol.Bin, atomic.AddInt64(&c.messageID, 1), []byte(event), data)
}

func (c *conn) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return c.Pack(protocol.Bin, atomic.AddInt64(&c.messageID, 1), []byte(event), msg)
}

func (c *conn) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return c.Pack(protocol.Bin, atomic.AddInt64(&c.messageID, 1), []byte(event), msg)
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Write(messageType int, msg []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return len(msg), c.conn.WriteMessage(messageType, msg)
}

func (c *conn) Pack(messageType byte, messageID int64, route []byte, body []byte) error {
	var msg = c.Encode(messageType, messageID, route, body)
	_, err := c.Write(int(protocol.Bin), msg)
	return err
}
