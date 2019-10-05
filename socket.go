package lemo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Lemo-yxk/tire"
	"github.com/gorilla/websocket"
)

type Receive struct {
	Context Context
	Params  *Params
	Message *MessagePackage
}

type MessagePackage struct {
	Type  int
	Event string
	Msg   interface{}
}

type PushPackage struct {
	Type int
	FD   uint32
	Msg  []byte
}

// 连接
var connOpen chan *Connection

// 关闭
var connClose chan *Connection

// 写入
var connPush chan *PushPackage

var connBack chan error

var upgrade websocket.Upgrader

type M map[string]interface{}

// PingMessage PING
const PingMessage int = websocket.PingMessage

// PongMessage PONG
const PongMessage int = websocket.PongMessage

// TextMessage 文本
const TextMessage int = websocket.TextMessage

// BinaryMessage 二进制
const BinaryMessage int = websocket.BinaryMessage

const Json = 1
const ProtoBuf = 2

// Connection Connection
type Connection struct {
	Fd       uint32
	Conn     *websocket.Conn
	socket   *Socket
	Response http.ResponseWriter
	Request  *http.Request
}

// Socket conn
type Socket struct {
	Fd          uint32
	count       uint32
	connections sync.Map
	OnClose     func(fd uint32)
	OnMessage   func(conn *Connection, messageType int, msg []byte)
	OnOpen      func(conn *Connection)
	OnError     func(err func() *Error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	CheckOrigin       func(r *http.Request) bool
	Path              string

	Router *tire.Tire

	TransportType int
	IgnoreCase    bool
}

func (socket *Socket) CheckPath(p1 string, p2 string) bool {
	if socket.IgnoreCase {
		p1 = strings.ToLower(p1)
		p2 = strings.ToLower(p2)
	}
	return p1 == p2
}

func (conn *Connection) IP() (string, string, error) {

	if ip := conn.Request.Header.Get(XRealIP); ip != "" {
		return net.SplitHostPort(ip)
	}

	if ip := conn.Request.Header.Get(XForwardedFor); ip != "" {
		return net.SplitHostPort(ip)
	}

	return net.SplitHostPort(conn.Request.RemoteAddr)
}

func (conn *Connection) Emit(fd uint32, msg *MessagePackage) error {
	return conn.socket.Emit(fd, msg)
}

func (conn *Connection) EmitAll(msg *MessagePackage) {
	conn.socket.EmitAll(msg)
}

func (conn *Connection) GetConnections() []*Connection {
	return conn.socket.GetConnections()
}

func (conn *Connection) GetSocket() *Socket {
	return conn.socket
}

func (conn *Connection) GetConnectionsCount() uint32 {
	return conn.socket.GetConnectionsCount()
}

func (conn *Connection) GetConnection(fd uint32) (*Connection, bool) {
	return conn.socket.GetConnection(fd)
}

// Push 发送消息
func (socket *Socket) Push(fd uint32, messageType int, msg []byte) error {

	// 默认为文本
	if messageType == 0 {
		messageType = TextMessage
	}

	connPush <- &PushPackage{
		Type: messageType,
		FD:   fd,
		Msg:  msg,
	}

	return <-connBack
}

// Push Json 发送消息
func (socket *Socket) Json(fd uint32, messageType int, msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return socket.Push(fd, messageType, messageJson)
}

func (socket *Socket) ProtoBuf(fd uint32, messageType int, msg interface{}) error {
	return nil
}

func (socket *Socket) EmitAll(msg *MessagePackage) {
	socket.connections.Range(func(key, value interface{}) bool {
		_ = value.(*Connection).Emit(key.(uint32), msg)
		return true
	})
}

func (socket *Socket) Emit(fd uint32, msg *MessagePackage) error {

	switch socket.TransportType {
	case Json:
		return socket.jsonEmit(fd, msg)
	case ProtoBuf:
		return socket.protoBufEmit(fd, msg)
	}

	return fmt.Errorf("unknown Transport Type")
}

func (socket *Socket) protoBufEmit(fd uint32, msg *MessagePackage) error {
	return nil
}

func (socket *Socket) jsonEmit(fd uint32, msg *MessagePackage) error {

	var messageJson = &M{"event": msg.Event, "data": msg.Msg}

	return socket.Json(fd, msg.Type, messageJson)

}

func (socket *Socket) addConnect(conn *Connection) {

	// +1
	socket.Fd++

	// 溢出
	if socket.Fd == 0 {
		socket.Fd++
	}

	var _, ok = socket.connections.Load(socket.Fd)

	if !ok {
		socket.connections.Store(socket.Fd, conn)
	} else {
		// 否则查找最大值
		var maxFd uint32 = 0

		for {

			maxFd++

			if maxFd == 0 {
				println("connections overflow")
				return
			}

			var _, ok = socket.connections.Load(socket.Fd)

			if !ok {
				socket.connections.Store(maxFd, conn)
				break
			}

		}

		socket.Fd = maxFd
	}

	// 赋值
	conn.Fd = socket.Fd

}

func (socket *Socket) delConnect(conn *Connection) {
	socket.connections.Delete(conn.Fd)
}

func (socket *Socket) GetConnections() []*Connection {

	var connections []*Connection

	socket.connections.Range(func(key, value interface{}) bool {
		connections = append(connections, value.(*Connection))
		return true
	})

	return connections
}

func (socket *Socket) GetConnection(fd uint32) (*Connection, bool) {
	conn, ok := socket.connections.Load(fd)
	if !ok {
		return nil, false
	}
	return conn.(*Connection), true
}

func (socket *Socket) GetConnectionsCount() uint32 {
	return socket.count
}

func (socket *Socket) Init() {
	if socket.TransportType == 0 {
		socket.TransportType = Json
	}

	if socket.HeartBeatTimeout == 0 {
		socket.HeartBeatTimeout = 30
	}

	if socket.HeartBeatInterval == 0 {
		socket.HeartBeatInterval = 20
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
		socket.OnOpen = func(conn *Connection) {
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

	upgrade = websocket.Upgrader{
		HandshakeTimeout: time.Duration(socket.HandshakeTimeout) * time.Second,
		ReadBufferSize:   socket.ReadBufferSize,
		WriteBufferSize:  socket.WriteBufferSize,
		CheckOrigin:      socket.CheckOrigin,
	}

	// 连接
	connOpen = make(chan *Connection, socket.WaitQueueSize)

	// 关闭
	connClose = make(chan *Connection, socket.WaitQueueSize)

	// 写入
	connPush = make(chan *PushPackage, socket.WaitQueueSize)

	// 返回
	connBack = make(chan error, socket.WaitQueueSize)

	go func() {
		for {
			select {
			case conn := <-connOpen:
				socket.addConnect(conn)
				socket.count++
				// 触发OPEN事件
				go socket.OnOpen(conn)
			case conn := <-connClose:
				var fd = conn.Fd
				socket.delConnect(conn)
				socket.count--
				// 触发CLOSE事件
				go socket.OnClose(fd)
			case push := <-connPush:
				var conn, ok = socket.connections.Load(push.FD)
				if !ok {
					connBack <- fmt.Errorf("client %d is close", push.FD)
				} else {
					connBack <- conn.(*Connection).Conn.WriteMessage(push.Type, push.Msg)
				}
			}
		}
	}()

}

func (socket *Socket) catchError() {
	if err := recover(); err != nil {
		socket.OnError(NewErrorFromDeep(err, 2))
	}
}

func (socket *Socket) upgrade(w http.ResponseWriter, r *http.Request) {

	defer socket.catchError()

	// 升级协议
	conn, err := upgrade.Upgrade(w, r, nil)

	// 错误处理
	if err != nil {
		go socket.OnError(NewError(err))
		return
	}

	// 设置PING处理函数
	conn.SetPingHandler(func(status string) error {
		err := conn.WriteMessage(PongMessage, nil)
		err = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
		return err
	})

	connection := &Connection{
		Conn:     conn,
		socket:   socket,
		Response: w,
		Request:  r,
	}

	// 打开连接 记录
	connOpen <- connection

	// 收到消息 处理 单一连接接受不冲突 但是不能并发写入
	for {

		// 重置心跳
		err := conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
		messageType, message, err := conn.ReadMessage()

		// 关闭连接
		if err != nil {
			break
		}

		go func() {

			if socket.OnMessage != nil {
				socket.OnMessage(connection, messageType, message)
			}

			if socket.Router != nil {

				if len(message) < 12 {
					return
				}

				var event, data = ParseMessage(message)
				event = strings.Replace(event, "\\", "", -1)

				var msg = &MessagePackage{Type: messageType, Event: event, Msg: data}

				socket.router(connection, msg)
			}

		}()

	}

	// 关闭连接 清理
	_ = conn.Close()
	connClose <- connection
}
