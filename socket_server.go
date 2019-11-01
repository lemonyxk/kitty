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
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/Lemo-yxk/tire"
	"github.com/golang/protobuf/proto"
)

type Socket struct {
	Fd     uint32
	Conn   net.Conn
	socket *SocketServer
}

func (conn *Socket) ClientIP() (string, string, error) {
	return net.SplitHostPort(conn.Conn.RemoteAddr().String())
}

func (conn *Socket) Push(fd uint32, msg []byte) error {
	return conn.socket.Push(fd, msg)
}

func (conn *Socket) Json(fd uint32, msg interface{}) error {
	return conn.socket.Json(fd, msg)
}

func (conn *Socket) ProtoBuf(fd uint32, msg proto.Message) error {
	return conn.socket.ProtoBuf(fd, msg)
}

func (conn *Socket) JsonEmit(fd uint32, msg JsonPackage) error {
	return conn.socket.JsonEmit(fd, msg)
}

func (conn *Socket) ProtoBufEmit(fd uint32, msg ProtoBufPackage) error {
	return conn.socket.ProtoBufEmit(fd, msg)
}

func (conn *Socket) JsonEmitAll(msg JsonPackage) {
	conn.socket.JsonEmitAll(msg)
}

func (conn *Socket) ProtoBufEmitAll(msg ProtoBufPackage) {
	conn.socket.ProtoBufEmitAll(msg)
}

func (conn *Socket) GetConnections() chan *Socket {
	return conn.socket.GetConnections()
}

func (conn *Socket) GetSocket() *SocketServer {
	return conn.socket
}

func (conn *Socket) GetConnectionsCount() uint32 {
	return conn.socket.GetConnectionsCount()
}

func (conn *Socket) GetConnection(fd uint32) (*Socket, bool) {
	return conn.socket.GetConnection(fd)
}

type SocketServer struct {
	Host      string
	Port      int
	OnClose   func(fd uint32)
	OnMessage func(conn *Socket, messageType int, msg []byte)
	OnOpen    func(conn *Socket)
	OnError   func(err func() *Error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	ReadBufferSize    int
	WaitQueueSize     int
	HandshakeTimeout  int

	tire *tire.Tire

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
	group       *socketServerGroup
	route       *socketServerRoute
}

// Push 发送消息
func (socket *SocketServer) Push(fd uint32, msg []byte) error {

	socket.connPush <- &PushPackage{
		FD:      fd,
		Message: msg,
	}

	return <-socket.connBack
}

// Push Json 发送消息
func (socket *SocketServer) Json(fd uint32, msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return errors.New("message err: " + err.Error())
	}

	return socket.Push(fd, messageJson)
}

func (socket *SocketServer) ProtoBuf(fd uint32, msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return errors.New("protobuf err: " + err.Error())
	}

	return socket.Push(fd, messageProtoBuf)
}

func (socket *SocketServer) JsonEmitAll(msg JsonPackage) {
	socket.connections.Range(func(key, value interface{}) bool {
		_ = socket.JsonEmit(key.(uint32), msg)
		return true
	})
}

func (socket *SocketServer) ProtoBufEmitAll(msg ProtoBufPackage) {
	socket.connections.Range(func(key, value interface{}) bool {
		_ = socket.ProtoBufEmit(key.(uint32), msg)
		return true
	})
}

func (socket *SocketServer) ProtoBufEmit(fd uint32, msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return errors.New("protobuf err: " + err.Error())
	}

	return socket.Push(fd, Pack([]byte(msg.Event), messageProtoBuf, BinData, ProtoBuf))

}

func (socket *SocketServer) JsonEmit(fd uint32, msg JsonPackage) error {

	var data []byte

	if mb, ok := msg.Message.([]byte); ok {
		data = mb
	} else {
		messageJson, err := json.Marshal(msg.Message)
		if err != nil {
			return errors.New("protobuf err: " + err.Error())
		}
		data = messageJson
	}

	return socket.Push(fd, Pack([]byte(msg.Event), data, TextData, Json))

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
			println(err().Error)
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

	netListen, err := net.Listen("tcp", socket.Host+":"+strconv.Itoa(socket.Port))
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
			err := socket.decodeMessage(connection, message[0:singleMessageLen])
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

func (socket *SocketServer) decodeMessage(connection *Socket, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := UnPack(message)

	// check version
	if version != Version {
		if socket.OnMessage != nil {
			go socket.OnMessage(connection, messageType, message)
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
	if socket.tire != nil {
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
