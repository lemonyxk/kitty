/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-10-17 20:09
**/

package server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/protocol"
	"github.com/lemonyxk/structure/map"

	"github.com/lemonyxk/kitty/socket"
)

type Server struct {
	Name string
	Addr string

	OnClose   func(conn Conn)
	OnMessage func(conn Conn, msg []byte)
	OnOpen    func(conn Conn)
	OnError   func(stream *socket.Stream[Conn], err error)
	OnSuccess func()
	OnUnknown func(conn Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HandshakeTimeout  time.Duration
	DailTimeout       time.Duration

	ReadBufferSize  int
	WriteBufferSize int

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error
	Protocol    protocol.UDPProtocol

	fd          int64
	connections *hash.Hash[int64, Conn]
	addrMap     *hash.Hash[string, int64]
	router      *router.Router[*socket.Stream[Conn]]
	middle      []func(Middle) Middle
	netListen   *net.UDPConn
	processLock sync.RWMutex
}

type Middle router.Middle[*socket.Stream[Conn]]

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.LocalAddr()
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

	if s.HandshakeTimeout == 0 {
		s.HandshakeTimeout = 2 * time.Second
	}

	// if s.HeartBeatTimeout == 0 {
	// 	s.HeartBeatTimeout = 6 * time.Second
	// }
	//
	// if s.HeartBeatInterval == 0 {
	// 	s.HeartBeatInterval = 3 * time.Second
	// }

	if s.ReadBufferSize == 0 {
		s.ReadBufferSize = 512
	}

	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 512
	}

	if s.OnOpen == nil {
		s.OnOpen = func(conn Conn) {
			fmt.Println("udp server:", conn.FD(), "is open")
		}
	}

	if s.OnClose == nil {
		s.OnClose = func(conn Conn) {
			fmt.Println("udp server:", conn.FD(), "is close")
		}
	}

	if s.OnError == nil {
		s.OnError = func(stream *socket.Stream[Conn], err error) {
			fmt.Println("udp server:", err)
		}
	}

	if s.Protocol == nil {
		s.Protocol = &protocol.DefaultUdpProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(conn Conn) func(data string) error {
			return func(data string) error {
				var err error
				var t = time.Now()
				conn.SetLastPing(t)
				if s.HeartBeatTimeout != 0 {
					err = conn.SetReadDeadline(t.Add(s.HeartBeatTimeout))
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
	s.addrMap = hash.New[string, int64]()
}

func (s *Server) onOpen(conn Conn) {
	s.addConnect(conn)
	s.OnOpen(conn)
}

func (s *Server) onClose(conn Conn) {
	s.delConnect(conn)
	s.OnClose(conn)
	conn.CloseChan() <- struct{}{}
}

func (s *Server) onError(stream *socket.Stream[Conn], err error) {
	s.OnError(stream, err)
}

func (s *Server) addConnect(conn Conn) {
	var fd = atomic.AddInt64(&s.fd, 1)
	s.connections.Set(fd, conn)
	s.addrMap.Set(conn.Host(), fd)
	conn.SetFD(fd)
}

func (s *Server) delConnect(conn Conn) {
	s.connections.Delete(conn.FD())
	s.addrMap.Delete(conn.Host())
}

func (s *Server) Range(fn func(conn Conn)) {
	s.connections.Range(func(fd int64, conn Conn) bool {
		fn(conn)
		return true
	})
}

func (s *Server) ConnByAddr(addr string) (Conn, error) {
	fd := s.addrMap.Get(addr)
	conn := s.connections.Get(fd)
	if conn == nil {
		return nil, errors.ConnNotFount
	}
	return conn, nil
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
	var netListen *net.UDPConn

	addr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		panic(err)
	}

	if addr.IP.IsMulticast() {
		netListen, err = net.ListenMulticastUDP("udp", nil, addr)
	} else {
		netListen, err = net.ListenUDP("udp", addr)
	}

	if err != nil {
		panic(err)
	}

	s.netListen = netListen

	// start success
	if s.OnSuccess != nil {
		s.OnSuccess()
	}

	// var reader = s.Protocol.Reader()

	var buffer = make([]byte, s.ReadBufferSize+s.Protocol.HeadLen())

	for {

		n, addr, err := netListen.ReadFromUDP(buffer)

		if err != nil {
			break
		}

		s.process(addr, buffer[:n])

		// split the message
		// create scheduler for every addr
		// then every addr process messages are sortable
	}

}

func (s *Server) Shutdown() error {
	return s.netListen.Close()
}

func (s *Server) process(addr *net.UDPAddr, message []byte) {
	var reader = s.Protocol.Reader()
	var err error
	err = reader(len(message), message, func(bytes []byte) {
		err = s.readMessage(addr, bytes)
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server) readMessage(addr *net.UDPAddr, message []byte) error {
	// unpack
	messageType := s.Protocol.GetMessageType(message)

	if s.Protocol.IsPing(messageType) || s.Protocol.IsPong(messageType) {
		var conn, _ = s.ConnByAddr(addr.String())
		if conn == nil {
			return nil
		}
		conn.AcceptChan() <- message

	} else if s.Protocol.IsOpen(messageType) {
		s.processLock.Lock()
		defer s.processLock.Unlock()

		var c, _ = s.ConnByAddr(addr.String())
		if c != nil {
			return nil
		}

		var conn = &conn{
			fd:          0,
			conn:        addr,
			server:      s,
			lastPing:    time.Now(),
			accept:      make(chan []byte, 128),
			close:       make(chan struct{}, 1),
			UDPProtocol: s.Protocol,
		}

		var heartBeatTimeout = s.HeartBeatTimeout
		if s.HeartBeatTimeout == 0 {
			heartBeatTimeout = time.Second
		}

		conn.timeoutTimer = time.NewTimer(heartBeatTimeout)

		if s.HeartBeatTimeout == 0 {
			conn.timeoutTimer.Stop()
		}

		// make sure this goroutine will run over
		go func() {
			<-conn.timeoutTimer.C
			_ = conn.SendClose()
			s.onClose(conn)
		}()

		// make sure this goroutine will run over
		go func() {
			for {
				select {
				case message := <-conn.accept:
					var err = s.decodeMessage(conn, message)
					if err != nil {
						fmt.Println(err)
					}
				case <-conn.close:
					conn.timeoutTimer.Stop()
					return
				}
			}
		}()

		s.onOpen(conn)

		err := conn.SendOpen()
		if err != nil {
			fmt.Println(err)
		}
	} else if s.Protocol.IsClose(messageType) {
		s.processLock.Lock()
		defer s.processLock.Unlock()

		var conn, _ = s.ConnByAddr(addr.String())
		if conn == nil {
			return nil
		}
		s.onClose(conn)
	} else {
		// bin message
		var conn, _ = s.ConnByAddr(addr.String())
		if conn == nil {
			return nil
		}
		conn.AcceptChan() <- message
	}

	return nil
}

func (s *Server) decodeMessage(conn Conn, message []byte) error {
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
			s.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event))
		}
		return
	}

	var n, formatPath = s.router.GetRoute(stream.Event)
	if n == nil {
		if s.OnError != nil {
			s.OnError(stream, errors.Wrap(errors.RouteNotFount, stream.Event))
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
