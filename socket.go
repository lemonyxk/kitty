package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Fd          uint32
	MessageType int
	Message     []byte
}

type WebSocketFunction func(conn *Connection, messageType int, message []byte, context interface{})

// Connection Connection
type Connection struct {
	Fd       uint32
	Socket   *websocket.Conn
	Handler  *Socket
	Response http.ResponseWriter
	Request  *http.Request
	push     chan *Message
	back     chan error
}

// Socket conn
type Socket struct {
	Fd          uint32
	Connections map[uint32]*Connection
	OnClose     func(conn *Connection)
	OnMessage   func(conn *Connection, message *Message)
	OnOpen      func(conn *Connection)
	OnError     func(err error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int
	CheckOrigin       func(r *http.Request) bool

	BeforeSend func() error

	// PingMessage PING
	PingMessage int
	// PongMessage PONG
	PongMessage int
	// TextMessage 文本
	TextMessage int
	// BinaryMessage 二进制
	BinaryMessage int

	WebSocketRouter map[string]WebSocketFunction
}

func (socket *Connection) IP() (string, string, error) {

	if ip := socket.Request.Header.Get("X-Real-IP"); ip != "" {
		return net.SplitHostPort(ip)
	}

	if ip := socket.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return net.SplitHostPort(ip)
	}

	return net.SplitHostPort(socket.Request.RemoteAddr)
}

func (socket *Socket) InitRouter() {
	socket.WebSocketRouter = make(map[string]WebSocketFunction)
}

func (socket *Socket) WSetRoute(route string, f WebSocketFunction) {
	socket.WebSocketRouter[route] = f
}

func (socket *Socket) WGetRouter(route string) WebSocketFunction {
	if f, ok := socket.WebSocketRouter[route]; ok {
		return f
	}
	return nil
}

// Push 发送消息
func (socket *Socket) Push(fd uint32, messageType int, message []byte) error {

	if _, ok := socket.Connections[fd]; !ok {
		return fmt.Errorf("client %d is close", fd)
	}

	socket.Connections[fd].push <- &Message{fd, messageType, message}

	return <-socket.Connections[fd].back
}

func (socket *Socket) AddConnect(conn *Connection) {

	// +1
	socket.Fd++

	// 赋值
	conn.Fd = socket.Fd

	// 如果不存在 则存储
	if _, ok := socket.Connections[conn.Fd]; !ok {
		socket.Connections[conn.Fd] = conn
	} else {

		// 否则查找最大值
		var maxFd uint32 = 0

		for fd, _ := range socket.Connections {
			if fd > maxFd {
				maxFd = fd
			}
		}

		// +1
		maxFd++

		// 溢出
		if maxFd == 0 {
			maxFd++
		}

		socket.Connections[maxFd] = conn

	}

	// 触发OPEN事件
	socket.OnOpen(conn)
}
func (socket *Socket) DelConnect(conn *Connection) {
	delete(socket.Connections, conn.Fd)
	socket.OnClose(conn)
}

// WebSocket 默认设置
func WebSocket(socket *Socket) http.HandlerFunc {

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

	if socket.OnMessage == nil {
		socket.OnMessage = func(conn *Connection, message *Message) {
			log.Println(*message)
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

	socket.PingMessage = websocket.PingMessage

	socket.PongMessage = websocket.PongMessage

	socket.TextMessage = websocket.TextMessage

	socket.BinaryMessage = websocket.BinaryMessage

	upgrade := websocket.Upgrader{
		HandshakeTimeout: time.Duration(socket.HandshakeTimeout) * time.Second,
		ReadBufferSize:   socket.ReadBufferSize,
		WriteBufferSize:  socket.WriteBufferSize,
		CheckOrigin:      socket.CheckOrigin,
	}

	socket.Connections = make(map[uint32]*Connection)

	// 连接
	var connOpen = make(chan *Connection, socket.WaitQueueSize)

	// 关闭
	var connClose = make(chan *Connection, socket.WaitQueueSize)

	go func() {
		for {
			select {
			case conn := <-connOpen:
				socket.AddConnect(conn)
			case conn := <-connClose:
				socket.DelConnect(conn)
			}
		}
	}()

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
			Fd:       0,
			Socket:   conn,
			Handler:  socket,
			Response: w,
			Request:  r,
			push:     make(chan *Message, 1024),
			back:     make(chan error, 1024),
		}

		// 打开连接 记录
		connOpen <- &connection

		// 关闭连接 清理
		defer func() {
			conn.Close()
			connClose <- &connection
		}()

		go func() {
			for {
				select {
				case message := <-connection.push:
					connection.back <- socket.Connections[message.Fd].Socket.WriteMessage(message.MessageType, message.Message)
				}
			}
		}()

		// 收到消息 处理 单一连接收发不冲突 但是不能并发写入
		for {

			// 重置心跳
			conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
			messageType, message, err := conn.ReadMessage()

			// 关闭连接
			if err != nil {
				//log.Println(err)
				break
			}

			// 处理消息
			go socket.OnMessage(&connection, &Message{Fd: connection.Fd, MessageType: messageType, Message: message})
		}

	}

	return handler
}
