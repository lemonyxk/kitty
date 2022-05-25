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
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
	"github.com/lemonyxk/structure/v3/map"

	"github.com/lemonyxk/kitty/v2/socket"
)

type Server struct {
	Name string
	Addr string

	OnClose   func(conn Conn)
	OnMessage func(conn Conn, msg []byte)
	OnOpen    func(conn Conn)
	OnError   func(err error)
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

func (s *Server) protocol(fd int64, messageType byte, route []byte, body []byte) error {
	var conn = s.GetConnection(fd)
	if conn == nil {
		return errors.ClientClosed
	}

	err := conn.protocol(messageType, route, body)
	return err
}

func (s *Server) Push(fd int64, msg []byte) error {
	var conn = s.GetConnection(fd)
	if conn == nil {
		return errors.ClientClosed
	}

	_, err := conn.Write(msg)
	return err
}

func (s *Server) Emit(fd int64, pack socket.Pack) error {
	return s.protocol(fd, protocol.Bin, []byte(pack.Event), pack.Data)
}

func (s *Server) EmitAll(pack socket.Pack) (int, int) {
	var counter = 0
	var success = 0
	s.connections.Range(func(fd int64, conn Conn) bool {
		counter++
		if s.Emit(fd, pack) == nil {
			success++
		}
		return true
	})
	return counter, success
}

func (s *Server) JsonEmit(fd int64, pack socket.JsonPack) error {
	data, err := jsoniter.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return s.protocol(fd, protocol.Bin, []byte(pack.Event), data)
}

func (s *Server) JsonEmitAll(msg socket.JsonPack) (int, int) {
	var counter = 0
	var success = 0

	s.connections.Range(func(fd int64, conn Conn) bool {
		counter++
		if s.JsonEmit(fd, msg) == nil {
			success++
		}
		return true
	})

	return counter, success
}

func (s *Server) ProtoBufEmit(fd int64, pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return s.protocol(fd, protocol.Bin, []byte(pack.Event), data)
}

func (s *Server) ProtoBufEmitAll(msg socket.ProtoBufPack) (int, int) {
	var counter = 0
	var success = 0

	s.connections.Range(func(fd int64, conn Conn) bool {
		counter++
		if s.ProtoBufEmit(fd, msg) == nil {
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

	if s.HeartBeatTimeout == 0 {
		s.HeartBeatTimeout = 6 * time.Second
	}

	if s.HeartBeatInterval == 0 {
		s.HeartBeatInterval = 3 * time.Second
	}

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
		s.OnError = func(err error) {
			fmt.Println("udp server:", err)
		}
	}

	if s.Protocol == nil {
		s.Protocol = &protocol.DefaultUdpProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(connection Conn) func(data string) error {
			return func(data string) error {
				var t = time.Now()
				connection.SetLastPing(t)
				connection.Tick().Reset(s.HeartBeatTimeout)
				return connection.Pong()
			}
		}
	}

	// no answer
	if s.PongHandler == nil {
		s.PongHandler = func(connection Conn) func(data string) error {
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

func (s *Server) onError(err error) {
	s.OnError(err)
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

func (s *Server) GetConnections(fn func(conn Conn)) {
	s.connections.Range(func(fd int64, conn Conn) bool {
		fn(conn)
		return true
	})
}

func (s *Server) Close(fd int64) error {
	conn := s.GetConnection(fd)
	if conn == nil {
		return errors.ConnNotFount
	}
	return conn.Close()
}

func (s *Server) GetConnectionByAddr(addr string) Conn {
	fd := s.addrMap.Get(addr)
	conn := s.connections.Get(fd)
	return conn
}

func (s *Server) GetConnection(fd int64) Conn {
	conn := s.connections.Get(fd)
	return conn
}

func (s *Server) GetConnectionsCount() int {
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

	for {

		var buffer = make([]byte, s.ReadBufferSize+s.Protocol.HeadLen())
		n, addr, err := netListen.ReadFromUDP(buffer)

		if err != nil {
			break
		}

		go s.process(addr, buffer[:n])

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
		s.OnError(err)
	}
}

func (s *Server) readMessage(addr *net.UDPAddr, message []byte) error {
	// unpack
	messageType := s.Protocol.GetMessageType(message)

	if s.Protocol.IsPing(messageType) || s.Protocol.IsPong(messageType) {
		var conn = s.GetConnectionByAddr(addr.String())
		if conn == nil {
			return nil
		}
		conn.AcceptChan() <- message

	} else if s.Protocol.IsOpen(messageType) {
		s.processLock.Lock()
		defer s.processLock.Unlock()

		var c = s.GetConnectionByAddr(addr.String())
		if c != nil {
			return nil
		}

		var conn = &conn{
			fd:       0,
			conn:     addr,
			server:   s,
			lastPing: time.Now(),
			accept:   make(chan []byte, 128),
			close:    make(chan struct{}, 1),
		}

		conn.tick = time.NewTimer(s.HeartBeatTimeout)

		// make sure this goroutine will run over
		go func() {
			for range conn.tick.C {
				_ = conn.SendClose()
				s.onClose(conn)
			}
		}()

		// make sure this goroutine will run over
		go func() {
			for {
				select {
				case message := <-conn.accept:
					var err = s.decodeMessage(conn, message)
					if err != nil {
						s.OnError(err)
					}
				case <-conn.close:
					conn.tick.Stop()
					return
				}
			}
		}()

		s.onOpen(conn)

		err := conn.SendOpen()
		if err != nil {
			s.OnError(err)
		}
	} else if s.Protocol.IsClose(messageType) {
		s.processLock.Lock()
		defer s.processLock.Unlock()

		var conn = s.GetConnectionByAddr(addr.String())
		if conn == nil {
			return nil
		}
		s.onClose(conn)
	} else {
		// bin message
		var conn = s.GetConnectionByAddr(addr.String())
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
	s.middleware(&socket.Stream[Conn]{Conn: conn, Pack: socket.Pack{Event: string(route), Data: body}})

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
