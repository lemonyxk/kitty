/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-17 20:09
**/

package server

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/udp"
)

type Server struct {
	Name      string
	Host      string
	OnClose   func(conn *Conn)
	OnMessage func(conn *Conn, msg []byte)
	OnOpen    func(conn *Conn)
	OnError   func(err error)
	OnSuccess func()
	OnUnknown func(conn *Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration

	PingHandler func(conn *Conn) func(appData string) error

	PongHandler func(conn *Conn) func(appData string) error

	Protocol udp.Protocol

	fd          int64
	connections map[int64]*Conn
	addrMap     map[string]int64
	mux         sync.RWMutex
	router      *Router
	middle      []func(Middle) Middle
	netListen   *net.UDPConn
	processLock sync.RWMutex
}

type Middle func(conn *Conn, stream *socket.Stream)

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.LocalAddr()
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

	_, err := conn.Write(msg)
	return err
}

func (s *Server) Emit(fd int64, pack socket.Pack) error {
	return s.Push(fd, s.Protocol.Encode(socket.BinData, pack.ID, []byte(pack.Event), pack.Data))
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
	return s.Push(fd, s.Protocol.Encode(socket.BinData, pack.ID, []byte(pack.Event), data))
}

func (s *Server) JsonEmitAll(msg socket.JsonPack) (int, int) {
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

func (s *Server) ProtoBufEmit(fd int64, pack socket.ProtoBufPack) error {
	data, err := proto.Marshal(pack.Data)
	if err != nil {
		return err
	}
	return s.Push(fd, s.Protocol.Encode(socket.BinData, pack.ID, []byte(pack.Event), data))
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
		s.ReadBufferSize = 512
	}

	if s.WriteBufferSize == 0 {
		s.WriteBufferSize = 512
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
		s.Protocol = &udp.DefaultProtocol{}
	}

	if s.PingHandler == nil {
		s.PingHandler = func(connection *Conn) func(appData string) error {
			return func(appData string) error {
				connection.tick.Reset(s.HeartBeatTimeout)
				return nil
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
	s.addrMap = make(map[string]int64)
}

func (s *Server) onOpen(conn *Conn) {
	s.addConnect(conn)
	s.OnOpen(conn)
}

func (s *Server) onClose(conn *Conn) {
	s.delConnect(conn)
	s.OnClose(conn)
	conn.close <- struct{}{}
}

func (s *Server) onError(err error) {
	s.OnError(err)
}

func (s *Server) addConnect(conn *Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.fd++
	s.connections[s.fd] = conn
	s.addrMap[conn.Host()] = s.fd
	conn.FD = s.fd
}

func (s *Server) delConnect(conn *Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.connections, conn.FD)
	delete(s.addrMap, conn.Host())
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
	var netListen *net.UDPConn

	addr, err := net.ResolveUDPAddr("udp", s.Host)
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

	// netListen, err = net.ListenUDP("udp", addr)
	// if err != nil {
	// 	panic(err)
	// }

	s.netListen = netListen

	// start success
	if s.OnSuccess != nil {
		s.OnSuccess()
	}

	// var reader = s.Protocol.Reader()

	for {

		var buffer = make([]byte, s.ReadBufferSize+udp.HeadLen)
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

	s.processLock.Lock()
	defer s.processLock.Unlock()
	fd, ok := s.addrMap[addr.String()]
	conn, ok := s.connections[fd]

	switch message[2] {
	case socket.BinData, socket.PingData, socket.PongData:
		if !ok {
			return
		}
		conn.accept <- message
	case socket.OpenData:
		if ok {
			return
		}

		var conn = &Conn{
			FD:     0,
			Conn:   addr,
			Server: s,
			accept: make(chan []byte, 128),
			close:  make(chan struct{}),
		}

		conn.tick = time.NewTimer(s.HeartBeatTimeout)

		// make sure this goroutine will run over
		go func() {
			for range conn.tick.C {
				_, _ = conn.Write(udp.CloseMessage)
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
					break
				}
			}
		}()

		s.onOpen(conn)

		_, err := conn.Write(udp.OpenMessage)
		if err != nil {
			s.OnError(err)
		}

	case socket.CloseData:
		if !ok {
			return
		}
		s.onClose(conn)
	default:
		return
	}

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
	if messageType == socket.PingData {
		return s.PingHandler(conn)("")
	}

	// Pong
	if messageType == socket.PongData {
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
