package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/lemoyxk/kitty"
	websocket2 "github.com/lemoyxk/kitty/websocket"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type WebSocket struct {
	FD       int64
	Conn     *websocket.Conn
	Server   *Server
	Response http.ResponseWriter
	Request  *http.Request
	Context  kitty.Context
	mux      sync.Mutex
}

func (conn *WebSocket) Host() string {
	if host := conn.Request.Header.Get(kitty.Host); host != "" {
		return host
	}
	return conn.Request.Host
}

func (conn *WebSocket) ClientIP() string {

	if ip := strings.Split(conn.Request.Header.Get(kitty.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := conn.Request.Header.Get(kitty.XRealIP); ip != "" {
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

func (conn *WebSocket) Json(msg kitty.JsonPackage) error {
	return conn.Server.Json(conn.FD, msg)
}

func (conn *WebSocket) JsonEmit(msg kitty.JsonPackage) error {
	return conn.Server.JsonEmit(conn.FD, msg)
}

func (conn *WebSocket) ProtoBufEmit(msg kitty.ProtoBufPackage) error {
	return conn.Server.ProtoBufEmit(conn.FD, msg)
}

func (conn *WebSocket) Close() error {
	return conn.Conn.Close()
}

type Server struct {

	// Host 服务Host
	Name string
	Host string
	// Protocol 协议
	TLS bool
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	OnClose   func(conn *WebSocket)
	OnMessage func(conn *WebSocket, messageType int, msg []byte)
	OnOpen    func(conn *WebSocket)
	OnError   func(err error)
	OnSuccess func()

	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HandshakeTimeout  time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	CheckOrigin       func(r *http.Request) bool
	Path              string

	PingHandler func(connection *WebSocket) func(appData string) error

	PongHandler func(connection *WebSocket) func(appData string) error

	Protocol websocket2.Protocol

	upgrade websocket.Upgrader

	fd          int64
	connections map[int64]*WebSocket
	mux         sync.RWMutex
	router      *Router
	middle      []func(next Middle) Middle
	server      *http.Server
	netListen   net.Listener
}

type Middle func(conn *WebSocket, receive *kitty.ReceivePackage)

func (socket *Server) LocalAddr() net.Addr {
	return socket.netListen.Addr()
}

func (socket *Server) Use(middle ...func(next Middle) Middle) {
	socket.middle = append(socket.middle, middle...)
}

func (socket *Server) CheckPath(p1 string, p2 string) bool {
	return p1 == p2
}

func (socket *Server) Push(fd int64, messageType int, msg []byte) error {
	var conn, ok = socket.GetConnection(fd)
	if !ok {
		return errors.New("client is close")
	}

	conn.mux.Lock()
	defer conn.mux.Unlock()

	return conn.Conn.WriteMessage(messageType, msg)
}

func (socket *Server) Json(fd int64, msg kitty.JsonPackage) error {
	data, err := jsoniter.Marshal(kitty.JsonPackage{Event: msg.Event, Data: msg.Data})
	if err != nil {
		return err
	}
	return socket.Push(fd, kitty.TextData, data)
}

func (socket *Server) Emit(fd int64, event []byte, body []byte, dataType int, protoType int) error {
	return socket.Push(fd, kitty.BinData, socket.Protocol.Encode(event, body, dataType, protoType))
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

func (socket *Server) JsonAll(msg kitty.JsonPackage) (int, int) {
	var counter = 0
	var success = 0
	for fd := range socket.connections {
		counter++
		if socket.Json(fd, msg) == nil {
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
	messageProtoBuf, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return socket.Push(fd, kitty.BinData, socket.Protocol.Encode([]byte(msg.Event), messageProtoBuf, kitty.BinData, kitty.ProtoBuf))
}

func (socket *Server) JsonEmit(fd int64, msg kitty.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return socket.Push(fd, kitty.TextData, socket.Protocol.Encode([]byte(msg.Event), data, kitty.TextData, kitty.Json))
}

func (socket *Server) addConnect(conn *WebSocket) {
	socket.mux.Lock()
	defer socket.mux.Unlock()
	socket.fd++
	socket.connections[socket.fd] = conn
	conn.FD = socket.fd
}

func (socket *Server) delConnect(conn *WebSocket) {
	socket.mux.Lock()
	defer socket.mux.Unlock()
	delete(socket.connections, conn.FD)
}

func (socket *Server) GetConnections() chan *WebSocket {
	var ch = make(chan *WebSocket, 1)
	go func() {
		for _, conn := range socket.connections {
			ch <- conn
		}
		close(ch)
	}()
	return ch
}

func (socket *Server) GetConnection(fd int64) (*WebSocket, bool) {
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

func (socket *Server) Close(fd int64) error {
	conn, ok := socket.GetConnection(fd)
	if !ok {
		return errors.New("fd not found")
	}
	return conn.Close()
}

func (socket *Server) onOpen(conn *WebSocket) {
	socket.addConnect(conn)
	socket.OnOpen(conn)
}

func (socket *Server) onClose(conn *WebSocket) {
	_ = conn.Conn.Close()
	socket.delConnect(conn)
	socket.OnClose(conn)
}

func (socket *Server) onError(err error) {
	socket.OnError(err)
}

func (socket *Server) Ready() {

	if socket.Path == "" {
		socket.Path = "/"
	}

	if socket.Host == "" {
		panic("Host must set")
	}

	if socket.HeartBeatTimeout == 0 {
		socket.HeartBeatTimeout = 30 * time.Second
	}

	if socket.HeartBeatInterval == 0 {
		socket.HeartBeatInterval = 15 * time.Second
	}

	if socket.HandshakeTimeout == 0 {
		socket.HandshakeTimeout = 2 * time.Second
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
			println(conn.FD, "is open")
		}
	}

	if socket.OnClose == nil {
		socket.OnClose = func(conn *WebSocket) {
			println(conn.FD, "is close")
		}
	}

	if socket.OnError == nil {
		socket.OnError = func(err error) {
			println(err)
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
				return connection.Conn.SetReadDeadline(time.Now().Add(socket.HeartBeatTimeout))
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

	socket.upgrade = websocket.Upgrader{
		HandshakeTimeout: socket.HandshakeTimeout,
		ReadBufferSize:   socket.ReadBufferSize,
		WriteBufferSize:  socket.WriteBufferSize,
		CheckOrigin:      socket.CheckOrigin,
	}

	socket.connections = make(map[int64]*WebSocket)
}

func (socket *Server) process(w http.ResponseWriter, r *http.Request) {

	// 升级协议
	conn, err := socket.upgrade.Upgrade(w, r, nil)

	// 错误处理
	if err != nil {
		socket.onError(err)
		return
	}

	// 超时时间
	err = conn.SetReadDeadline(time.Now().Add(socket.HeartBeatTimeout))
	if err != nil {
		socket.onError(err)
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
	socket.onOpen(connection)

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
			_ = conn.SetReadDeadline(time.Now().Add(socket.HeartBeatTimeout))
		}

		err = socket.decodeMessage(connection, message, messageFrame)
		if err != nil {
			socket.onError(err)
			break
		}

	}

	// close and clean
	socket.onClose(connection)

}

func (socket *Server) decodeMessage(connection *WebSocket, message []byte, messageFrame int) error {

	// unpack
	version, messageType, protoType, route, body := socket.Protocol.Decode(message)

	if socket.OnMessage != nil {
		socket.OnMessage(connection, messageFrame, message)
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

func (socket *Server) middleware(conn *WebSocket, msg *kitty.ReceivePackage) {
	var next Middle = socket.handler
	for i := len(socket.middle) - 1; i >= 0; i-- {
		next = socket.middle[i](next)
	}
	next(conn, msg)
}

func (socket *Server) handler(conn *WebSocket, msg *kitty.ReceivePackage) {

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

func (socket *Server) Start() {

	socket.Ready()

	var server = http.Server{Addr: socket.Host, Handler: socket}

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", server.Addr)

	if err != nil {
		panic(err)
	}

	socket.netListen = netListen
	socket.server = &server

	// start success
	if socket.OnSuccess != nil {
		socket.OnSuccess()
	}

	if socket.TLS {
		err = server.ServeTLS(netListen, socket.CertFile, socket.KeyFile)
	} else {
		err = server.Serve(netListen)
	}

	if err != nil {
		println(err)
	}
}

func (socket *Server) Shutdown() error {
	return socket.server.Shutdown(context.Background())
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
