package client

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"github.com/lemoyxk/lemo"

	websocket2 "github.com/lemoyxk/lemo/websocket"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Client struct {
	// 服务器信息
	Name   string
	Scheme string
	Host   string
	Path   string
	// Origin   http.Header

	// 客户端信息
	Conn              *websocket.Conn
	Response          *http.Response
	AutoHeartBeat     bool
	HeartBeatTimeout  time.Duration
	HeartBeatInterval time.Duration
	HeartBeat         func(c *Client) error
	Reconnect         bool
	ReconnectInterval time.Duration
	WriteBufferSize   int
	ReadBufferSize    int
	HandshakeTimeout  time.Duration

	// 消息处理
	OnOpen    func(c *Client)
	OnClose   func(c *Client)
	OnMessage func(c *Client, messageType int, msg []byte)
	OnError   func(err error)
	OnSuccess func()

	Context kitty.Context

	PingHandler func(c *Client) func(appData string) error

	PongHandler func(c *Client) func(appData string) error

	Protocol websocket2.Protocol

	mux    sync.RWMutex
	router *Router
	middle []func(Middle) Middle
}

type Middle func(c *Client, receive *kitty.ReceivePackage)

func (client *Client) LocalAddr() net.Addr {
	return client.Conn.LocalAddr()
}

func (client *Client) RemoteAddr() net.Addr {
	return client.Conn.RemoteAddr()
}

func (client *Client) Use(middle ...func(Middle) Middle) {
	client.middle = append(client.middle, middle...)
}

func (client *Client) Emit(event []byte, body []byte, dataType int, protoType int) error {
	return client.Push(dataType, client.Protocol.Encode(event, body, dataType, protoType))
}

func (client *Client) Json(msg kitty.JsonPackage) error {
	messageJson, err := jsoniter.Marshal(kitty.JsonPackage{Event: msg.Event, Data: msg.Data})
	if err != nil {
		return err
	}
	return client.Push(kitty.TextData, messageJson)
}

func (client *Client) JsonEmit(msg kitty.JsonPackage) error {
	data, err := jsoniter.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return client.Push(kitty.TextData, client.Protocol.Encode([]byte(msg.Event), data, kitty.TextData, kitty.Json))
}

func (client *Client) ProtoBufEmit(msg kitty.ProtoBufPackage) error {
	messageProtoBuf, err := proto.Marshal(msg.Data)
	if err != nil {
		return err
	}
	return client.Push(kitty.BinData, client.Protocol.Encode([]byte(msg.Event), messageProtoBuf, kitty.BinData, kitty.ProtoBuf))
}

// Push 发送消息
func (client *Client) Push(messageType int, message []byte) error {
	client.mux.Lock()
	defer client.mux.Unlock()
	return client.Conn.WriteMessage(messageType, message)
}

func (client *Client) Close() error {
	return client.Conn.Close()
}

func (client *Client) reconnecting() {
	if client.Reconnect == true {
		time.AfterFunc(client.ReconnectInterval, func() {
			client.Connect()
		})
	}
}

// Connect 连接服务器
func (client *Client) Connect() {

	if client.Path == "" {
		client.Path = "/"
	}

	if client.Host == "" {
		panic("Host must set")
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
		client.HandshakeTimeout = 2 * time.Second
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
		client.HeartBeatInterval = 15 * time.Second
	}

	if client.HeartBeatTimeout == 0 {
		client.HeartBeatTimeout = 30 * time.Second
	}

	// 自动重连间隔
	if client.ReconnectInterval == 0 {
		client.ReconnectInterval = 1 * time.Second
	}

	if client.Protocol == nil {
		client.Protocol = &websocket2.DefaultProtocol{}
	}

	// heartbeat function
	if client.HeartBeat == nil {
		client.HeartBeat = func(client *Client) error {
			return client.Push(kitty.BinData, client.Protocol.Encode(nil, nil, kitty.PingData, kitty.BinData))
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
		HandshakeTimeout: client.HandshakeTimeout,
		WriteBufferSize:  client.WriteBufferSize,
		ReadBufferSize:   client.ReadBufferSize,
	}

	// 连接服务器
	handler, response, err := dialer.Dial(client.Scheme+"://"+client.Host+client.Path, nil)
	if err != nil {
		client.OnError(err)
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
	ticker := time.NewTicker(client.HeartBeatInterval)

	// 如果有心跳设置
	if client.AutoHeartBeat != true {
		ticker.Stop()
	}

	go func() {
		for range ticker.C {
			if err := client.HeartBeat(client); err != nil {
				client.OnError(err)
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
			client.OnError(err)
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
	if version != kitty.Version {
		return nil
	}

	// Ping
	if messageType == kitty.PingData {
		return client.PingHandler(client)("")
	}

	// Pong
	if messageType == kitty.PongData {
		return client.PongHandler(client)("")
	}

	// on router
	if client.router != nil {
		client.middleware(client, &kitty.ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType, Raw: message})
	}

	return nil
}

func (client *Client) middleware(conn *Client, msg *kitty.ReceivePackage) {
	var next Middle = client.handler
	for i := len(client.middle) - 1; i >= 0; i-- {
		next = client.middle[i](next)
	}
	next(conn, msg)
}

func (client *Client) handler(conn *Client, msg *kitty.ReceivePackage) {

	var n, formatPath = client.router.getRoute(msg.Event)
	if n == nil {
		if client.OnError != nil {
			client.OnError(errors.New(msg.Event + " " + "404 not found"))
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
