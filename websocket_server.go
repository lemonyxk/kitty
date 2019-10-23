package lemo

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lemo-yxk/tire"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Receive struct {
	Context Context
	Params  *Params
	Message *ReceivePackage
}

type ReceivePackage struct {
	MessageType int
	Event       []byte
	Message     []byte
	ProtoType   int
}

type JsonPackage struct {
	Event   string
	Message interface{}
}

type ProtoBufPackage struct {
	Event   string
	Message proto.Message
}

type PushPackage struct {
	MessageType int
	FD          uint32
	Message     []byte
}

type M map[string]interface{}

// WebSocket WebSocket
type WebSocket struct {
	Fd       uint32
	Conn     *websocket.Conn
	socket   *WebSocketServer
	Response http.ResponseWriter
	Request  *http.Request
}

// WebSocketServer conn
type WebSocketServer struct {
	OnClose   func(fd uint32)
	OnMessage func(conn *WebSocket, messageType int, msg []byte)
	OnOpen    func(conn *WebSocket)
	OnError   func(err func() *Error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	CheckOrigin       func(r *http.Request) bool
	Path              string

	Router *tire.Tire

	IgnoreCase bool

	// 连接
	connOpen chan *WebSocket

	// 关闭
	connClose chan *WebSocket

	// 写入
	connPush chan *PushPackage

	// 返回
	connBack chan error

	// 错误
	connError chan func() *Error

	upgrade websocket.Upgrader

	PingHandler func(connection *WebSocket) func(appData string) error

	PongHandler func(connection *WebSocket) func(appData string) error

	fd          uint32
	count       uint32
	connections sync.Map
	group       *webSocketServerGroup
	route       *webSocketServerRoute
}

func (socket *WebSocketServer) CheckPath(p1 string, p2 string) bool {
	if socket.IgnoreCase {
		p1 = strings.ToLower(p1)
		p2 = strings.ToLower(p2)
	}
	return p1 == p2
}

func (conn *WebSocket) ClientIP() (string, string, error) {

	if ip := conn.Request.Header.Get(XRealIP); ip != "" {
		return net.SplitHostPort(ip)
	}

	if ip := conn.Request.Header.Get(XForwardedFor); ip != "" {
		return net.SplitHostPort(ip)
	}

	return net.SplitHostPort(conn.Request.RemoteAddr)
}

func (conn *WebSocket) Push(fd uint32, messageType int, msg []byte) error {
	return conn.socket.Push(fd, messageType, msg)
}

func (conn *WebSocket) Json(fd uint32, msg interface{}) error {
	return conn.socket.Json(fd, msg)
}

func (conn *WebSocket) JsonFormat(fd uint32, msg JsonPackage) error {
	return conn.socket.JsonFormat(fd, msg)
}

func (conn *WebSocket) ProtoBuf(fd uint32, msg proto.Message) error {
	return conn.socket.ProtoBuf(fd, msg)
}

func (conn *WebSocket) JsonEmit(fd uint32, msg JsonPackage) error {
	return conn.socket.JsonEmit(fd, msg)
}

func (conn *WebSocket) ProtoBufEmit(fd uint32, msg ProtoBufPackage) error {
	return conn.socket.ProtoBufEmit(fd, msg)
}

func (conn *WebSocket) JsonEmitAll(msg JsonPackage) {
	conn.socket.JsonEmitAll(msg)
}

func (conn *WebSocket) ProtoBufEmitAll(msg ProtoBufPackage) {
	conn.socket.ProtoBufEmitAll(msg)
}

func (conn *WebSocket) GetConnections() chan *WebSocket {
	return conn.socket.GetConnections()
}

func (conn *WebSocket) GetSocket() *WebSocketServer {
	return conn.socket
}

func (conn *WebSocket) GetConnectionsCount() uint32 {
	return conn.socket.GetConnectionsCount()
}

func (conn *WebSocket) GetConnection(fd uint32) (*WebSocket, bool) {
	return conn.socket.GetConnection(fd)
}

// Push 发送消息
func (socket *WebSocketServer) Push(fd uint32, messageType int, msg []byte) error {

	socket.connPush <- &PushPackage{
		MessageType: messageType,
		FD:          fd,
		Message:     msg,
	}

	return <-socket.connBack
}

func (socket *WebSocketServer) JsonFormat(fd uint32, msg JsonPackage) error {

	messageJsonFormat, err := json.Marshal(M{"data": msg.Message, "event": msg.Event})
	if err != nil {
		return errors.New("message error: " + err.Error())
	}

	return socket.Push(fd, TextData, messageJsonFormat)
}

// Push Json 发送消息
func (socket *WebSocketServer) Json(fd uint32, msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return errors.New("message error: " + err.Error())
	}

	return socket.Push(fd, TextData, messageJson)
}

func (socket *WebSocketServer) ProtoBuf(fd uint32, msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return errors.New("protobuf error: " + err.Error())
	}

	return socket.Push(fd, BinData, messageProtoBuf)
}

func (socket *WebSocketServer) JsonFormatAll(msg JsonPackage) {
	socket.connections.Range(func(key, value interface{}) bool {
		_ = socket.JsonFormat(key.(uint32), msg)
		return true
	})
}

func (socket *WebSocketServer) JsonEmitAll(msg JsonPackage) {
	socket.connections.Range(func(key, value interface{}) bool {
		_ = socket.JsonEmit(key.(uint32), msg)
		return true
	})
}

func (socket *WebSocketServer) ProtoBufEmitAll(msg ProtoBufPackage) {
	socket.connections.Range(func(key, value interface{}) bool {
		_ = socket.ProtoBufEmit(key.(uint32), msg)
		return true
	})
}

func (socket *WebSocketServer) ProtoBufEmit(fd uint32, msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return errors.New("protobuf error: " + err.Error())
	}

	return socket.Push(fd, BinData, Pack([]byte(msg.Event), messageProtoBuf, BinData, ProtoBuf))

}

func (socket *WebSocketServer) JsonEmit(fd uint32, msg JsonPackage) error {

	var data []byte

	if mb, ok := msg.Message.([]byte); ok {
		data = mb
	} else {
		messageJson, err := json.Marshal(msg.Message)
		if err != nil {
			return errors.New("protobuf error: " + err.Error())
		}
		data = messageJson
	}

	return socket.Push(fd, TextData, Pack([]byte(msg.Event), data, TextData, Json))

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
				println("connections overflow")
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
	conn.Fd = socket.fd

}

func (socket *WebSocketServer) delConnect(conn *WebSocket) {
	socket.connections.Delete(conn.Fd)
}

func (socket *WebSocketServer) GetConnections() chan *WebSocket {

	var ch = make(chan *WebSocket, 1024)

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
		socket.ReadBufferSize = 4096
	}
	// must be 4096 or the memory will leak
	if socket.WriteBufferSize == 0 {
		socket.WriteBufferSize = 4096
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
			println(conn.Fd, "is open")
		}
	}

	if socket.OnClose == nil {
		socket.OnClose = func(fd uint32) {
			println(fd, "is close")
		}
	}

	if socket.OnError == nil {
		socket.OnError = func(err func() *Error) {
			println(err())
		}
	}

	if socket.PingHandler == nil {
		socket.PingHandler = func(connection *WebSocket) func(appData string) error {
			return func(appData string) error {
				// unnecessary
				// err := socket.Push(connection.Fd, BinData, Pack(nil, nil, PongData, BinData))
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
	socket.connPush = make(chan *PushPackage, socket.WaitQueueSize)

	// 返回
	socket.connBack = make(chan error, socket.WaitQueueSize)

	// 错误
	socket.connError = make(chan func() *Error, socket.WaitQueueSize)

	go func() {
		for {
			select {
			case conn := <-socket.connOpen:
				socket.addConnect(conn)
				socket.count++
				// 触发OPEN事件
				go socket.OnOpen(conn)
			case conn := <-socket.connClose:
				var fd = conn.Fd
				_ = conn.Conn.Close()
				socket.delConnect(conn)
				socket.count--
				// 触发CLOSE事件
				go socket.OnClose(fd)
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

func (socket *WebSocketServer) handler(w http.ResponseWriter, r *http.Request) {

	// 升级协议
	conn, err := socket.upgrade.Upgrade(w, r, nil)

	// 错误处理
	if err != nil {
		socket.connError <- NewError(err)
		return
	}

	// 超时时间
	err = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
	if err != nil {
		socket.connError <- NewError(err)
		return
	}

	connection := &WebSocket{
		Fd:       0,
		Conn:     conn,
		socket:   socket,
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
		// for web
		if len(message) == 0 {
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
		}

		err = socket.decodeMessage(connection, message, messageFrame)
		if err != nil {
			socket.connError <- NewError(err)
			break
		}

	}

	// close and clean
	socket.connClose <- connection

}

func (socket *WebSocketServer) decodeMessage(connection *WebSocket, message []byte, messageFrame int) error {

	// unpack
	version, messageType, protoType, route, body := UnPack(message)

	// check version
	if version != Version {

		route, body := ParseMessage(message)

		if route != nil {
			if socket.Router != nil {
				var receivePackage = &ReceivePackage{MessageType: messageFrame, Event: route, Message: body, ProtoType: Json}
				go socket.router(connection, receivePackage)
				return nil
			}
		}

		if socket.OnMessage != nil {
			go socket.OnMessage(connection, messageFrame, message)
		}

		return nil
	}

	// Ping
	if messageType == PingData {
		err := socket.PingHandler(connection)("")
		if err != nil {
			return err
		}
		return nil
	}

	// Pong
	if messageType == PongData {
		err := socket.PongHandler(connection)("")
		if err != nil {
			return err
		}
		return nil
	}

	// // check message type
	// if frameType != messageType {
	// 	return errors.New("frame type not match message type")
	// }

	// on router
	if socket.Router != nil {
		var receivePackage = &ReceivePackage{MessageType: messageType, Event: route, Message: body, ProtoType: protoType}
		go socket.router(connection, receivePackage)
		return nil
	}

	// any way run check on message
	if socket.OnMessage != nil {
		go socket.OnMessage(connection, messageFrame, message)
	}

	return nil
}
