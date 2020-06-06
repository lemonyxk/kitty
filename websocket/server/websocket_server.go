package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	websocket2 "github.com/Lemo-yxk/lemo/websocket"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// WebSocket WebSocket
type WebSocket struct {
	FD       uint32
	Conn     *websocket.Conn
	Server   *Server
	Response http.ResponseWriter
	Request  *http.Request
	Context  lemo.Context
}

func (conn *WebSocket) Host() string {
	if host := conn.Request.Header.Get(lemo.Host); host != "" {
		return host
	}
	return conn.Request.Host
}

func (conn *WebSocket) ClientIP() string {

	if ip := strings.Split(conn.Request.Header.Get(lemo.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := conn.Request.Header.Get(lemo.XRealIP); ip != "" {
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

func (conn *WebSocket) Emit(event []byte, body []byte, dataType int, protoType int) error {
	return conn.Server.Emit(conn.FD, event, body, dataType, protoType)
}

func (conn *WebSocket) Json(msg lemo.JsonPackage) exception.Error {
	return conn.Server.Json(conn.FD, msg)
}

func (conn *WebSocket) JsonEmit(msg lemo.JsonPackage) error {
	return conn.Server.JsonEmit(conn.FD, msg)
}

func (conn *WebSocket) ProtoBufEmit(msg lemo.ProtoBufPackage) error {
	return conn.Server.ProtoBufEmit(conn.FD, msg)
}

func (conn *WebSocket) Close() error {
	return conn.Conn.Close()
}

type Server struct {

	// Host 服务Host
	Host string
	// Port 服务端口
	Port int
	// Protocol 协议
	TLS bool
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	AutoBind bool

	OnClose   func(conn *WebSocket)
	OnMessage func(conn *WebSocket, messageType int, msg []byte)
	OnOpen    func(conn *WebSocket)
	OnError   func(err exception.Error)

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

	Protocol websocket2.Protocol

	// 连接
	connOpen chan *WebSocket

	// 关闭
	connClose chan *WebSocket

	// 写入
	connPush chan *lemo.PushPackage

	// 返回
	connBack chan error

	// 错误
	connError chan exception.Error

	upgrade websocket.Upgrader

	fd          uint32
	count       uint32
	connections sync.Map
	router      *Router
	middle      []func(next Middle) Middle
	server      *http.Server
	netListen   net.Listener
}

type Middle func(conn *WebSocket, receive *lemo.ReceivePackage)

func (socket *Server) Use(middle ...func(next Middle) Middle) {
	socket.middle = append(socket.middle, middle...)
}

func (socket *Server) CheckPath(p1 string, p2 string) bool {
	return p1 == p2
}

// Push 发送消息
func (socket *Server) Push(fd uint32, messageType int, msg []byte) error {

	socket.connPush <- &lemo.PushPackage{
		Type: messageType,
		FD:   fd,
		Data: msg,
	}

	return <-socket.connBack
}

func (socket *Server) Json(fd uint32, msg lemo.JsonPackage) exception.Error {
	data, err := jsoniter.Marshal(lemo.JsonPackage{Event: msg.Event, Data: msg.Data})
	if err != nil {
		return exception.New(err)
	}
	return exception.New(socket.Push(fd, lemo.TextData, data))
}

func (socket *Server) Emit(fd uint32, event []byte, body []byte, dataType int, protoType int) error {
	return socket.Push(fd, lemo.BinData, socket.Protocol.Encode(event, body, dataType, protoType))
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

func (socket *Server) JsonAll(msg lemo.JsonPackage) (int, int) {
	var counter = 0
	var success = 0
	socket.connections.Range(func(key, value interface{}) bool {
		counter++
		if socket.Json(key.(uint32), msg) == nil {
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

	messageProtoBuf, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}

	return socket.Push(fd, lemo.BinData, socket.Protocol.Encode([]byte(msg.Event), messageProtoBuf, lemo.BinData, lemo.ProtoBuf))

}

func (socket *Server) JsonEmit(fd uint32, msg lemo.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return socket.Push(fd, lemo.TextData, socket.Protocol.Encode([]byte(msg.Event), data, lemo.TextData, lemo.Json))
}

func (socket *Server) addConnect(conn *WebSocket) {

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

func (socket *Server) delConnect(conn *WebSocket) {
	socket.connections.Delete(conn.FD)
}

func (socket *Server) GetConnections() chan *WebSocket {
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

func (socket *Server) GetConnection(fd uint32) (*WebSocket, bool) {
	conn, ok := socket.connections.Load(fd)
	if !ok {
		return nil, false
	}
	return conn.(*WebSocket), true
}

func (socket *Server) GetConnectionsCount() uint32 {
	return socket.count
}

func (socket *Server) Close(fd uint32) error {
	conn, ok := socket.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (socket *Server) Ready() {

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
		socket.OnError = func(err exception.Error) {
			console.Error(err)
		}
	}

	if socket.Protocol == nil {
		socket.Protocol = &websocket2.DefaultProtocol{}
	}

	if socket.PingHandler == nil {
		socket.PingHandler = func(connection *WebSocket) func(appData string) error {
			return func(appData string) error {
				// unnecessary
				// err := Server.Push(connection.FD, BinData, socket.Protocol.Encode(nil, nil, PongData, BinData))
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
					socket.connBack <- conn.(*WebSocket).Conn.WriteMessage(push.Type, push.Data)
				}
			case err := <-socket.connError:
				go socket.OnError(err)
			}
		}
	}()
}

func (socket *Server) process(w http.ResponseWriter, r *http.Request) {

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

func (socket *Server) decodeMessage(connection *WebSocket, message []byte, messageFrame int) error {

	// unpack
	version, messageType, protoType, route, body := socket.Protocol.Decode(message)

	if socket.OnMessage != nil {
		go socket.OnMessage(connection, messageFrame, message)
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

func (socket *Server) middleware(conn *WebSocket, msg *lemo.ReceivePackage) {
	var next Middle = socket.handler
	for i := len(socket.middle) - 1; i >= 0; i-- {
		next = socket.middle[i](next)
	}
	next(conn, msg)
}

func (socket *Server) handler(conn *WebSocket, msg *lemo.ReceivePackage) {

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

	err := nodeData.ServerFunction(conn, receive)
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

// Start Http
func (socket *Server) Start() {

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

	if socket.TLS {
		err = server.ServeTLS(netListen, socket.CertFile, socket.KeyFile)
	} else {
		err = server.Serve(netListen)
	}

	console.Exit(err)
}

func (socket *Server) Shutdown() {
	err := socket.server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}

func (socket *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Match the websocket router
	if r.Method != http.MethodGet || !socket.CheckPath(r.URL.Path, socket.Path) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	socket.process(w, r)
	return
}
