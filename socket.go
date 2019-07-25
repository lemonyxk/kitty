package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Fte struct {
	Fd    uint32
	Type  int
	Event string
}

type FteMessage struct {
	Fte
	Msg []byte
}

type M map[string]interface{}

type WebSocketServerFunction func(conn *Connection, fte *Fte, msg []byte)

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
	Socket   *Socket
	Response http.ResponseWriter
	Request  *http.Request
}

// Socket conn
type Socket struct {
	Fd          uint32
	Connections map[uint32]*Connection
	OnClose     func(conn *Connection)
	OnMessage   func(conn *Connection, fte *Fte, msg []byte)
	OnOpen      func(conn *Connection)
	OnError     func(err error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	CheckOrigin       func(r *http.Request) bool

	Before func(conn *Connection, fte *Fte, msg []byte) error
	After  func(conn *Connection, fte *Fte, msg []byte) error

	WebSocketRouter map[string]WebSocketServerFunction

	TsProto int

	mux sync.RWMutex
}

func (conn *Connection) IP() (string, string, error) {

	if ip := conn.Request.Header.Get("X-Real-IP"); ip != "" {
		return net.SplitHostPort(ip)
	}

	if ip := conn.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return net.SplitHostPort(ip)
	}

	return net.SplitHostPort(conn.Request.RemoteAddr)
}

func (conn *Connection) Emit(fte *Fte, msg interface{}) error {
	return conn.Socket.Emit(fte, msg)
}

func (conn *Connection) EmitAll(fte *Fte, msg interface{}) {
	conn.Socket.EmitAll(fte, msg)
}

// Push 发送消息
func (socket *Socket) Push(fd uint32, messageType int, msg []byte) error {

	socket.mux.Lock()
	defer socket.mux.Unlock()

	conn, ok := socket.Connections[fd]

	if !ok {
		return fmt.Errorf("client %d is close", fd)
	}

	// 默认为文本
	if messageType == 0 {
		messageType = TextMessage
	}

	return conn.Conn.WriteMessage(messageType, msg)
}

// Push Json 发送消息
func (socket *Socket) Json(fte *Fte, msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return socket.Push(fte.Fd, fte.Type, messageJson)
}

func (socket *Socket) ProtoBuf(fte *Fte, msg interface{}) error {
	return nil
}

func (socket *Socket) EmitAll(fte *Fte, msg interface{}) {

	socket.mux.RLock()
	var fds []uint32
	for fd := range socket.Connections {
		fds = append(fds, fd)
	}
	socket.mux.RUnlock()

	for _, fd := range fds {
		fte.Fd = fd
		_ = socket.Emit(fte, msg)
	}
}

func (socket *Socket) Emit(fte *Fte, msg interface{}) error {

	if fte.Type == BinaryMessage {
		if j, b := msg.([]byte); b {
			return socket.Push(fte.Fd, fte.Type, j)
		}

		return fmt.Errorf("message type is bin that message must be []byte")
	}

	switch socket.TsProto {
	case Json:
		return socket.jsonEmit(fte.Fd, fte.Type, fte.Event, msg)
	case ProtoBuf:
		return socket.protoBufEmit(fte.Fd, fte.Type, fte.Event, msg)
	}

	return fmt.Errorf("unknown ts ptoto")

}

func (socket *Socket) protoBufEmit(fd uint32, messageType int, event string, msg interface{}) error {
	return nil
}

func (socket *Socket) jsonEmit(fd uint32, messageType int, event string, msg interface{}) error {

	var messageJson = M{"event": event, "data": msg}

	if j, b := msg.([]byte); b {
		messageJson["data"] = string(j)
	}

	return socket.Json(&Fte{Fd: fd, Type: messageType, Event: event}, messageJson)

}

func (socket *Socket) addConnect(conn *Connection) {

	socket.mux.Lock()

	// +1
	socket.Fd++

	// 溢出
	if socket.Fd == 0 {
		socket.Fd++
	}

	// 如果不存在 则存储
	if _, ok := socket.Connections[socket.Fd]; !ok {
		socket.Connections[socket.Fd] = conn
	} else {

		// 否则查找最大值
		var maxFd uint32 = 0

		for {

			maxFd++

			if maxFd == 0 {
				log.Println("connections overflow")
				socket.mux.Unlock()
				return
			}

			if _, ok := socket.Connections[maxFd]; !ok {
				socket.Connections[maxFd] = conn
				break
			}

		}

		socket.Fd = maxFd

	}

	// 赋值
	conn.Fd = socket.Fd

	socket.mux.Unlock()

	// 触发OPEN事件
	socket.OnOpen(conn)
}
func (socket *Socket) delConnect(conn *Connection) {
	socket.mux.Lock()

	delete(socket.Connections, conn.Fd)

	socket.mux.Unlock()

	socket.OnClose(conn)
}

// WebSocket 默认设置
func WebSocket(socket *Socket) http.HandlerFunc {

	if socket.TsProto == 0 {
		socket.TsProto = Json
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

	if socket.ReadBufferSize == 0 {
		socket.ReadBufferSize = 2 * 1024 * 1024
	}

	if socket.WriteBufferSize == 0 {
		socket.WriteBufferSize = 2 * 1024 * 1024
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
			log.Println(conn.Fd, "is open at", time.Now())
		}
	}

	if socket.OnClose == nil {
		socket.OnClose = func(conn *Connection) {
			log.Println(conn.Fd, "is close at", time.Now())
		}
	}

	if socket.OnError == nil {
		socket.OnError = func(err error) {
			log.Println(err)
		}
	}

	upgrade := websocket.Upgrader{
		HandshakeTimeout: time.Duration(socket.HandshakeTimeout) * time.Second,
		ReadBufferSize:   socket.ReadBufferSize,
		WriteBufferSize:  socket.WriteBufferSize,
		CheckOrigin:      socket.CheckOrigin,
	}

	socket.Connections = make(map[uint32]*Connection)

	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {

		// 升级协议
		conn, err := upgrade.Upgrade(w, r, nil)

		// 错误处理
		if err != nil {
			socket.OnError(err)
			return
		}

		// 设置PING处理函数
		conn.SetPingHandler(func(status string) error {
			return conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
		})

		connection := Connection{
			Conn:     conn,
			Socket:   socket,
			Response: w,
			Request:  r,
		}

		// 打开连接 记录
		socket.addConnect(&connection)

		// 关闭连接 清理
		defer func() {
			_ = conn.Close()
			socket.delConnect(&connection)
		}()

		// 收到消息 处理 单一连接接受不冲突 但是不能并发写入
		for {

			// 重置心跳
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
			messageType, message, err := conn.ReadMessage()

			// 关闭连接
			if err != nil {
				// log.Println(err)
				break
			}

			go func() {
				// 处理消息
				if socket.Before != nil {
					if err := socket.Before(&connection, &Fte{Fd: connection.Fd, Type: messageType}, message); err != nil {
						return
					}
				}

				if socket.OnMessage != nil {
					socket.OnMessage(&connection, &Fte{Fd: connection.Fd, Type: messageType}, message)
				}

				if socket.WebSocketRouter != nil {
					socket.router(&connection, &Fte{Fd: connection.Fd, Type: messageType}, message)
				}

				if socket.After != nil {
					_ = socket.After(&connection, &Fte{Fd: connection.Fd, Type: messageType}, message)
				}
			}()

		}

	}

	return handler
}
