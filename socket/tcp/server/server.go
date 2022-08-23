/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-10-09 14:06
**/

package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
	"github.com/lemonyxk/kitty/v2/ssl"
	"github.com/lemonyxk/structure/v3/map"

	"github.com/golang/protobuf/proto"

	"github.com/lemonyxk/kitty/v2/socket"
)

type Server struct {
	Name string
	Addr string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	OnClose   func(conn Conn)
	OnMessage func(conn Conn, msg []byte)
	OnOpen    func(conn Conn)
	OnError   func(err error)
	OnSuccess func()
	OnUnknown func(conn Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	DailTimeout       time.Duration

	ReadBufferSize  int
	WriteBufferSize int

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error
	Protocol    protocol.Protocol

	fd          int64
	connections *hash.Hash[int64, Conn]
	router      *router.Router[*socket.Stream[Conn]]
	middle      []func(Middle) Middle
	netListen   net.Listener
}

type Middle router.Middle[*socket.Stream[Conn]]

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(Middle) Middle) {
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

func (s *Server) Ready() {

	if s.Addr == "" {
		panic("Addr must set")
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

	if s.ReadBufferSize == 0 {
		s.ReadBufferSize = 1024
	}

	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 1024
	}

	if s.OnOpen == nil {
		s.OnOpen = func(conn Conn) {
			fmt.Println("tcp server:", conn.FD(), "is open")
		}
	}

	if s.OnClose == nil {
		s.OnClose = func(conn Conn) {
			fmt.Println("tcp server:", conn.FD(), "is close")
		}
	}

	if s.OnError == nil {
		s.OnError = func(err error) {
			fmt.Println("tcp server:", err)
		}
	}

	if s.Protocol == nil {
		s.Protocol = &protocol.DefaultTcpProtocol{}
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

	s.connections = hash.New[int64, Conn]()
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

func (s *Server) Start() {

	s.Ready()

	var err error
	var netListen net.Listener

	if s.CertFile != "" && s.KeyFile != "" {
		var config, err = ssl.NewTLSConfig(s.CertFile, s.KeyFile)
		if err != nil {
			panic(err)
		}
		netListen, err = tls.Listen("tcp", s.Addr, config)
	} else {
		netListen, err = net.Listen("tcp", s.Addr)
	}

	// netListen, err = net.Listen("tcp", s.Addr)

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
	if s.HeartBeatTimeout != 0 {
		err := netConn.SetReadDeadline(time.Now().Add(s.HeartBeatTimeout))
		if err != nil {
			s.onError(err)
			return
		}
	}

	switch netConn.(type) {
	case *tls.Conn:
		err := netConn.(*tls.Conn).NetConn().(*net.TCPConn).SetReadBuffer(s.ReadBufferSize)
		if err != nil {
			panic(err)
		}

		err = netConn.(*tls.Conn).NetConn().(*net.TCPConn).SetWriteBuffer(s.WriteBufferSize)
		if err != nil {
			panic(err)
		}
	case *net.TCPConn:
		err := netConn.(*net.TCPConn).SetReadBuffer(s.ReadBufferSize)
		if err != nil {
			panic(err)
		}

		err = netConn.(*net.TCPConn).SetWriteBuffer(s.WriteBufferSize)
		if err != nil {
			panic(err)
		}
	}

	var conn = &conn{
		fd:       0,
		conn:     netConn,
		server:   s,
		lastPing: time.Now(),
		Protocol: s.Protocol,
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

	stream.Params = kitty.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

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
