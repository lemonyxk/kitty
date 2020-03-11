/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-09 14:06
**/

package lemo

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/protocol"

	"github.com/json-iterator/go"

	"github.com/golang/protobuf/proto"
)

type Socket struct {
	FD      uint32
	Conn    net.Conn
	Server  *SocketServer
	Context Context
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

func (conn *Socket) Json(msg interface{}) error {
	return conn.Server.Json(conn.FD, msg)
}

func (conn *Socket) ProtoBuf(msg proto.Message) error {
	return conn.Server.ProtoBuf(conn.FD, msg)
}

func (conn *Socket) JsonEmit(msg JsonPackage) error {
	return conn.Server.JsonEmit(conn.FD, msg)
}

func (conn *Socket) ProtoBufEmit(msg ProtoBufPackage) error {
	return conn.Server.ProtoBufEmit(conn.FD, msg)
}

func (conn *Socket) Close() error {
	return conn.Conn.Close()
}

type SocketServer struct {
	Host      string
	Port      int
	AutoBind  bool
	OnClose   func(conn *Socket)
	OnMessage func(conn *Socket, messageType int, msg []byte)
	OnOpen    func(conn *Socket)
	OnError   func(err exception.ErrorFunc)

	HeartBeatTimeout  int
	HeartBeatInterval int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	HandshakeTimeout  int

	PingHandler func(connection *Socket) func(appData string) error

	PongHandler func(connection *Socket) func(appData string) error

	// 连接
	connOpen chan *Socket

	// 关闭
	connClose chan *Socket

	// 写入
	connPush chan *PushInfo

	// 返回
	connBack chan error

	// 错误
	connError chan exception.ErrorFunc

	fd          uint32
	count       uint32
	connections sync.Map
	router      *SocketServerRouter
	middle      []func(SocketServerMiddle) SocketServerMiddle

	netListen net.Listener
	shutdown  chan bool
}

type SocketServerMiddle func(conn *Socket, receive *ReceivePackage)

func (socket *SocketServer) Use(middle ...func(SocketServerMiddle) SocketServerMiddle) {
	socket.middle = append(socket.middle, middle...)
}

// Push 发送消息
func (socket *SocketServer) Push(fd uint32, msg []byte) error {

	socket.connPush <- &PushInfo{
		FD:      fd,
		Message: msg,
	}

	return <-socket.connBack
}

// Push Json 发送消息
func (socket *SocketServer) Json(fd uint32, msg interface{}) error {

	messageJson, err := jsoniter.Marshal(msg)
	if err != nil {
		return err
	}

	return socket.Push(fd, messageJson)
}

func (socket *SocketServer) ProtoBuf(fd uint32, msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return socket.Push(fd, messageProtoBuf)
}

func (socket *SocketServer) JsonEmitAll(msg JsonPackage) (int, int) {
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

func (socket *SocketServer) ProtoBufEmitAll(msg ProtoBufPackage) (int, int) {
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

func (socket *SocketServer) ProtoBufEmit(fd uint32, msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return err
	}

	return socket.Push(fd, protocol.Pack([]byte(msg.Event), messageProtoBuf, protocol.BinData, protocol.ProtoBuf))

}

func (socket *SocketServer) JsonEmit(fd uint32, msg JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Message)
	if err != nil {
		return err
	}
	return socket.Push(fd, protocol.Pack([]byte(msg.Event), data, protocol.TextData, protocol.Json))
}

func (socket *SocketServer) Ready() {

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
		socket.OnError = func(err exception.ErrorFunc) {
			console.Error(err)
		}
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
	socket.connPush = make(chan *PushInfo, socket.WaitQueueSize)

	// 返回
	socket.connBack = make(chan error, socket.WaitQueueSize)

	// 错误
	socket.connError = make(chan exception.ErrorFunc, socket.WaitQueueSize)

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
					_, err := conn.(*Socket).Conn.Write(push.Message)
					socket.connBack <- err
				}
			case err := <-socket.connError:
				go socket.OnError(err)
			}
		}
	}()
}

func (socket *SocketServer) addConnect(conn *Socket) {

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

func (socket *SocketServer) delConnect(conn *Socket) {
	socket.connections.Delete(conn.FD)
}

func (socket *SocketServer) GetConnections() chan *Socket {
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

func (socket *SocketServer) Close(fd uint32) error {
	conn, ok := socket.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (socket *SocketServer) GetConnection(fd uint32) (*Socket, bool) {
	conn, ok := socket.connections.Load(fd)
	if !ok {
		return nil, false
	}
	return conn.(*Socket), true
}

func (socket *SocketServer) GetConnectionsCount() uint32 {
	return socket.count
}

func (socket *SocketServer) Start() {

	socket.Ready()

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", socket.Host+":"+strconv.Itoa(socket.Port))

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

func (socket *SocketServer) Shutdown() {
	socket.shutdown <- true
}

func (socket *SocketServer) process(conn net.Conn) {

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

	var singleMessageLen = 0

	var message []byte

	var buffer = make([]byte, socket.ReadBufferSize)

	for {

		n, err := conn.Read(buffer)

		// close error
		if err != nil {
			goto OUT
		}

		message = append(message, buffer[0:n]...)

		// read continue
		if len(message) < 8 {
			continue
		}

		for {

			// jump out and read continue
			if len(message) < 8 {
				break
			}

			// just begin
			if singleMessageLen == 0 {

				// proto error
				if !protocol.IsHeaderInvalid(message) {
					socket.connError <- exception.New("invalid header")
					goto OUT
				}

				singleMessageLen = protocol.GetLen(message)
			}

			// jump out and read continue
			if len(message) < singleMessageLen {
				break
			}

			// a complete message
			err := socket.decodeMessage(connection, message[0:singleMessageLen])
			if err != nil {
				socket.connError <- exception.New(err)
				goto OUT
			}

			// delete this message
			message = message[singleMessageLen:]

			// reset len
			singleMessageLen = 0

		}

	}

OUT:
	socket.connClose <- connection
}

func (socket *SocketServer) decodeMessage(connection *Socket, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := protocol.UnPack(message)

	if socket.OnMessage != nil {
		go socket.OnMessage(connection, messageType, message)
	}

	// check version
	if version != protocol.Version {
		return nil
	}

	// Ping
	if messageType == protocol.PingData {
		return socket.PingHandler(connection)("")
	}

	// Pong
	if messageType == protocol.PongData {
		return socket.PongHandler(connection)("")
	}

	// on router
	if socket.router != nil {
		go socket.middleware(connection, &ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
		return nil
	}

	return nil
}

func (socket *SocketServer) middleware(conn *Socket, msg *ReceivePackage) {
	var next SocketServerMiddle = socket.handler
	for i := len(socket.middle) - 1; i >= 0; i-- {
		next = socket.middle[i](next)
	}
	next(conn, msg)
}

func (socket *SocketServer) handler(conn *Socket, msg *ReceivePackage) {

	var node, formatPath = socket.router.getRoute(msg.Event)
	if node == nil {
		return
	}

	var nodeData = node.Data.(*SocketServerNode)

	var receive = &Receive{}
	receive.Body = msg
	receive.Context = nil
	receive.Params = Params{Keys: node.Keys, Values: node.ParseParams(formatPath)}

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

	err := nodeData.SocketServerFunction(conn, receive)
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

func (socket *SocketServer) SetRouter(router *SocketServerRouter) *SocketServer {
	socket.router = router
	return socket
}

func (socket *SocketServer) GetRouter() *SocketServerRouter {
	return socket.router
}
