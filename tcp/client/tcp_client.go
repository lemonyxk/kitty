/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-16 16:10
**/

package client

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/golang/protobuf/proto"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/tcp"
)

type Client struct {
	Host string
	Port int

	Conn              net.Conn
	AutoHeartBeat     bool
	HeartBeatTimeout  int
	HeartBeatInterval int
	HeartBeat         func(c *Client) error
	Reconnect         bool
	ReconnectInterval int
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  int

	// 消息处理
	OnOpen    func(c *Client)
	OnClose   func(c *Client)
	OnMessage func(c *Client, messageType int, msg []byte)
	OnError   func(err exception.Error)
	Status    bool

	Context lemo.Context

	PingHandler func(c *Client) func(appData string) error

	PongHandler func(c *Client) func(appData string) error

	Protocol tcp.Protocol

	router *Router

	middle []func(Middle) Middle

	mux sync.RWMutex
}

type Middle func(c *Client, receive *lemo.ReceivePackage)

func (client *Client) Use(middle ...func(Middle) Middle) {
	client.middle = append(client.middle, middle...)
}

// Json 发送JSON字符
func (client *Client) Emit(event []byte, body []byte, dataType int, protoType int) error {
	return client.Push(client.Protocol.Encode(event, body, dataType, protoType))
}

func (client *Client) JsonEmit(msg lemo.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return client.Push(client.Protocol.Encode([]byte(msg.Event), data, lemo.TextData, lemo.Json))
}

func (client *Client) ProtoBufEmit(msg lemo.ProtoBufPackage) error {
	data, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return client.Push(client.Protocol.Encode([]byte(msg.Event), data, lemo.BinData, lemo.ProtoBuf))
}

// Push 发送消息
func (client *Client) Push(message []byte) error {

	if client.Status == false {
		return errors.New("client is close")
	}

	client.mux.Lock()
	_, err := client.Conn.Write(message)
	client.mux.Unlock()

	return err
}

func (client *Client) Close() error {
	client.Reconnect = false
	return client.Conn.Close()
}

func (client *Client) reconnecting() {
	if client.Reconnect == true {
		time.AfterFunc(time.Duration(client.ReconnectInterval)*time.Second, func() {
			client.Connect()
		})
	}
}

func (client *Client) Connect() {

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

	if client.Protocol == nil {
		client.Protocol = &tcp.DefaultProtocol{}
	}

	// heartbeat function
	if client.HeartBeat == nil {
		client.HeartBeat = func(client *Client) error {
			return client.Push(client.Protocol.Encode(nil, nil, lemo.PingData, lemo.BinData))
		}
	}

	if client.PingHandler == nil {
		client.PingHandler = func(connection *Client) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if client.PongHandler == nil {
		client.PongHandler = func(connection *Client) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	// 连接服务器
	handler, err := net.DialTimeout("tcp", client.Host+":"+strconv.Itoa(client.Port), time.Duration(client.HandshakeTimeout)*time.Second)
	if err != nil {
		go client.OnError(exception.New(err))
		client.reconnecting()
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

		var reader = client.Protocol.Reader()

		var buffer = make([]byte, client.ReadBufferSize)

		for {

			n, err := client.Conn.Read(buffer)

			// close error
			if err != nil {
				break
			}

			message, err := reader(n, buffer)

			if err != nil {
				go client.OnError(exception.New(err))
				break
			}

			if message == nil {
				continue
			}

			err = client.decodeMessage(client, message)

			if err != nil {
				go client.OnError(exception.New(err))
				break
			}

		}

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

func (client *Client) decodeMessage(connection *Client, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := client.Protocol.Decode(message)

	if client.OnMessage != nil {
		go client.OnMessage(connection, messageType, message)
	}

	// check version
	if version != lemo.Version {
		return nil
	}

	// Ping
	if messageType == lemo.PingData {
		return client.PingHandler(connection)("")
	}

	// Pong
	if messageType == lemo.PongData {
		return client.PongHandler(connection)("")
	}

	// on router
	if client.router != nil {
		go client.middleware(connection, &lemo.ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
		return nil
	}

	return nil
}

func (client *Client) middleware(conn *Client, msg *lemo.ReceivePackage) {
	var next Middle = client.handler
	for i := len(client.middle) - 1; i >= 0; i-- {
		next = client.middle[i](next)
	}
	next(conn, msg)
}

func (client *Client) handler(conn *Client, msg *lemo.ReceivePackage) {

	var n, formatPath = client.router.getRoute(msg.Event)
	if n == nil {
		if client.OnError != nil {
			client.OnError(exception.New(msg.Event + " " + "404 not found"))
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
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
		receive.Context = ctx
	}

	err := nodeData.Function(conn, receive)
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

func (client *Client) SetRouter(router *Router) *Client {
	client.router = router
	return client
}

func (client *Client) GetRouter() *Router {
	return client.router
}
