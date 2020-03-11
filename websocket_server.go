package lemo

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/protocol"

	"github.com/json-iterator/go"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// WebSocket WebSocket
type WebSocket struct {
	FD       uint32
	Conn     *websocket.Conn
	Server   *WebSocketServer
	Response http.ResponseWriter
	Request  *http.Request
	Context  interface{}
}

func (conn *WebSocket) Host() string {
	if host := conn.Request.Header.Get(Host); host != "" {
		return host
	}
	return conn.Request.Host
}

func (conn *WebSocket) ClientIP() string {

	if ip := strings.Split(conn.Request.Header.Get(XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := conn.Request.Header.Get(XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(conn.Request.RemoteAddr); err == nil {
		return ip
	}

	return ""
}

func (conn *WebSocket) Push(messageType int, msg []byte) error {
	return conn.Server.Push(conn.FD, messageType, msg)
}

func (conn *WebSocket) JsonFormat(msg JsonPackage) exception.ErrorFunc {
	return conn.Server.JsonFormat(conn.FD, msg)
}

func (conn *WebSocket) Json(msg interface{}) error {
	return conn.Server.Json(conn.FD, msg)
}

func (conn *WebSocket) ProtoBuf(msg proto.Message) error {
	return conn.Server.ProtoBuf(conn.FD, msg)
}

func (conn *WebSocket) JsonEmit(msg JsonPackage) error {
	return conn.Server.JsonEmit(conn.FD, msg)
}

func (conn *WebSocket) ProtoBufEmit(msg ProtoBufPackage) error {
	return conn.Server.ProtoBufEmit(conn.FD, msg)
}

func (conn *WebSocket) Close() error {
	return conn.Conn.Close()
}

// WebSocketServer conn
type WebSocketServer struct {

	// Host 服务Host
	Host string
	// Port 服务端口
	Port int
	// Protocol 协议
	Protocol string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	AutoBind bool

	OnClose   func(conn *WebSocket)
	OnMessage func(conn *WebSocket, messageType int, msg []byte)
	OnOpen    func(conn *WebSocket)
	OnError   func(err exception.ErrorFunc)

	HeartBeatTimeout  int
	HeartBeatInterval int
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	CheckOrigin       func(r *http.Request) bool
	Path              string

	PingHandler func(connection *WebSocket) func(appData string) error

	PongHandler func(connection *WebSocket) func(appData string) error

	// 连接
	connOpen chan *WebSocket

	// 关闭
	connClose chan *WebSocket

	// 写入
	connPush chan *PushInfo

	// 返回
	connBack chan error

	// 错误
	connError chan exception.ErrorFunc

	upgrade websocket.Upgrader

	fd          uint32
	count       uint32
	connections sync.Map
	router      *WebSocketServerRouter
	middle      []func(next WebSocketServerMiddle) WebSocketServerMiddle
	server      *http.Server
	netListen   net.Listener
}

type WebSocketServerMiddle func(conn *WebSocket, receive *ReceivePackage)

func (socket *WebSocketServer) Use(middle ...func(next WebSocketServerMiddle) WebSocketServerMiddle) {
	socket.middle = append(socket.middle, middle...)
}

func (socket *WebSocketServer) CheckPath(p1 string, p2 string) bool {
	return p1 == p2
}

// Push 发送消息
func (socket *WebSocketServer) Push(fd uint32, messageType int, msg []byte) error {

	socket.connPush <- &PushInfo{
		MessageType: messageType,
		FD:          fd,
		Message:     msg,
	}

	return <-socket.connBack
}

func (socket *WebSocketServer) JsonFormat(fd uint32, msg JsonPackage) exception.ErrorFunc {
	messageJsonFormat, err := jsoniter.Marshal(JsonMessage{msg.Event, msg.Message})
	if err != nil {
		return exception.New(err)
	}
	return exception.New(socket.Push(fd, protocol.TextData, messageJsonFormat))
}

// Push Json 发送消息
func (socket *WebSocketServer) Json(fd uint32, msg interface{}) error {

	messageJson, err := jsoniter.Marshal(msg)
	if err != nil {
		return err
	}

	return socket.Push(fd, protocol.TextData, messageJson)
}

func (socket *WebSocketServer) ProtoBuf(fd uint32, msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return socket.Push(fd, protocol.BinData, messageProtoBuf)
}

func (socket *WebSocketServer) JsonFormatAll(msg JsonPackage) (int, int) {
	var counter = 0
	var success = 0
	socket.connections.Range(func(key, value interface{}) bool {
		counter++
		if socket.JsonFormat(key.(uint32), msg) == nil {
			success++
		}
		return true
	})
	return counter, success
}

func (socket *WebSocketServer) JsonEmitAll(msg JsonPackage) (int, int) {
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

func (socket *WebSocketServer) ProtoBufEmitAll(msg ProtoBufPackage) (int, int) {
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

func (socket *WebSocketServer) ProtoBufEmit(fd uint32, msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return err
	}

	return socket.Push(fd, protocol.BinData, protocol.Pack([]byte(msg.Event), messageProtoBuf, protocol.BinData, protocol.ProtoBuf))

}

func (socket *WebSocketServer) JsonEmit(fd uint32, msg JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Message)
	if err != nil {
		return err
	}
	return socket.Push(fd, protocol.TextData, protocol.Pack([]byte(msg.Event), data, protocol.TextData, protocol.Json))
}

func (socket *WebSocketServer) addConnect(conn *WebSocket) {

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

func (socket *WebSocketServer) delConnect(conn *WebSocket) {
	socket.connections.Delete(conn.FD)
}

func (socket *WebSocketServer) GetConnections() chan *WebSocket {
	var ch = make(chan *WebSocket, 1)
	go func() {
		socket.connections.Range(func(key, value interface{}) bool {
			ch <- value.(*WebSocket)
			return true
		})
		close(ch)
	}()
	return ch
}

func (socket *WebSocketServer) GetConnection(fd uint32) (*WebSocket, bool) {
	conn, ok := socket.connections.Load(fd)
	if !ok {
		return nil, false
	}
	return conn.(*WebSocket), true
}

func (socket *WebSocketServer) GetConnectionsCount() uint32 {
	return socket.count
}

func (socket *WebSocketServer) Close(fd uint32) error {
	conn, ok := socket.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (socket *WebSocketServer) Ready() {

	if socket.HeartBeatTimeout == 0 {
		socket.HeartBeatTimeout = 30
	}

	if socket.HeartBeatInterval == 0 {
		socket.HeartBeatInterval = 15
	}

	if socket.HandshakeTimeout == 0 {
		socket.HandshakeTimeout = 2
	}

	// must be 4096 or the memory will leak
	if socket.ReadBufferSize == 0 {
		socket.ReadBufferSize = 1024
	}
	// must be 4096 or the memory will leak
	if socket.WriteBufferSize == 0 {
		socket.WriteBufferSize = 1024
	}

	if socket.WaitQueueSize == 0 {
		socket.WaitQueueSize = 1024
	}

	if socket.CheckOrigin == nil {
		socket.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	if socket.OnOpen == nil {
		socket.OnOpen = func(conn *WebSocket) {
			console.Println(conn.FD, "is open")
		}
	}

	if socket.OnClose == nil {
		socket.OnClose = func(conn *WebSocket) {
			console.Println(conn.FD, "is close")
		}
	}

	if socket.OnError == nil {
		socket.OnError = func(err exception.ErrorFunc) {
			console.Error(err)
		}
	}

	if socket.PingHandler == nil {
		socket.PingHandler = func(connection *WebSocket) func(appData string) error {
			return func(appData string) error {
				// unnecessary
				// err := Server.Push(connection.FD, BinData, Pack(nil, nil, PongData, BinData))
				return connection.Conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
			}
		}
	}

	if socket.PongHandler == nil {
		socket.PongHandler = func(connection *WebSocket) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if socket.Path == "" {
		socket.Path = "/"
	}

	socket.upgrade = websocket.Upgrader{
		HandshakeTimeout: time.Duration(socket.HandshakeTimeout) * time.Second,
		ReadBufferSize:   socket.ReadBufferSize,
		WriteBufferSize:  socket.WriteBufferSize,
		CheckOrigin:      socket.CheckOrigin,
	}

	// 连接
	socket.connOpen = make(chan *WebSocket, socket.WaitQueueSize)

	// 关闭
	socket.connClose = make(chan *WebSocket, socket.WaitQueueSize)

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
					socket.connBack <- conn.(*WebSocket).Conn.WriteMessage(push.MessageType, push.Message)
				}
			case err := <-socket.connError:
				go socket.OnError(err)
			}
		}
	}()

}

func (socket *WebSocketServer) process(w http.ResponseWriter, r *http.Request) {

	// 升级协议
	conn, err := socket.upgrade.Upgrade(w, r, nil)

	// 错误处理
	if err != nil {
		socket.connError <- exception.New(err)
		return
	}

	// 超时时间
	err = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
	if err != nil {
		socket.connError <- exception.New(err)
		return
	}

	connection := &WebSocket{
		FD:       0,
		Conn:     conn,
		Server:   socket,
		Response: w,
		Request:  r,
	}

	// 设置PING处理函数
	conn.SetPingHandler(socket.PingHandler(connection))

	// 设置PONG处理函数
	conn.SetPongHandler(socket.PongHandler(connection))

	// 打开连接 记录
	socket.connOpen <- connection

	// 收到消息 处理 单一连接接受不冲突 但是不能并发写入
	for {

		// read message
		messageFrame, message, err := conn.ReadMessage()
		// close
		if err != nil {
			break
		}

		// do not let it dead
		// for web ping
		if len(message) == 0 {
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
		}

		err = socket.decodeMessage(connection, message, messageFrame)
		if err != nil {
			socket.connError <- exception.New(err)
			break
		}

	}

	// close and clean
	socket.connClose <- connection

}

func (socket *WebSocketServer) decodeMessage(connection *WebSocket, message []byte, messageFrame int) error {

	// unpack
	version, messageType, protoType, route, body := protocol.UnPack(message)

	if socket.OnMessage != nil {
		go socket.OnMessage(connection, messageFrame, message)
	}

	// check version
	if version != protocol.Version {
		route, body := protocol.ParseMessage(message)
		if route != nil {
			go socket.middleware(connection, &ReceivePackage{MessageType: messageFrame, Event: string(route), Message: body, ProtoType: protocol.Json, Raw: message})
		}
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

func (socket *WebSocketServer) middleware(conn *WebSocket, msg *ReceivePackage) {
	var next WebSocketServerMiddle = socket.handler
	for i := len(socket.middle) - 1; i >= 0; i-- {
		next = socket.middle[i](next)
	}
	next(conn, msg)
}

func (socket *WebSocketServer) handler(conn *WebSocket, msg *ReceivePackage) {

	var node, formatPath = socket.router.getRoute(msg.Event)
	if node == nil {
		return
	}

	var nodeData = node.Data.(*WebSocketServerNode)

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

	err := nodeData.WebSocketServerFunction(conn, receive)
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

func (socket *WebSocketServer) SetRouter(router *WebSocketServerRouter) *WebSocketServer {
	socket.router = router
	return socket
}

func (socket *WebSocketServer) GetRouter() *WebSocketServerRouter {
	return socket.router
}

// Start Http
func (socket *WebSocketServer) Start() {

	socket.Ready()

	var server = http.Server{Addr: socket.Host + ":" + strconv.Itoa(socket.Port), Handler: socket}

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", server.Addr)

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
	socket.server = &server

	switch socket.Protocol {
	case "TLS":
		err = server.ServeTLS(netListen, socket.CertFile, socket.KeyFile)
	default:
		err = server.Serve(netListen)
	}

	console.Exit(err)
}

func (socket *WebSocketServer) Shutdown() {
	err := socket.server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}

func (socket *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Match the websocket router
	if r.Method != http.MethodGet || !socket.CheckPath(r.URL.Path, socket.Path) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	socket.process(w, r)
	return
}
