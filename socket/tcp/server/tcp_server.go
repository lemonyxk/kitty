/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-09 14:06
**/

package server

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/golang/protobuf/proto"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/tcp"
)

type Conn struct {
	FD     int64
	Conn   net.Conn
	Server *Server
	mux    sync.RWMutex
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

func (c *Conn) Push(msg []byte) error {
	return c.Server.Push(c.FD, msg)
}

func (c *Conn) Emit(event []byte, body []byte, dataType int, protoType int) error {
	return c.Server.Emit(c.FD, event, body, dataType, protoType)
}

func (c *Conn) JsonEmit(msg socket.JsonPackage) error {
	return c.Server.JsonEmit(c.FD, msg)
}

func (c *Conn) ProtoBufEmit(msg socket.ProtoBufPackage) error {
	return c.Server.ProtoBufEmit(c.FD, msg)
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

type Server struct {
	Name      string
	Host      string
	OnClose   func(conn *Conn)
	OnMessage func(conn *Conn, messageType int, msg []byte)
	OnOpen    func(conn *Conn)
	OnError   func(err error)
	OnSuccess func()

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	HandshakeTimeout  time.Duration

	PingHandler func(conn *Conn) func(appData string) error

	PongHandler func(conn *Conn) func(appData string) error

	Protocol tcp.Protocol

	fd          int64
	connections map[int64]*Conn
	mux         sync.RWMutex
	router      *Router
	middle      []func(Middle) Middle
	netListen   net.Listener
}

type Middle func(conn *Conn, stream *socket.Stream)

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server) Push(fd int64, msg []byte) error {
	var conn, ok = s.GetConnection(fd)
	if !ok {
		return errors.New("client is close")
	}

	conn.mux.Lock()
	defer conn.mux.Unlock()

	_, err := conn.Conn.Write(msg)
	return err
}

func (s *Server) Emit(fd int64, event []byte, body []byte, dataType int, protoType int) error {
	return s.Push(fd, s.Protocol.Encode(event, body, dataType, protoType))
}

func (s *Server) EmitAll(event []byte, body []byte, dataType int, protoType int) (int, int) {
	var counter = 0
	var success = 0
	for fd := range s.connections {
		counter++
		if s.Emit(fd, event, body, dataType, protoType) == nil {
			success++
		}
	}
	return counter, success
}

func (s *Server) JsonEmitAll(msg socket.JsonPackage) (int, int) {
	var counter = 0
	var success = 0
	for fd := range s.connections {
		counter++
		if s.JsonEmit(fd, msg) == nil {
			success++
		}
	}
	return counter, success
}

func (s *Server) ProtoBufEmitAll(msg socket.ProtoBufPackage) (int, int) {
	var counter = 0
	var success = 0
	for fd := range s.connections {
		counter++
		if s.ProtoBufEmit(fd, msg) == nil {
			success++
		}
	}
	return counter, success
}

func (s *Server) ProtoBufEmit(fd int64, msg socket.ProtoBufPackage) error {
	data, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return s.Push(fd, s.Protocol.Encode([]byte(msg.Event), data, socket.BinData, socket.ProtoBuf))
}

func (s *Server) JsonEmit(fd int64, msg socket.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return s.Push(fd, s.Protocol.Encode([]byte(msg.Event), data, socket.TextData, socket.Json))
}

func (s *Server) Ready() {

	if s.Host == "" {
		panic("Host must set")
	}

	if s.HandshakeTimeout == 0 {
		s.HandshakeTimeout = 2 * time.Second
	}

	if s.HeartBeatTimeout == 0 {
		s.HeartBeatTimeout = 30 * time.Second
	}

	if s.HeartBeatInterval == 0 {
		s.HeartBeatInterval = 15 * time.Second
	}

	if s.ReadBufferSize == 0 {
		s.ReadBufferSize = 1024
	}

	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 1024
	}

	if s.WaitQueueSize == 0 {
		s.WaitQueueSize = 1024
	}

	if s.OnOpen == nil {
		s.OnOpen = func(conn *Conn) {
			println(conn.FD, "is open")
		}
	}

	if s.OnClose == nil {
		s.OnClose = func(conn *Conn) {
			println(conn.FD, "is close")
		}
	}

	if s.OnError == nil {
		s.OnError = func(err error) {
			println(err)
		}
	}

	if s.Protocol == nil {
		s.Protocol = &tcp.DefaultProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(connection *Conn) func(appData string) error {
			return func(appData string) error {
				// unnecessary
				// err := Server.Push(connection.FD, PongMessage, nil)
				return connection.Conn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
			}
		}
	}

	if s.PongHandler == nil {
		s.PongHandler = func(connection *Conn) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	s.connections = make(map[int64]*Conn)
}

func (s *Server) onOpen(conn *Conn) {
	s.addConnect(conn)
	s.OnOpen(conn)
}

func (s *Server) onClose(conn *Conn) {
	_ = conn.Conn.Close()
	s.delConnect(conn)
	s.OnClose(conn)
}

func (s *Server) onError(err error) {
	s.OnError(err)
}

func (s *Server) addConnect(conn *Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.fd++
	s.connections[s.fd] = conn
	conn.FD = s.fd
}

func (s *Server) delConnect(conn *Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.connections, conn.FD)
}

func (s *Server) GetConnections() chan *Conn {
	var ch = make(chan *Conn, 1)
	go func() {
		for _, conn := range s.connections {
			ch <- conn
		}
		close(ch)
	}()
	return ch
}

func (s *Server) Close(fd int64) error {
	conn, ok := s.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (s *Server) GetConnection(fd int64) (*Conn, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	conn, ok := s.connections[fd]
	return conn, ok
}

func (s *Server) GetConnectionsCount() int {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.connections)
}

func (s *Server) Start() {

	s.Ready()

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", s.Host)

	if err != nil {
		panic(err)
	}

	s.netListen = netListen

	// start success
	if s.OnSuccess != nil {
		s.OnSuccess()
	}

	for {
		conn, err := netListen.Accept()
		if err != nil {
			break
		}

		go s.process(conn)
	}

}

func (s *Server) Shutdown() error {
	return s.netListen.Close()
}

func (s *Server) process(netConn net.Conn) {

	// 超时时间
	err := netConn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
	if err != nil {
		s.onError(err)
		return
	}

	err = netConn.(*net.TCPConn).SetReadBuffer(s.ReadBufferSize)
	if err != nil {
		panic(err)
	}
	err = netConn.(*net.TCPConn).SetWriteBuffer(s.WriteBufferSize)
	if err != nil {
		panic(err)
	}

	var conn = &Conn{
		FD:     0,
		Conn:   netConn,
		Server: s,
	}

	s.onOpen(conn)

	var reader = s.Protocol.Reader()

	var buffer = make([]byte, s.ReadBufferSize)

	for {

		n, err := netConn.Read(buffer)

		// close error
		if err != nil {
			break
		}

		err = reader(n, buffer, func(bytes []byte) {
			err = s.decodeMessage(conn, bytes)
		})

		if err != nil {
			s.onError(err)
			break
		}

	}

	s.onClose(conn)
}

func (s *Server) decodeMessage(conn *Conn, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := s.Protocol.Decode(message)

	if s.OnMessage != nil {
		s.OnMessage(conn, messageType, message)
	}

	// check version
	if version != socket.Version {
		return nil
	}

	// Ping
	if messageType == socket.PingData {
		return s.PingHandler(conn)("")
	}

	// Pong
	if messageType == socket.PongData {
		return s.PongHandler(conn)("")
	}

	// on router
	if s.router != nil {
		s.middleware(conn, &socket.Stream{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
		return nil
	}

	return nil
}

func (s *Server) middleware(conn *Conn, stream *socket.Stream) {
	var next Middle = s.handler
	for i := len(s.middle) - 1; i >= 0; i-- {
		next = s.middle[i](next)
	}
	next(conn, stream)
}

func (s *Server) handler(conn *Conn, stream *socket.Stream) {

	var n, formatPath = s.router.getRoute(stream.Event)
	if n == nil {
		if s.OnError != nil {
			s.OnError(errors.New(stream.Event + " " + "404 not found"))
		}
		return
	}

	var nodeData = n.Data.(*node)

	stream.Params = kitty.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](conn, stream); err != nil {
			if s.OnError != nil {
				s.OnError(err)
			}
			return
		}
	}

	err := nodeData.Function(conn, stream)
	if err != nil {
		if s.OnError != nil {
			s.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](conn, stream); err != nil {
			if s.OnError != nil {
				s.OnError(err)
			}
			return
		}
	}

}

func (s *Server) SetRouter(router *Router) *Server {
	s.router = router
	return s
}

func (s *Server) GetRouter() *Router {
	return s.router
}
