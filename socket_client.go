/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-16 16:10
**/

package lemo

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/protocol"

	"github.com/golang/protobuf/proto"
)

type SocketClient struct {
	Host string
	Port int

	Conn              net.Conn
	AutoHeartBeat     bool
	HeartBeatTimeout  int
	HeartBeatInterval int
	HeartBeat         func(c *SocketClient) error
	Reconnect         bool
	ReconnectInterval int
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  int

	// 消息处理
	OnOpen    func(c *SocketClient)
	OnClose   func(c *SocketClient)
	OnMessage func(c *SocketClient, messageType int, msg []byte)
	OnError   func(err exception.Error)
	Status    bool

	Context Context

	PingHandler func(c *SocketClient) func(appData string) error

	PongHandler func(c *SocketClient) func(appData string) error

	router *SocketClientRouter

	middle []func(SocketClientMiddle) SocketClientMiddle

	mux sync.RWMutex
}

type SocketClientMiddle func(c *SocketClient, receive *ReceivePackage)

func (client *SocketClient) Use(middle ...func(SocketClientMiddle) SocketClientMiddle) {
	client.middle = append(client.middle, middle...)
}

// Json 发送JSON字符
func (client *SocketClient) Json(msg interface{}) error {

	messageJson, err := jsoniter.Marshal(msg)
	if err != nil {
		return err
	}

	return client.Push(messageJson)
}

func (client *SocketClient) ProtoBuf(msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return client.Push(messageProtoBuf)

}

func (client *SocketClient) JsonEmit(msg JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Message)
	if err != nil {
		return err
	}
	return client.Push(protocol.Pack([]byte(msg.Event), data, protocol.TextData, protocol.Json))

}

func (client *SocketClient) ProtoBufEmit(msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return err
	}

	return client.Push(protocol.Pack([]byte(msg.Event), messageProtoBuf, protocol.BinData, protocol.ProtoBuf))

}

// Push 发送消息
func (client *SocketClient) Push(message []byte) error {

	if client.Status == false {
		return errors.New("client is close")
	}

	client.mux.Lock()
	_, err := client.Conn.Write(message)
	client.mux.Unlock()

	return err
}

func (client *SocketClient) Close() error {
	client.Reconnect = false
	return client.Conn.Close()
}

func (client *SocketClient) reconnecting() {
	if client.Reconnect == true {
		time.AfterFunc(time.Duration(client.ReconnectInterval)*time.Second, func() {
			client.Connect()
		})
	}
}

func (client *SocketClient) Connect() {

	// 设置LOG信息

	var closeChan = make(chan bool)

	if client.Host == "" {
		client.Host = "127.0.0.1"
	}

	if client.Port == 0 {
		client.Port = 1207
	}

	if client.OnOpen == nil {
		panic("OnOpen must set")
	}

	if client.OnClose == nil {
		panic("OnClose must set")
	}

	if client.OnError == nil {
		panic("OnError must set")
	}

	// 握手
	if client.HandshakeTimeout == 0 {
		client.HandshakeTimeout = 2
	}

	// 读出BUF大小
	if client.ReadBufferSize == 0 {
		client.ReadBufferSize = 1024
	}

	// 读出BUF大小
	if client.WriteBufferSize == 0 {
		client.WriteBufferSize = 1024
	}

	// 定时心跳间隔
	if client.HeartBeatInterval == 0 {
		client.HeartBeatInterval = 15
	}

	if client.HeartBeatTimeout == 0 {
		client.HeartBeatTimeout = 30
	}

	// 自动重连间隔
	if client.ReconnectInterval == 0 {
		client.ReconnectInterval = 1
	}

	// heartbeat function
	if client.HeartBeat == nil {
		client.HeartBeat = func(client *SocketClient) error {
			return client.Push(protocol.Pack(nil, nil, protocol.PingData, protocol.BinData))
		}
	}

	if client.PingHandler == nil {
		client.PingHandler = func(connection *SocketClient) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if client.PongHandler == nil {
		client.PongHandler = func(connection *SocketClient) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	// 连接服务器
	handler, err := net.DialTimeout("tcp", client.Host+":"+strconv.Itoa(client.Port), time.Duration(client.HandshakeTimeout)*time.Second)
	if err != nil {
		go client.OnError(exception.New(err))
		return
	}

	err = handler.(*net.TCPConn).SetReadBuffer(client.ReadBufferSize)
	if err != nil {
		panic(err)
	}

	err = handler.(*net.TCPConn).SetWriteBuffer(client.WriteBufferSize)
	if err != nil {
		panic(err)
	}

	client.Conn = handler

	client.Status = true

	// 连接成功
	go client.OnOpen(client)

	// 定时器 心跳
	ticker := time.NewTicker(time.Duration(client.HeartBeatInterval) * time.Second)

	// 如果有心跳设置
	if client.AutoHeartBeat != true {
		ticker.Stop()
	}

	go func() {
		for range ticker.C {
			if err := client.HeartBeat(client); err != nil {
				closeChan <- false
				break
			}
		}
	}()

	go func() {

		var singleMessageLen = 0

		var message []byte

		var buffer = make([]byte, client.ReadBufferSize)

		for {

			n, err := client.Conn.Read(buffer)

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
						go client.OnError(exception.New("invalid header"))
						goto OUT
					}

					singleMessageLen = protocol.GetLen(message)
				}

				// jump out and read continue
				if len(message) < singleMessageLen {
					break
				}

				// a complete message
				err := client.decodeMessage(client, message[0:singleMessageLen])
				if err != nil {
					go client.OnError(exception.New(err))
					goto OUT
				}

				// delete this message
				message = message[singleMessageLen:]

				// reset len
				singleMessageLen = 0

			}

		}

	OUT:
		closeChan <- false
	}()

	<-closeChan

	// 关闭定时器
	ticker.Stop()
	// 更改状态
	client.Status = false
	// 关闭连接
	_ = client.Close()
	// 触发回调
	go client.OnClose(client)
	// 触发重连设置
	client.reconnecting()
}

func (client *SocketClient) decodeMessage(connection *SocketClient, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := protocol.UnPack(message)

	if client.OnMessage != nil {
		go client.OnMessage(connection, messageType, message)
	}

	// check version
	if version != protocol.Version {
		return nil
	}

	// Ping
	if messageType == protocol.PingData {
		return client.PingHandler(connection)("")
	}

	// Pong
	if messageType == protocol.PongData {
		return client.PongHandler(connection)("")
	}

	// on router
	if client.router != nil {
		go client.middleware(connection, &ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
		return nil
	}

	return nil
}

func (client *SocketClient) middleware(conn *SocketClient, msg *ReceivePackage) {
	var next SocketClientMiddle = client.handler
	for i := len(client.middle) - 1; i >= 0; i-- {
		next = client.middle[i](next)
	}
	next(conn, msg)
}

func (client *SocketClient) handler(conn *SocketClient, msg *ReceivePackage) {

	var node, formatPath = client.router.getRoute(msg.Event)
	if node == nil {
		if client.OnError != nil {
			client.OnError(exception.New(msg.Event + " " + "404 not found"))
		}
		return
	}

	var nodeData = node.Data.(*SocketClientNode)

	var receive = &Receive{}
	receive.Body = msg
	receive.Context = nil
	receive.Params = Params{Keys: node.Keys, Values: node.ParseParams(formatPath)}

	for i := 0; i < len(nodeData.Before); i++ {
		ctx, err := nodeData.Before[i](conn, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
		receive.Context = ctx
	}

	err := nodeData.SocketClientFunction(conn, receive)
	if err != nil {
		if client.OnError != nil {
			client.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](conn, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
	}

}

func (client *SocketClient) SetRouter(router *SocketClientRouter) *SocketClient {
	client.router = router
	return client
}

func (client *SocketClient) GetRouter() *SocketClientRouter {
	return client.router
}
