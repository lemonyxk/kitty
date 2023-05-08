package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/kitty/header"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/protocol"
	hash "github.com/lemonyxk/structure/map"

	"github.com/lemonyxk/kitty/socket"
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
	OnError   func(stream *socket.Stream[Conn], err error)
	OnException    func(err error)
	OnSuccess func()
	OnRaw     func(w http.ResponseWriter, r *http.Request)
	OnUnknown func(conn Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HandshakeTimeout  time.Duration
	DailTimeout       time.Duration

	ReadBufferSize  int
	WriteBufferSize int

	SubProtocols []string
	CheckOrigin  func(r *http.Request) bool
	PingHandler  func(conn Conn) func(data string) error
	PongHandler  func(conn Conn) func(data string) error

	fd          int64
	senders     *hash.Hash[int64, socket.Emitter[Conn]]
	connections *hash.Hash[int64, Conn]
	router      *router.Router[*socket.Stream[Conn]]
	middle      []func(next Middle) Middle
	server      *http.Server
	netListen   net.Listener
	protocol    protocol.Protocol
}

type Middle router.Middle[*socket.Stream[Conn]]

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(next Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server) Sender(fd int64) (socket.Emitter[Conn], error) {
	var sender = s.senders.Get(fd)
	if sender == nil {
		return nil, errors.ConnNotFount
	}
	return sender, nil
}

func (s *Server) addConnect(conn Conn) {
	var fd = atomic.AddInt64(&s.fd, 1)
	s.connections.Set(fd, conn)
	s.senders.Set(fd, socket.NewSender(conn))
	conn.SetFD(fd)
}

func (s *Server) delConnect(conn Conn) {
	s.connections.Delete(conn.FD())
	s.senders.Delete(conn.FD())
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

func (s *Server) onError(stream *socket.Stream[Conn], err error) {
	s.OnError(stream, err)
}

func (s *Server) Ready() {

	if s.Addr == "" {
		panic("addr can not be empty")
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
		s.ReadBufferSize = 4096
	}
	// suggest 4096
	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 4096
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
		s.OnError = func(stream *socket.Stream[Conn], err error) {
			fmt.Println("webSocket server err:", err)
		}
	}

	if s.OnException == nil {
		s.OnException = func(err error) {
			fmt.Println("webSocket server exception:", err)
		}
	}

	if s.protocol == nil {
		s.protocol = &protocol.DefaultWsProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(conn Conn) func(data string) error {
			return func(data string) error {
				var err error
				var t = time.Now()
				conn.SetLastPing(t)
				if s.HeartBeatTimeout != 0 {
					err = conn.SetDeadline(t.Add(s.HeartBeatTimeout))
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

	s.connections = hash.New[int64, Conn]()
	s.senders = hash.New[int64, socket.Emitter[Conn]]()
}

func (s *Server) process(w http.ResponseWriter, r *http.Request) {
	var upgrade = websocket.Upgrader{
		HandshakeTimeout: s.HandshakeTimeout,
		ReadBufferSize:   s.ReadBufferSize,
		WriteBufferSize:  s.WriteBufferSize,
		CheckOrigin:      s.CheckOrigin,
		Subprotocols:     s.SubProtocols,
	}

	var swp = strings.Split(r.Header.Get(header.SecWebsocketProtocol), ",")
	for i := 0; i < len(swp); i++ {
		upgrade.Subprotocols = append(upgrade.Subprotocols, strings.TrimSpace(swp[i]))
	}

	netConn, err := upgrade.Upgrade(w, r, nil)

	if err != nil {
		s.OnException(err)
		return
	}

	if s.HeartBeatTimeout != 0 {
		err = netConn.NetConn().SetDeadline(time.Now().Add(s.HeartBeatTimeout))
		if err != nil {
			s.OnException(err)
			return
		}
	}

	var conn = &conn{
		fd:           0,
		conn:         netConn,
		server:       s,
		response:     w,
		request:      r,
		lastPing:     time.Now(),
		subProtocols: upgrade.Subprotocols,
		Protocol:     s.protocol,
	}

	netConn.SetPingHandler(s.PingHandler(conn))

	netConn.SetPongHandler(s.PongHandler(conn))

	s.onOpen(conn)

	var reader = s.protocol.Reader()

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
			s.OnException(err)
			break
		}
	}

	// close and clean
	s.onClose(conn)

}

func (s *Server) decodeMessage(conn Conn, message []byte) error {

	// unpack
	order, messageType, code, id, route, body := conn.UnPack(message)

	if s.OnMessage != nil {
		s.OnMessage(conn, message)
	}

	if s.protocol.IsUnknown(messageType) {
		if s.OnUnknown != nil {
			s.OnUnknown(conn, message, s.middleware)
		}
		return nil
	}

	// Ping
	if s.protocol.IsPing(messageType) {
		return s.PingHandler(conn)("")
	}

	// Pong
	if s.protocol.IsPong(messageType) {
		return s.PongHandler(conn)("")
	}

	// on router
	s.middleware(socket.NewStream(conn, order, messageType, code, id, route, body))

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
			s.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event()))
		}
		return
	}

	var n, formatPath = s.router.GetRoute(stream.Event())
	if n == nil {
		if s.OnError != nil {
			s.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event()))
		}
		return
	}

	var nodeData = n.Data

	stream.Params = socket.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if s.OnError != nil {
				s.OnError(stream, err)
			}
			return
		}
	}

	err := nodeData.Function(stream)
	if err != nil {
		if s.OnError != nil {
			s.OnError(stream, err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](stream); err != nil {
			if s.OnError != nil {
				s.OnError(stream, err)
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
		s.OnException(err)
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
