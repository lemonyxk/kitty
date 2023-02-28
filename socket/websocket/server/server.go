package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
	hash "github.com/lemonyxk/structure/map"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/lemonyxk/kitty/v2/socket"
)

type Server struct {
	Name string
	// Host 服务Host
	Addr string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string
	Path    string

	OnOpen    func(conn Conn)
	OnMessage func(conn Conn, msg []byte)
	OnClose   func(conn Conn)
	OnError   func(err error)
	OnSuccess func()
	OnRaw     func(w http.ResponseWriter, r *http.Request)
	OnUnknown func(conn Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HandshakeTimeout  time.Duration
	DailTimeout       time.Duration

	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error
	Protocol    protocol.Protocol

	upgrade websocket.Upgrader

	fd          int64
	connections *hash.Hash[int64, Conn]
	router      *router.Router[*socket.Stream[Conn]]
	middle      []func(next Middle) Middle
	server      *http.Server
	netListen   net.Listener
}

type Middle router.Middle[*socket.Stream[Conn]]

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(next Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server) EmitAll(event string, data []byte) (int, int) {
	var counter = 0
	var success = 0

	s.connections.Range(func(fd int64, conn Conn) bool {
		counter++
		if conn.Emit(event, data) == nil {
			success++
		}
		return true
	})

	return counter, success
}

func (s *Server) JsonEmitAll(event string, data any) (int, int) {
	var counter = 0
	var success = 0

	s.connections.Range(func(fd int64, conn Conn) bool {
		counter++
		if conn.JsonEmit(event, data) == nil {
			success++
		}
		return true
	})

	return counter, success
}

func (s *Server) ProtoBufEmitAll(event string, data proto.Message) (int, int) {
	var counter = 0
	var success = 0
	s.connections.Range(func(fd int64, conn Conn) bool {
		counter++
		if conn.ProtoBufEmit(event, data) == nil {
			success++
		}
		return true
	})
	return counter, success
}

func (s *Server) addConnect(conn Conn) {
	var fd = atomic.AddInt64(&s.fd, 1)
	s.connections.Set(fd, conn)
	conn.SetFD(fd)
}

func (s *Server) delConnect(conn Conn) {
	s.connections.Delete(conn.FD())
}

func (s *Server) Range(fn func(conn Conn)) {
	s.connections.Range(func(fd int64, conn Conn) bool {
		fn(conn)
		return true
	})
}

func (s *Server) Conn(fd int64) (Conn, error) {
	conn := s.connections.Get(fd)
	if conn == nil {
		return nil, errors.ConnNotFount
	}
	return conn, nil
}

func (s *Server) ConnLen() int {
	return s.connections.Len()
}

func (s *Server) onOpen(conn Conn) {
	s.addConnect(conn)
	s.OnOpen(conn)
}

func (s *Server) onClose(conn Conn) {
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

	if s.Path == "" {
		s.Path = "/"
	}

	if s.DailTimeout == 0 {
		s.DailTimeout = time.Second * 3
	}

	// if s.HeartBeatTimeout == 0 {
	// 	s.HeartBeatTimeout = 6 * time.Second
	// }
	//
	// if s.HeartBeatInterval == 0 {
	// 	s.HeartBeatInterval = 3 * time.Second
	// }

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
		s.OnOpen = func(conn Conn) {
			fmt.Println("webSocket server:", conn.FD(), "is open")
		}
	}

	if s.OnClose == nil {
		s.OnClose = func(conn Conn) {
			fmt.Println("webSocket server:", conn.FD(), "is close")
		}
	}

	if s.OnError == nil {
		s.OnError = func(err error) {
			fmt.Println("webSocket server:", err)
		}
	}

	if s.Protocol == nil {
		s.Protocol = &protocol.DefaultWsProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(conn Conn) func(data string) error {
			return func(data string) error {
				var err error
				var t = time.Now()
				conn.SetLastPing(t)
				if s.HeartBeatTimeout != 0 {
					err = conn.Conn().SetReadDeadline(t.Add(s.HeartBeatTimeout))
				}
				err = conn.Pong()
				return err
			}
		}
	}

	// no answer
	if s.PongHandler == nil {
		s.PongHandler = func(conn Conn) func(data string) error {
			return func(data string) error {
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

	s.connections = hash.New[int64, Conn]()
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
	if s.HeartBeatTimeout != 0 {
		err = netConn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
		if err != nil {
			s.onError(err)
			return
		}
	}

	var conn = &conn{
		fd:       0,
		conn:     netConn,
		server:   s,
		response: w,
		request:  r,
		lastPing: time.Now(),
		Protocol: s.Protocol,
	}

	// 设置PING处理函数
	netConn.SetPingHandler(s.PingHandler(conn))

	// 设置PONG处理函数
	netConn.SetPongHandler(s.PongHandler(conn))

	// 打开连接 记录
	s.onOpen(conn)

	var reader = s.Protocol.Reader()

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
			_ = s.PingHandler(conn)("")
		}

		err = reader(len(message), message, func(bytes []byte) {
			err = s.decodeMessage(conn, bytes)
		})

		if err != nil {
			s.onError(err)
			break
		}
	}

	// close and clean
	s.onClose(conn)

}

func (s *Server) decodeMessage(conn Conn, message []byte) error {

	// unpack
	messageType, id, route, body := s.Protocol.Decode(message)
	_ = id

	if s.OnMessage != nil {
		s.OnMessage(conn, message)
	}

	if s.Protocol.IsUnknown(messageType) {
		if s.OnUnknown != nil {
			s.OnUnknown(conn, message, s.middleware)
		}
		return nil
	}

	// Ping
	if s.Protocol.IsPing(messageType) {
		return s.PingHandler(conn)("")
	}

	// Pong
	if s.Protocol.IsPong(messageType) {
		return s.PongHandler(conn)("")
	}

	// on router
	s.middleware(socket.NewStream(conn, id, string(route), body))

	return nil
}

func (s *Server) middleware(stream *socket.Stream[Conn]) {
	var next Middle = s.handler
	for i := len(s.middle) - 1; i >= 0; i-- {
		next = s.middle[i](next)
	}
	next(stream)
}

func (s *Server) handler(stream *socket.Stream[Conn]) {

	if s.router == nil {
		if s.OnError != nil {
			s.OnError(errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var n, formatPath = s.router.GetRoute(stream.Event)
	if n == nil {
		if s.OnError != nil {
			s.OnError(errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var nodeData = n.Data

	stream.Params = socket.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if s.OnError != nil {
				s.OnError(err)
			}
			return
		}
	}

	err := nodeData.Function(stream)
	if err != nil {
		if s.OnError != nil {
			s.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](stream); err != nil {
			if s.OnError != nil {
				s.OnError(err)
			}
			return
		}
	}
}

func (s *Server) GetDailTimeout() time.Duration {
	return s.DailTimeout
}

func (s *Server) SetRouter(router *router.Router[*socket.Stream[Conn]]) *Server {
	s.router = router
	return s
}

func (s *Server) GetRouter() *router.Router[*socket.Stream[Conn]] {
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

	if s.KeyFile != "" && s.CertFile != "" {
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

	if r.URL.Path != s.Path {
		if s.OnRaw != nil {
			s.OnRaw(w, r)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.process(w, r)
	return
}
