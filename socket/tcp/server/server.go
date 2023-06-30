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

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/protocol"
	"github.com/lemonyxk/kitty/ssl"
	"github.com/lemonyxk/structure/map"

	"github.com/lemonyxk/kitty/socket"
)

type Server struct {
	Name string
	Addr string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string
	// TLS
	TLSConfig *tls.Config

	OnClose   func(conn Conn)
	OnMessage func(conn Conn, msg []byte)
	OnOpen    func(conn Conn)
	OnError   func(stream *socket.Stream[Conn], err error)
	OnException    func(err error)
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
	senders     *hash.Hash[int64, socket.Emitter[Conn]]
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

func (s *Server) Sender(fd int64) (socket.Emitter[Conn], error) {
	var sender = s.senders.Get(fd)
	if sender == nil {
		return nil, errors.ConnNotFount
	}
	return sender, nil
}

func (s *Server) Ready() {

	if s.Addr == "" {
		panic("addr can not be empty")
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
		s.ReadBufferSize = 8192
	}

	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 8192
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
		s.OnError = func(stream *socket.Stream[Conn], err error) {
			fmt.Println("tcp server err:", err)
		}
	}

	if s.OnException == nil {
		s.OnException = func(err error) {
			fmt.Println("tcp server exception:", err)
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

func (s *Server) Start() {

	s.Ready()

	var err error
	var netListen net.Listener

	if s.CertFile != "" && s.KeyFile != "" || s.TLSConfig != nil {
		var config *tls.Config
		if s.TLSConfig != nil {
			config = s.TLSConfig
		} else {
			config, err = ssl.LoadTLSConfig(s.CertFile, s.KeyFile)
			if err != nil {
				panic(err)
			}
		}
		netListen, err = tls.Listen("tcp", s.Addr, config)
	} else {
		netListen, err = net.Listen("tcp", s.Addr)
	}

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
	if s.HeartBeatTimeout != 0 {
		err := netConn.SetDeadline(time.Now().Add(s.HeartBeatTimeout))
		if err != nil {
			panic(err)
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
			s.OnException(err)
			break
		}

	}

	s.onClose(conn)
}

func (s *Server) decodeMessage(conn Conn, message []byte) error {
	// unpack
	order, messageType, code, id, route, body := conn.UnPack(message)

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
