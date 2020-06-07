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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/tcp"
	"github.com/Lemo-yxk/lemo/utils"

	"github.com/golang/protobuf/proto"
)

type Socket struct {
	FD      uint32
	Conn    net.Conn
	Server  *Server
	Context lemo.Context
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

func (conn *Socket) JsonEmit(msg lemo.JsonPackage) error {
	return conn.Server.JsonEmit(conn.FD, msg)
}

func (conn *Socket) ProtoBufEmit(msg lemo.ProtoBufPackage) error {
	return conn.Server.ProtoBufEmit(conn.FD, msg)
}

func (conn *Socket) Close() error {
	return conn.Conn.Close()
}

type Server struct {
	Name      string
	Host      string
	Port      int
	IP        string
	AutoBind  bool
	OnClose   func(conn *Socket)
	OnMessage func(conn *Socket, messageType int, msg []byte)
	OnOpen    func(conn *Socket)
	OnError   func(err exception.Error)
	OnSuccess func()

	HeartBeatTimeout  int
	HeartBeatInterval int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	HandshakeTimeout  int

	PingHandler func(connection *Socket) func(appData string) error

	PongHandler func(connection *Socket) func(appData string) error

	Protocol tcp.Protocol

	// 连接
	connOpen chan *Socket

	// 关闭
	connClose chan *Socket

	// 写入
	connPush chan *lemo.PushPackage

	// 返回
	connBack chan error

	// 错误
	connError chan exception.Error

	fd          uint32
	count       uint32
	connections sync.Map
	router      *Router
	middle      []func(Middle) Middle

	netListen net.Listener
	shutdown  chan bool
}

type Middle func(conn *Socket, receive *lemo.ReceivePackage)

func (socket *Server) LocalAddr() net.Addr {
	return socket.netListen.Addr()
}

func (socket *Server) Use(middle ...func(Middle) Middle) {
	socket.middle = append(socket.middle, middle...)
}

// Push 发送消息
func (socket *Server) Push(fd uint32, msg []byte) error {

	socket.connPush <- &lemo.PushPackage{
		FD:   fd,
		Data: msg,
	}

	return <-socket.connBack
}

func (socket *Server) Emit(fd uint32, event []byte, body []byte, dataType int, protoType int) error {
	return socket.Push(fd, socket.Protocol.Encode(event, body, dataType, protoType))
}

func (socket *Server) EmitAll(event []byte, body []byte, dataType int, protoType int) (int, int) {

	var counter = 0
	var success = 0
	socket.connections.Range(func(key, value interface{}) bool {
		counter++
		if socket.Emit(key.(uint32), event, body, dataType, protoType) == nil {
			success++
		}
		return true
	})
	return counter, success

}

func (socket *Server) JsonEmitAll(msg lemo.JsonPackage) (int, int) {
	var counter = 0
	var success = 0
	socket.connections.Range(func(key, value interface{}) bool {
		counter++
		if socket.JsonEmit(key.(uint32), msg) == nil {
			success++
		}
		return true
	})
	return counter, success
}

func (socket *Server) ProtoBufEmitAll(msg lemo.ProtoBufPackage) (int, int) {
	var counter = 0
	var success = 0
	socket.connections.Range(func(key, value interface{}) bool {
		counter++
		if socket.ProtoBufEmit(key.(uint32), msg) == nil {
			success++
		}
		return true
	})
	return counter, success
}

func (socket *Server) ProtoBufEmit(fd uint32, msg lemo.ProtoBufPackage) error {

	data, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}

	return socket.Push(fd, socket.Protocol.Encode([]byte(msg.Event), data, lemo.BinData, lemo.ProtoBuf))

}

func (socket *Server) JsonEmit(fd uint32, msg lemo.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return socket.Push(fd, socket.Protocol.Encode([]byte(msg.Event), data, lemo.TextData, lemo.Json))
}

func (socket *Server) Ready() {

	if socket.HandshakeTimeout == 0 {
		socket.HandshakeTimeout = 2
	}

	if socket.HeartBeatTimeout == 0 {
		socket.HeartBeatTimeout = 30
	}

	if socket.HeartBeatInterval == 0 {
		socket.HeartBeatInterval = 15
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
			console.Println(conn.FD, "is open")
		}
	}

	if socket.OnClose == nil {
		socket.OnClose = func(conn *Socket) {
			console.Println(conn.FD, "is close")
		}
	}

	if socket.OnError == nil {
		socket.OnError = func(err exception.Error) {
			console.Error(err)
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
				return connection.Conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
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

	socket.shutdown = make(chan bool)

	// 连接
	socket.connOpen = make(chan *Socket, socket.WaitQueueSize)

	// 关闭
	socket.connClose = make(chan *Socket, socket.WaitQueueSize)

	// 写入
	socket.connPush = make(chan *lemo.PushPackage, socket.WaitQueueSize)

	// 返回
	socket.connBack = make(chan error, socket.WaitQueueSize)

	// 错误
	socket.connError = make(chan exception.Error, socket.WaitQueueSize)

	go func() {
		for {
			select {
			case conn := <-socket.connOpen:
				socket.addConnect(conn)
				socket.count++
				// 触发OPEN事件
				go socket.OnOpen(conn)
			case conn := <-socket.connClose:
				_ = conn.Conn.Close()
				socket.delConnect(conn)
				socket.count--
				// 触发CLOSE事件
				go socket.OnClose(conn)
			case push := <-socket.connPush:
				var conn, ok = socket.connections.Load(push.FD)
				if !ok {
					socket.connBack <- errors.New("client " + strconv.Itoa(int(push.FD)) + " is close")
				} else {
					socket.connBack <- exception.Inspect(conn.(*Socket).Conn.Write(push.Data))
				}
			case err := <-socket.connError:
				go socket.OnError(err)
			}
		}
	}()
}

func (socket *Server) addConnect(conn *Socket) {

	// +1
	socket.fd++

	// 溢出
	if socket.fd == 0 {
		socket.fd++
	}

	var _, ok = socket.connections.Load(socket.fd)

	if !ok {
		socket.connections.Store(socket.fd, conn)
	} else {
		// 否则查找最大值
		var maxFd uint32 = 0

		for {

			maxFd++

			if maxFd == 0 {
				console.Println("connections overflow")
				return
			}

			var _, ok = socket.connections.Load(socket.fd)

			if !ok {
				socket.connections.Store(maxFd, conn)
				break
			}

		}

		socket.fd = maxFd
	}

	// 赋值
	conn.FD = socket.fd

}

func (socket *Server) delConnect(conn *Socket) {
	socket.connections.Delete(conn.FD)
}

func (socket *Server) GetConnections() chan *Socket {
	var ch = make(chan *Socket, 1)
	go func() {
		socket.connections.Range(func(key, value interface{}) bool {
			ch <- value.(*Socket)
			return true
		})
		close(ch)
	}()
	return ch
}

func (socket *Server) Close(fd uint32) error {
	conn, ok := socket.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (socket *Server) GetConnection(fd uint32) (*Socket, bool) {
	conn, ok := socket.connections.Load(fd)
	if !ok {
		return nil, false
	}
	return conn.(*Socket), true
}

func (socket *Server) GetConnectionsCount() uint32 {
	return socket.count
}

func (socket *Server) Start() {

	socket.Ready()

	if socket.Host != "" {
		var ip, port, err = utils.Addr.Parse(socket.Host)
		if err != nil {
			panic(err)
		}
		socket.IP = ip
		socket.Port = port
	}

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", socket.IP+":"+strconv.Itoa(socket.Port))

	if err != nil {
		if strings.HasSuffix(err.Error(), "address already in use") {
			if socket.AutoBind {
				socket.Port++
				socket.Start()
				return
			}
		}
		panic(err)
	}

	socket.netListen = netListen

	// start success
	if socket.OnSuccess != nil {
		socket.OnSuccess()
	}

	go func() {
		for {
			conn, err := netListen.Accept()
			if err != nil {
				socket.connError <- exception.New(err)
				continue
			}

			go socket.process(conn)
		}
	}()

	<-socket.shutdown

	err = netListen.Close()

	console.Exit(err)
}

func (socket *Server) Shutdown() {
	socket.shutdown <- true
}

func (socket *Server) process(conn net.Conn) {

	// 超时时间
	err := conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
	if err != nil {
		socket.connError <- exception.New(err)
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

	socket.connOpen <- connection

	var reader = socket.Protocol.Reader()

	var buffer = make([]byte, socket.ReadBufferSize)

	for {

		n, err := conn.Read(buffer)

		// close error
		if err != nil {
			break
		}

		message, err := reader(n, buffer)

		if err != nil {
			socket.connError <- exception.New(err)
			break
		}

		if message == nil {
			continue
		}

		err = socket.decodeMessage(connection, message)

		if err != nil {
			socket.connError <- exception.New(err)
			break
		}

	}

	socket.connClose <- connection
}

func (socket *Server) decodeMessage(connection *Socket, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := socket.Protocol.Decode(message)

	if socket.OnMessage != nil {
		go socket.OnMessage(connection, messageType, message)
	}

	// check version
	if version != lemo.Version {
		return nil
	}

	// Ping
	if messageType == lemo.PingData {
		return socket.PingHandler(connection)("")
	}

	// Pong
	if messageType == lemo.PongData {
		return socket.PongHandler(connection)("")
	}

	// on router
	if socket.router != nil {
		go socket.middleware(connection, &lemo.ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
		return nil
	}

	return nil
}

func (socket *Server) middleware(conn *Socket, msg *lemo.ReceivePackage) {
	var next Middle = socket.handler
	for i := len(socket.middle) - 1; i >= 0; i-- {
		next = socket.middle[i](next)
	}
	next(conn, msg)
}

func (socket *Server) handler(conn *Socket, msg *lemo.ReceivePackage) {

	var n, formatPath = socket.router.getRoute(msg.Event)
	if n == nil {
		if socket.OnError != nil {
			socket.OnError(exception.New(msg.Event + " " + "404 not found"))
		}
		return
	}

	var nodeData = n.Data.(*node)

	var receive = &lemo.Receive{}
	receive.Body = msg
	receive.Context = nil
	receive.Params = lemo.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

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
