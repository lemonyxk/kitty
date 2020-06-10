package client

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
	websocket2 "github.com/Lemo-yxk/lemo/websocket"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Client struct {
	// 服务器信息
	Name   string
	Scheme string
	Host   string
	IP     string
	Port   int
	Path   string
	// Origin   http.Header

	// 客户端信息
	Conn              *websocket.Conn
	Response          *http.Response
	AutoHeartBeat     bool
	HeartBeatTimeout  int
	HeartBeatInterval int
	HeartBeat         func(c *Client) error
	Reconnect         bool
	ReconnectInterval int
	WriteBufferSize   int
	ReadBufferSize    int
	HandshakeTimeout  int

	// 消息处理
	OnOpen    func(c *Client)
	OnClose   func(c *Client)
	OnMessage func(c *Client, messageType int, msg []byte)
	OnError   func(err exception.Error)
	OnSuccess func()

	Context lemo.Context

	PingHandler func(c *Client) func(appData string) error

	PongHandler func(c *Client) func(appData string) error

	Protocol websocket2.Protocol

	mux    sync.RWMutex
	router *Router
	middle []func(Middle) Middle
}

type Middle func(c *Client, receive *lemo.ReceivePackage)

func (client *Client) LocalAddr() net.Addr {
	return client.Conn.LocalAddr()
}

func (client *Client) RemoteAddr() net.Addr {
	return client.Conn.RemoteAddr()
}

func (client *Client) Use(middle ...func(Middle) Middle) {
	client.middle = append(client.middle, middle...)
}

func (client *Client) Emit(event []byte, body []byte, dataType int, protoType int) exception.Error {
	return client.Push(dataType, client.Protocol.Encode(event, body, dataType, protoType))
}

func (client *Client) Json(msg lemo.JsonPackage) exception.Error {
	messageJson, err := jsoniter.Marshal(lemo.JsonPackage{Event: msg.Event, Data: msg.Data})
	if err != nil {
		return exception.New(err)
	}
	return exception.New(client.Push(lemo.TextData, messageJson))
}

func (client *Client) JsonEmit(msg lemo.JsonPackage) exception.Error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return exception.New(err)
	}
	return client.Push(lemo.TextData, client.Protocol.Encode([]byte(msg.Event), data, lemo.TextData, lemo.Json))
}

func (client *Client) ProtoBufEmit(msg lemo.ProtoBufPackage) exception.Error {
	messageProtoBuf, err := proto.Marshal(msg.Data)
	if err != nil {
		return exception.New(err)
	}
	return client.Push(lemo.BinData, client.Protocol.Encode([]byte(msg.Event), messageProtoBuf, lemo.BinData, lemo.ProtoBuf))
}

// Push 发送消息
func (client *Client) Push(messageType int, message []byte) exception.Error {
	client.mux.Lock()
	defer client.mux.Unlock()
	return exception.New(client.Conn.WriteMessage(messageType, message))
}

func (client *Client) Close() error {
	return client.Conn.Close()
}

func (client *Client) reconnecting() {
	if client.Reconnect == true {
		time.AfterFunc(time.Duration(client.ReconnectInterval)*time.Second, func() {
			client.Connect()
		})
	}
}

// Connect 连接服务器
func (client *Client) Connect() {

	if client.Host == "" {
		client.Host = "127.0.0.1"
	}

	if client.Port == 0 {
		client.Port = 1207
	}

	if client.Path == "" {
		client.Path = "/"
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

	// 写入BUF大小
	if client.WriteBufferSize == 0 {
		client.WriteBufferSize = 1024
	}

	// 读出BUF大小
	if client.ReadBufferSize == 0 {
		client.ReadBufferSize = 1024
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
		client.Protocol = &websocket2.DefaultProtocol{}
	}

	// heartbeat function
	if client.HeartBeat == nil {
		client.HeartBeat = func(client *Client) error {
			return client.Push(lemo.BinData, client.Protocol.Encode(nil, nil, lemo.PingData, lemo.BinData))
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

	var dialer = websocket.Dialer{
		HandshakeTimeout: time.Duration(client.HandshakeTimeout) * time.Second,
		WriteBufferSize:  client.WriteBufferSize,
		ReadBufferSize:   client.ReadBufferSize,
	}

	if client.Host != "" {
		var ip, port, err = utils.Addr.Parse(client.Host)
		if err != nil {
			panic(err)
		}
		client.IP = ip
		client.Port = port
	}

	// 连接服务器
	handler, response, err := dialer.Dial(client.Scheme+"://"+client.IP+":"+strconv.Itoa(client.Port)+client.Path, nil)
	if err != nil {
		client.OnError(exception.New(err))
		client.reconnecting()
		return
	}

	// 设置PING处理函数
	handler.SetPingHandler(client.PingHandler(client))

	// 设置PONG处理函数
	handler.SetPongHandler(client.PongHandler(client))

	client.Response = response

	client.Conn = handler

	// start success
	if client.OnSuccess != nil {
		client.OnSuccess()
	}

	// 连接成功
	client.OnOpen(client)

	// 定时器 心跳
	ticker := time.NewTicker(time.Duration(client.HeartBeatInterval) * time.Second)

	// 如果有心跳设置
	if client.AutoHeartBeat != true {
		ticker.Stop()
	}

	go func() {
		for range ticker.C {
			if err := client.HeartBeat(client); err != nil {
				client.OnError(exception.New(err))
				_ = client.Close()
				break
			}
		}
	}()

	for {
		messageFrame, message, err := client.Conn.ReadMessage()
		// close error
		if err != nil {
			break
		}

		err = client.decodeMessage(client, messageFrame, message)

		if err != nil {
			client.OnError(exception.New(err))
			break
		}
	}

	// 关闭定时器
	ticker.Stop()
	// 关闭连接
	_ = client.Close()
	// 触发回调
	client.OnClose(client)
	// 触发重连设置
	client.reconnecting()
}

func (client *Client) decodeMessage(conn *Client, messageFrame int, message []byte) error {
	// unpack
	version, messageType, protoType, route, body := client.Protocol.Decode(message)

	if client.OnMessage != nil {
		client.OnMessage(client, messageFrame, message)
	}

	// check version
	if version != lemo.Version {
		return nil
	}

	// Ping
	if messageType == lemo.PingData {
		return client.PingHandler(client)("")
	}

	// Pong
	if messageType == lemo.PongData {
		return client.PongHandler(client)("")
	}

	// on router
	if client.router != nil {
		client.middleware(client, &lemo.ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
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
