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

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
	"github.com/lemonyxk/structure/map"
)

type Server[T any] struct {
	Name string
	Addr string

	OnClose     func(conn Conn)
	OnMessage   func(conn Conn, msg []byte)
	OnOpen      func(conn Conn)
	OnError     func(stream *socket.Stream[Conn], err error)
	OnSuccess   func()
	OnException func(err error)
	OnUnknown   func(conn Conn, message []byte, next Middle)

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HandshakeTimeout  time.Duration
	DailTimeout       time.Duration

	Mtu             int
	ReadBufferSize  int // not use yet
	WriteBufferSize int // not use yet

	PingHandler func(conn Conn) func(data string) error
	PongHandler func(conn Conn) func(data string) error
	Protocol    protocol.UDPProtocol

	fd          int64
	senders     *hash.Hash[int64, socket.Emitter[Conn]]
	addrMap     *hash.Hash[string, int64]
	router      *router.Router[*socket.Stream[Conn], T]
	middle      []func(Middle) Middle
	netListen   *net.UDPConn
	processLock sync.RWMutex
}

type Middle router.Middle[*socket.Stream[Conn]]

func (s *Server[T]) LocalAddr() net.Addr {
	return s.netListen.LocalAddr()
}

func (s *Server[T]) Use(middle ...func(Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server[T]) Sender(fd int64) (socket.Emitter[Conn], error) {
	var sender = s.senders.Get(fd)
	if sender == nil {
		return nil, errors.ConnNotFount
	}
	return sender, nil
}

func (s *Server[T]) Ready() {

	if s.Addr == "" {
		panic("addr can not be empty")
	}

	if s.DailTimeout == 0 {
		s.DailTimeout = time.Second * 3
	}

	if s.HandshakeTimeout == 0 {
		s.HandshakeTimeout = 3 * time.Second
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

	if s.Mtu == 0 {
		s.Mtu = 512
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
			fmt.Println("udp server err:", err)
		}
	}

	if s.OnException == nil {
		s.OnException = func(err error) {
			fmt.Println("udp server exception:", err)
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

	s.senders = hash.New[int64, socket.Emitter[Conn]]()
	s.addrMap = hash.New[string, int64]()
}

func (s *Server[T]) onOpen(conn Conn) {
	s.addConnect(conn)
	s.OnOpen(conn)
}

func (s *Server[T]) onClose(conn Conn) {
	s.delConnect(conn)
	s.OnClose(conn)
	conn.CloseChan() <- struct{}{}
}

func (s *Server[T]) onError(stream *socket.Stream[Conn], err error) {
	s.OnError(stream, err)
}

func (s *Server[T]) addConnect(conn Conn) {
	var fd = atomic.AddInt64(&s.fd, 1)
	s.senders.Set(fd, socket.NewSender(conn))
	s.addrMap.Set(conn.Host(), fd)
	conn.SetFD(fd)
}

func (s *Server[T]) delConnect(conn Conn) {
	s.senders.Delete(conn.FD())
	s.addrMap.Delete(conn.Host())
}

func (s *Server[T]) Range(fn func(conn Conn)) {
	s.senders.Range(func(k int64, v socket.Emitter[Conn]) bool {
		fn(v.Conn())
		return true
	})
}

func (s *Server[T]) ConnByAddr(addr string) (Conn, error) {
	fd := s.addrMap.Get(addr)
	sender := s.senders.Get(fd)
	if sender == nil {
		return nil, errors.ConnNotFount
	}
	return sender.Conn(), nil
}

func (s *Server[T]) Conn(fd int64) (Conn, error) {
	sender := s.senders.Get(fd)
	if sender == nil {
		return nil, errors.ConnNotFount
	}
	return sender.Conn(), nil
}

func (s *Server[T]) ConnLen() int {
	return s.senders.Len()
}

func (s *Server[T]) Start() {

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

	var buffer = make([]byte, s.Mtu+s.Protocol.HeadLen())

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

func (s *Server[T]) Shutdown() error {
	return s.netListen.Close()
}

func (s *Server[T]) process(addr *net.UDPAddr, message []byte) {
	var reader = s.Protocol.Reader()
	var err error
	err = reader(len(message), message, func(bytes []byte) {
		err = s.readMessage(addr, bytes)
	})
	if err != nil {
		s.OnException(err)
	}
}

func (s *Server[T]) readMessage(addr *net.UDPAddr, message []byte) error {
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
			mtu:         s.Mtu,
			netListen:   s.netListen,
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
						s.OnException(err)
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
			s.OnException(err)
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

func (s *Server[T]) decodeMessage(conn Conn, message []byte) error {
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

	// on router
	s.middleware(socket.NewStream(conn, order, messageType, code, id, route, body))

	return nil
}

func (s *Server[T]) middleware(stream *socket.Stream[Conn]) {
	var next Middle = s.handler
	for i := len(s.middle) - 1; i >= 0; i-- {
		next = s.middle[i](next)
	}
	next(stream)
}

func (s *Server[T]) handler(stream *socket.Stream[Conn]) {

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

	//stream.Node = n.Data

	stream.Params = n.ParseParams(formatPath)

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if errors.Is(err, errors.StopPropagation) {
				return
			}
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
			if errors.Is(err, errors.StopPropagation) {
				return
			}
			if s.OnError != nil {
				s.OnError(stream, err)
			}
			return
		}
	}
}

func (s *Server[T]) GetDailTimeout() time.Duration {
	return s.DailTimeout
}

func (s *Server[T]) SetRouter(router *router.Router[*socket.Stream[Conn], T]) *Server[T] {
	s.router = router
	return s
}

func (s *Server[T]) GetRouter() *router.Router[*socket.Stream[Conn], T] {
	return s.router
}
