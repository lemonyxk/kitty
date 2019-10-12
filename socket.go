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
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Lemo-yxk/tire"
)

type Socket struct {
	Fd     uint32
	Conn   net.Conn
	socket *SocketServer
}

type SocketServer struct {
	Host      string
	Port      uint16
	OnClose   func(fd uint32)
	OnMessage func(conn *Socket, messageType int, msg []byte)
	OnOpen    func(conn *Socket)
	OnError   func(err func() *Error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	ReadBufferSize    int
	WaitQueueSize     int

	Router *tire.Tire

	IgnoreCase bool

	PingHandler func(connection *Socket) func(appData string) error

	PongHandler func(connection *Socket) func(appData string) error

	// 连接
	connOpen chan *Socket

	// 关闭
	connClose chan *Socket

	// 写入
	connPush chan *PushPackage

	// 返回
	connBack chan error

	// 错误
	connError chan func() *Error

	fd          uint32
	count       uint32
	connections sync.Map
	group       *Group
	route       *Route
}

func (socket *SocketServer) Ready() {

	if socket.HeartBeatTimeout == 0 {
		socket.HeartBeatTimeout = 30
	}

	if socket.HeartBeatInterval == 0 {
		socket.HeartBeatInterval = 15
	}

	if socket.ReadBufferSize == 0 {
		socket.ReadBufferSize = 1024
	}

	if socket.WaitQueueSize == 0 {
		socket.WaitQueueSize = 1024
	}

	if socket.OnOpen == nil {
		socket.OnOpen = func(conn *Socket) {
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
			fmt.Println(err())
		}
	}

	if socket.PingHandler == nil {
		socket.PingHandler = func(connection *Socket) func(appData string) error {
			return func(appData string) error {
				// unnecessary
				// err := socket.Push(connection.Fd, PongMessage, nil)
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

	// 连接
	socket.connOpen = make(chan *Socket, socket.WaitQueueSize)

	// 关闭
	socket.connClose = make(chan *Socket, socket.WaitQueueSize)

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
					socket.connBack <- fmt.Errorf("client %d is close", push.FD)
				} else {
					socket.connBack <- conn.(*WebSocket).Conn.WriteMessage(push.MessageType, push.Message)
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

func (socket *SocketServer) delConnect(conn *Socket) {
	socket.connections.Delete(conn.Fd)
}

func (socket *SocketServer) GetConnections() chan *Socket {

	var ch = make(chan *Socket, 1024)

	go func() {
		socket.connections.Range(func(key, value interface{}) bool {
			ch <- value.(*Socket)
			return true
		})
		close(ch)
	}()

	return ch
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

	netListen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", socket.Host, socket.Port))
	if err != nil {
		panic(err)
	}

	defer func() { _ = netListen.Close() }()

	for {
		conn, err := netListen.Accept()
		if err != nil {
			socket.connError <- NewError(err)
			continue
		}

		go socket.handler(conn)
	}
}

func (socket *SocketServer) handler(conn net.Conn) {

	// 超时时间
	err := conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
	if err != nil {
		socket.connError <- NewError(err)
		return
	}

	var connection = &Socket{
		Fd:     0,
		Conn:   conn,
		socket: socket,
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

		// do not let it dead
		// but i want too
		// 超时时间
		// err = conn.SetReadDeadline(time.Now().Add(time.Duration(socket.HeartBeatTimeout) * time.Second))
		// if err != nil {
		// 	socket.connError <- err
		// 	goto OUT
		// }

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
				if !IsHeaderInvalid(message) {
					socket.connError <- NewError("invalid header")
					goto OUT
				}

				singleMessageLen = GetLen(message)
			}

			// jump out and read continue
			if len(message) < singleMessageLen {
				break
			}

			// a complete message
			err := socket.decodeMessage(connection, 2, message[0:singleMessageLen])
			if err != nil {
				socket.connError <- NewError(err)
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

func (socket *SocketServer) decodeMessage(connection *Socket, frameType int, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := UnPack(message)

	// check version
	if version != Version {
		if socket.OnMessage != nil {
			go socket.OnMessage(connection, messageType, message)
		}
		return nil
	}

	// check message type
	if frameType != messageType {
		return errors.New("frame type not match message type")
	}

	// Ping
	if messageType == PingMessage {
		err := socket.PingHandler(connection)("")
		if err != nil {
			return err
		}
		return nil
	}

	// Pong
	if messageType == PongMessage {
		err := socket.PongHandler(connection)("")
		if err != nil {
			return err
		}
		return nil
	}

	// on router
	if socket.Router != nil {
		var receivePackage = &ReceivePackage{MessageType: messageType, Event: route, Message: body, ProtoType: protoType}
		go socket.router(connection, receivePackage)
		return nil
	}

	// any way run check on message
	if socket.OnMessage != nil {
		go socket.OnMessage(connection, messageType, message)
	}
	return nil
}
