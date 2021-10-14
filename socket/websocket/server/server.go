package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/json-iterator/go"
	"github.com/lemoyxk/kitty/kitty"

	"github.com/lemoyxk/kitty/socket"
	websocket2 "github.com/lemoyxk/kitty/socket/websocket"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Server struct {
	Name string
	// Host 服务Host
	Addr string
	// Protocol 协议
	TLS bool
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	OnOpen    func(conn *Conn)
	OnMessage func(conn *Conn, msg []byte)
	OnClose   func(conn *Conn)
	OnError   func(err error)
	OnSuccess func()
	OnUnknown func(conn *Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HandshakeTimeout  time.Duration

	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool

	PingHandler func(conn *Conn) func(appData string) error
	PongHandler func(conn *Conn) func(appData string) error
	Protocol    websocket2.Protocol

	upgrade websocket.Upgrader

	fd          int64
	connections map[int64]*Conn
	mux         sync.RWMutex
	router      *Router
	middle      []func(next Middle) Middle
	server      *http.Server
	netListen   net.Listener
}

type Middle func(conn *Conn, stream *socket.Stream)

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(next Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server) Push(fd int64, msg []byte) error {
	var conn, ok = s.GetConnection(fd)
	if !ok {
		return errors.New("client is close")
	}

	_, err := conn.Write(int(socket.Bin), msg)
	return err
}

func (s *Server) Emit(fd int64, pack socket.Pack) error {
	return s.Push(fd, s.Protocol.Encode(socket.Bin, pack.ID, []byte(pack.Event), pack.Data))
}

func (s *Server) EmitAll(pack socket.Pack) (int, int) {
	var counter = 0
	var success = 0
	for fd := range s.connections {
		counter++
		if s.Emit(fd, pack) == nil {
			success++
		}
	}
	return counter, success
}

func (s *Server) JsonEmit(fd int64, pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return s.Push(fd, s.Protocol.Encode(socket.Bin, pack.ID, []byte(pack.Event), data))
}

func (s *Server) JsonEmitAll(pack socket.JsonPack) (int, int) {
	var counter = 0
	var success = 0
	for fd := range s.connections {
		counter++
		if s.JsonEmit(fd, pack) == nil {
			success++
		}
	}
	return counter, success
}

func (s *Server) ProtoBufEmit(fd int64, pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return s.Push(fd, s.Protocol.Encode(socket.Bin, pack.ID, []byte(pack.Event), data))
}

func (s *Server) ProtoBufEmitAll(msg socket.ProtoBufPack) (int, int) {
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

func (s *Server) Close(fd int64) error {
	conn, ok := s.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (s *Server) onOpen(conn *Conn) {
	s.addConnect(conn)
	s.OnOpen(conn)
}

func (s *Server) onClose(conn *Conn) {
	_ = conn.Close()
	s.delConnect(conn)
	s.OnClose(conn)
}

func (s *Server) onError(err error) {
	s.OnError(err)
}

func (s *Server) Ready() {

	if s.Addr == "" {
		panic("Addr must set")
	}

	if s.HeartBeatTimeout == 0 {
		s.HeartBeatTimeout = 6 * time.Second
	}

	if s.HeartBeatInterval == 0 {
		s.HeartBeatInterval = 3 * time.Second
	}

	if s.HandshakeTimeout == 0 {
		s.HandshakeTimeout = 2 * time.Second
	}

	// suggest 4096
	if s.ReadBufferSize == 0 {
		s.ReadBufferSize = 1024
	}
	// suggest 4096
	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 1024
	}

	if s.CheckOrigin == nil {
		s.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	if s.OnOpen == nil {
		s.OnOpen = func(conn *Conn) {
			fmt.Println("webSocket server:", conn.FD, "is open")
		}
	}

	if s.OnClose == nil {
		s.OnClose = func(conn *Conn) {
			fmt.Println("webSocket server:", conn.FD, "is close")
		}
	}

	if s.OnError == nil {
		s.OnError = func(err error) {
			fmt.Println("webSocket server:", err)
		}
	}

	if s.Protocol == nil {
		s.Protocol = &websocket2.DefaultProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(connection *Conn) func(appData string) error {
			return func(appData string) error {
				// back pong
				var err = connection.Push(connection.Server.Protocol.Encode(socket.Pong, 0, nil, nil))
				err = connection.Conn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
				return err
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

	s.upgrade = websocket.Upgrader{
		HandshakeTimeout: s.HandshakeTimeout,
		ReadBufferSize:   s.ReadBufferSize,
		WriteBufferSize:  s.WriteBufferSize,
		CheckOrigin:      s.CheckOrigin,
	}

	s.connections = make(map[int64]*Conn)
}

func (s *Server) process(w http.ResponseWriter, r *http.Request) {

	// 升级协议
	netConn, err := s.upgrade.Upgrade(w, r, nil)

	// 错误处理
	if err != nil {
		s.onError(err)
		return
	}

	// 超时时间
	err = netConn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
	if err != nil {
		s.onError(err)
		return
	}

	var conn = &Conn{
		FD:       0,
		Conn:     netConn,
		Server:   s,
		Response: w,
		Request:  r,
	}

	// 设置PING处理函数
	netConn.SetPingHandler(s.PingHandler(conn))

	// 设置PONG处理函数
	netConn.SetPongHandler(s.PongHandler(conn))

	// 打开连接 记录
	s.onOpen(conn)

	// 收到消息 处理 单一连接接受不冲突 但是不能并发写入
	for {

		// read message
		_, message, err := netConn.ReadMessage()
		// close
		if err != nil {
			break
		}

		// do not let it dead
		// for web ping
		if len(message) == 0 {
			_ = netConn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
		}

		err = s.decodeMessage(conn, message)
		if err != nil {
			s.onError(err)
			break
		}

	}

	// close and clean
	s.onClose(conn)

}

func (s *Server) decodeMessage(conn *Conn, message []byte) error {

	// unpack
	messageType, id, route, body := s.Protocol.Decode(message)

	if s.OnMessage != nil {
		s.OnMessage(conn, message)
	}

	if messageType == socket.Unknown {
		if s.OnUnknown != nil {
			s.OnUnknown(conn, message, s.middleware)
		}
		return nil
	}

	// Ping
	if messageType == socket.Ping {
		return s.PingHandler(conn)("")
	}

	// Pong
	if messageType == socket.Pong {
		return s.PongHandler(conn)("")
	}

	// on router
	s.middleware(conn, &socket.Stream{Pack: socket.Pack{Event: string(route), Data: body, ID: id}})

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

	if s.router == nil {
		if s.OnError != nil {
			s.OnError(errors.New(stream.Event + " " + "404 not found"))
		}
		return
	}

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

func (s *Server) Start() {

	s.Ready()

	var server = http.Server{Addr: s.Addr, Handler: s}

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", server.Addr)

	if err != nil {
		panic(err)
	}

	s.netListen = netListen
	s.server = &server

	// start success
	if s.OnSuccess != nil {
		s.OnSuccess()
	}

	if s.TLS {
		err = server.ServeTLS(netListen, s.CertFile, s.KeyFile)
	} else {
		err = server.Serve(netListen)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Match the webSocket router
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.process(w, r)
	return
}
