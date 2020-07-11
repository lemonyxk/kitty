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
	"github.com/lemoyxk/kitty/tcp"
)

type Socket struct {
	FD      int64
	Conn    net.Conn
	Server  *Server
	Context kitty.Context
	mux     sync.RWMutex
}

func (conn *Socket) Host() string {
	return conn.Conn.RemoteAddr().String()
}

func (conn *Socket) ClientIP() string {
	if ip, _, err := net.SplitHostPort(conn.Conn.RemoteAddr().String()); err == nil {
		return ip
	}
	return ""
}

func (conn *Socket) Push(msg []byte) error {
	return conn.Server.Push(conn.FD, msg)
}

func (conn *Socket) Emit(event []byte, body []byte, dataType int, protoType int) error {
	return conn.Server.Emit(conn.FD, event, body, dataType, protoType)
}

func (conn *Socket) JsonEmit(msg kitty.JsonPackage) error {
	return conn.Server.JsonEmit(conn.FD, msg)
}

func (conn *Socket) ProtoBufEmit(msg kitty.ProtoBufPackage) error {
	return conn.Server.ProtoBufEmit(conn.FD, msg)
}

func (conn *Socket) Close() error {
	return conn.Conn.Close()
}

type Server struct {
	Name      string
	Host      string
	OnClose   func(conn *Socket)
	OnMessage func(conn *Socket, messageType int, msg []byte)
	OnOpen    func(conn *Socket)
	OnError   func(err error)
	OnSuccess func()

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	HandshakeTimeout  time.Duration

	PingHandler func(connection *Socket) func(appData string) error

	PongHandler func(connection *Socket) func(appData string) error

	Protocol tcp.Protocol

	fd          int64
	connections map[int64]*Socket
	mux         sync.RWMutex
	router      *Router
	middle      []func(Middle) Middle
	netListen   net.Listener
}

type Middle func(conn *Socket, receive *kitty.ReceivePackage)

func (socket *Server) LocalAddr() net.Addr {
	return socket.netListen.Addr()
}

func (socket *Server) Use(middle ...func(Middle) Middle) {
	socket.middle = append(socket.middle, middle...)
}

func (socket *Server) Push(fd int64, msg []byte) error {
	var conn, ok = socket.GetConnection(fd)
	if !ok {
		return errors.New("client is close")
	}

	conn.mux.Lock()
	defer conn.mux.Unlock()

	_, err := conn.Conn.Write(msg)
	return err
}

func (socket *Server) Emit(fd int64, event []byte, body []byte, dataType int, protoType int) error {
	return socket.Push(fd, socket.Protocol.Encode(event, body, dataType, protoType))
}

func (socket *Server) EmitAll(event []byte, body []byte, dataType int, protoType int) (int, int) {
	var counter = 0
	var success = 0
	for fd := range socket.connections {
		counter++
		if socket.Emit(fd, event, body, dataType, protoType) == nil {
			success++
		}
	}
	return counter, success
}

func (socket *Server) JsonEmitAll(msg kitty.JsonPackage) (int, int) {
	var counter = 0
	var success = 0
	for fd := range socket.connections {
		counter++
		if socket.JsonEmit(fd, msg) == nil {
			success++
		}
	}
	return counter, success
}

func (socket *Server) ProtoBufEmitAll(msg kitty.ProtoBufPackage) (int, int) {
	var counter = 0
	var success = 0
	for fd := range socket.connections {
		counter++
		if socket.ProtoBufEmit(fd, msg) == nil {
			success++
		}
	}
	return counter, success
}

func (socket *Server) ProtoBufEmit(fd int64, msg kitty.ProtoBufPackage) error {
	data, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return socket.Push(fd, socket.Protocol.Encode([]byte(msg.Event), data, kitty.BinData, kitty.ProtoBuf))
}

func (socket *Server) JsonEmit(fd int64, msg kitty.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return socket.Push(fd, socket.Protocol.Encode([]byte(msg.Event), data, kitty.TextData, kitty.Json))
}

func (socket *Server) Ready() {

	if socket.Host == "" {
		panic("Host must set")
	}

	if socket.HandshakeTimeout == 0 {
		socket.HandshakeTimeout = 2 * time.Second
	}

	if socket.HeartBeatTimeout == 0 {
		socket.HeartBeatTimeout = 30 * time.Second
	}

	if socket.HeartBeatInterval == 0 {
		socket.HeartBeatInterval = 15 * time.Second
	}

	if socket.ReadBufferSize == 0 {
		socket.ReadBufferSize = 1024
	}

	if socket.WriteBufferSize == 0 {
		socket.WriteBufferSize = 1024
	}

	if socket.WaitQueueSize == 0 {
		socket.WaitQueueSize = 1024
	}

	if socket.OnOpen == nil {
		socket.OnOpen = func(conn *Socket) {
			println(conn.FD, "is open")
		}
	}

	if socket.OnClose == nil {
		socket.OnClose = func(conn *Socket) {
			println(conn.FD, "is close")
		}
	}

	if socket.OnError == nil {
		socket.OnError = func(err error) {
			println(err)
		}
	}

	if socket.Protocol == nil {
		socket.Protocol = &tcp.DefaultProtocol{}
	}

	if socket.PingHandler == nil {
		socket.PingHandler = func(connection *Socket) func(appData string) error {
			return func(appData string) error {
				// unnecessary
				// err := Server.Push(connection.FD, PongMessage, nil)
				return connection.Conn.SetReadDeadline(time.Now().Add(socket.HeartBeatTimeout))
			}
		}
	}

	if socket.PongHandler == nil {
		socket.PongHandler = func(connection *Socket) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	socket.connections = make(map[int64]*Socket)
}

func (socket *Server) onOpen(conn *Socket) {
	socket.addConnect(conn)
	socket.OnOpen(conn)
}

func (socket *Server) onClose(conn *Socket) {
	_ = conn.Conn.Close()
	socket.delConnect(conn)
	socket.OnClose(conn)
}

func (socket *Server) onError(err error) {
	socket.OnError(err)
}

func (socket *Server) addConnect(conn *Socket) {
	socket.mux.Lock()
	defer socket.mux.Unlock()
	socket.fd++
	socket.connections[socket.fd] = conn
	conn.FD = socket.fd
}

func (socket *Server) delConnect(conn *Socket) {
	socket.mux.Lock()
	defer socket.mux.Unlock()
	delete(socket.connections, conn.FD)
}

func (socket *Server) GetConnections() chan *Socket {
	var ch = make(chan *Socket, 1)
	go func() {
		for _, conn := range socket.connections {
			ch <- conn
		}
		close(ch)
	}()
	return ch
}

func (socket *Server) Close(fd int64) error {
	conn, ok := socket.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (socket *Server) GetConnection(fd int64) (*Socket, bool) {
	socket.mux.RLock()
	defer socket.mux.RUnlock()
	conn, ok := socket.connections[fd]
	return conn, ok
}

func (socket *Server) GetConnectionsCount() int {
	socket.mux.RLock()
	defer socket.mux.RUnlock()
	return len(socket.connections)
}

func (socket *Server) Start() {

	socket.Ready()

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", socket.Host)

	if err != nil {
		panic(err)
	}

	socket.netListen = netListen

	// start success
	if socket.OnSuccess != nil {
		socket.OnSuccess()
	}

	for {
		conn, err := netListen.Accept()
		if err != nil {
			break
		}

		go socket.process(conn)
	}

}

func (socket *Server) Shutdown() error {
	return socket.netListen.Close()
}

func (socket *Server) process(conn net.Conn) {

	// 超时时间
	err := conn.SetReadDeadline(time.Now().Add(socket.HeartBeatTimeout))
	if err != nil {
		socket.onError(err)
		return
	}

	err = conn.(*net.TCPConn).SetReadBuffer(socket.ReadBufferSize)
	if err != nil {
		panic(err)
	}
	err = conn.(*net.TCPConn).SetWriteBuffer(socket.WriteBufferSize)
	if err != nil {
		panic(err)
	}

	var connection = &Socket{
		FD:     0,
		Conn:   conn,
		Server: socket,
	}

	socket.onOpen(connection)

	var reader = socket.Protocol.Reader()

	var buffer = make([]byte, socket.ReadBufferSize)

	for {

		n, err := conn.Read(buffer)

		// close error
		if err != nil {
			break
		}

		err = reader(n, buffer, func(bytes []byte) {
			err = socket.decodeMessage(connection, bytes)
		})

		if err != nil {
			socket.onError(err)
			break
		}

	}

	socket.onClose(connection)
}

func (socket *Server) decodeMessage(connection *Socket, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := socket.Protocol.Decode(message)

	if socket.OnMessage != nil {
		socket.OnMessage(connection, messageType, message)
	}

	// check version
	if version != kitty.Version {
		return nil
	}

	// Ping
	if messageType == kitty.PingData {
		return socket.PingHandler(connection)("")
	}

	// Pong
	if messageType == kitty.PongData {
		return socket.PongHandler(connection)("")
	}

	// on router
	if socket.router != nil {
		socket.middleware(connection, &kitty.ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
		return nil
	}

	return nil
}

func (socket *Server) middleware(conn *Socket, msg *kitty.ReceivePackage) {
	var next Middle = socket.handler
	for i := len(socket.middle) - 1; i >= 0; i-- {
		next = socket.middle[i](next)
	}
	next(conn, msg)
}

func (socket *Server) handler(conn *Socket, msg *kitty.ReceivePackage) {

	var n, formatPath = socket.router.getRoute(msg.Event)
	if n == nil {
		if socket.OnError != nil {
			socket.OnError(errors.New(msg.Event + " " + "404 not found"))
		}
		return
	}

	var nodeData = n.Data.(*node)

	var receive = &kitty.Receive{}
	receive.Body = msg
	receive.Context = nil
	receive.Params = kitty.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		ctx, err := nodeData.Before[i](conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
		receive.Context = ctx
	}

	err := nodeData.Function(conn, receive)
	if err != nil {
		if socket.OnError != nil {
			socket.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
	}

}

func (socket *Server) SetRouter(router *Router) *Server {
	socket.router = router
	return socket
}

func (socket *Server) GetRouter() *Router {
	return socket.router
}
