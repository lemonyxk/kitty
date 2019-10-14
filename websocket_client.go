package lemo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Lemo-yxk/tire"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// WebSocketClient 客户端
type WebSocketClient struct {
	// 服务器信息
	Protocol string
	Host     string
	Port     int
	Path     string
	// Origin   http.Header

	// 客户端信息
	Conn              *websocket.Conn
	Response          *http.Response
	AutoHeartBeat     bool
	HeartBeatTimeout  int
	HeartBeatInterval int
	HeartBeat         func(c *WebSocketClient) error
	Reconnect         bool
	ReconnectInterval int
	WriteBufferSize   int
	ReadBufferSize    int
	HandshakeTimeout  int

	// 消息处理
	OnOpen    func(c *WebSocketClient)
	OnClose   func(c *WebSocketClient)
	OnMessage func(c *WebSocketClient, messageType int, msg []byte)
	OnError   func(err func() *Error)
	Status    bool

	Router *tire.Tire

	mux sync.RWMutex

	TsProto int

	Context interface{}

	IgnoreCase bool

	group *WebSocketClientGroup
	route *WebSocketClientRoute
}

// Json 发送JSON字符
func (client *WebSocketClient) Json(msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return client.Push(TextMessage, messageJson)
}

func (client *WebSocketClient) ProtoBuf(msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("protobuf error: %v", err)
	}

	return client.Push(BinaryMessage, messageProtoBuf)

}

func (client *WebSocketClient) JsonEmit(msg JsonPackage) error {

	var data []byte

	if mb, ok := msg.Message.([]byte); ok {
		data = mb
	} else {
		messageJson, err := json.Marshal(msg.Message)
		if err != nil {
			return fmt.Errorf("protobuf error: %v", err)
		}
		data = messageJson
	}

	return client.Push(TextMessage, Pack([]byte(msg.Event), data, Json, TextMessage))

}

func (client *WebSocketClient) ProtoBufEmit(msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return fmt.Errorf("protobuf error: %v", err)
	}

	return client.Push(BinaryMessage, Pack([]byte(msg.Event), messageProtoBuf, ProtoBuf, BinaryMessage))

}

// Push 发送消息
func (client *WebSocketClient) Push(messageType int, message []byte) error {

	if client.Status == false {
		return fmt.Errorf("client is close")
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	return client.Conn.WriteMessage(messageType, message)
}

func (client *WebSocketClient) Close() error {
	client.Reconnect = false
	return client.Conn.Close()
}

func (client *WebSocketClient) reconnecting() {
	if client.Reconnect == true {
		time.AfterFunc(time.Duration(client.ReconnectInterval)*time.Second, func() {
			client.Connect()
		})
	}
}

// Connect 连接服务器
func (client *WebSocketClient) Connect() {
	// 设置LOG信息

	var closeChan = make(chan bool)

	if client.Host == "" {
		client.Host = "127.0.0.1"
	}

	if client.Port == 0 {
		client.Port = 1207
	}

	if client.Protocol == "" {
		client.Protocol = "ws"
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
		client.WriteBufferSize = 1024 * 1024 * 2
	}

	// 读出BUF大小
	if client.ReadBufferSize == 0 {
		client.ReadBufferSize = 1024 * 1024 * 2
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
		client.HeartBeat = func(client *WebSocketClient) error {
			return client.Push(websocket.PingMessage, nil)
		}
	}

	var dialer = websocket.Dialer{
		HandshakeTimeout: time.Duration(client.HandshakeTimeout) * time.Second,
		WriteBufferSize:  client.WriteBufferSize,
		ReadBufferSize:   client.ReadBufferSize,
	}

	// 连接服务器
	handler, response, err := dialer.Dial(fmt.Sprintf("%s://%s:%d%s", client.Protocol, client.Host, client.Port, client.Path), nil)
	if err != nil {
		panic(err)
	}

	// 设置PING处理函数
	handler.SetPingHandler(func(appData string) error {
		return nil
	})

	// 设置PONG处理函数
	handler.SetPongHandler(func(appData string) error {
		return nil
	})

	client.Response = response

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
		for {

			// read message
			frameType, message, err := client.Conn.ReadMessage()
			if err != nil {
				closeChan <- false
				return
			}

			// unpack
			version, messageType, protoType, route, body := UnPack(message)

			// check version
			if version != Version {
				if client.OnMessage != nil {
					go client.OnMessage(client, messageType, message)
				}
				continue
			}

			// check message type
			if frameType != messageType {
				closeChan <- false
				return
			}

			// Ping
			if messageType == PingMessage {
				err := client.Conn.PingHandler()("")
				if err != nil {
					closeChan <- false
					return
				}
				continue
			}

			// Pong
			if messageType == PongMessage {
				err := client.Conn.PongHandler()("")
				if err != nil {
					closeChan <- false
					return
				}
				continue
			}

			// on router
			if client.Router != nil {
				var receivePackage = &ReceivePackage{MessageType: messageType, Event: route, Message: body, ProtoType: protoType}
				go client.router(client, receivePackage)
				continue
			}

		}
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
