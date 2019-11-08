package lemo

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

	tire *tire.Tire

	mux sync.RWMutex

	Context interface{}

	IgnoreCase bool

	PingHandler func(c *WebSocketClient) func(appData string) error

	PongHandler func(c *WebSocketClient) func(appData string) error

	group *webSocketClientGroup
	route *webSocketClientRoute
}

// Json 发送JSON字符
func (client *WebSocketClient) Json(msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return errors.New("message err: " + err.Error())
	}

	return client.Push(TextData, messageJson)
}

func (client *WebSocketClient) JsonFormat(msg JsonPackage) error {

	messageJson, err := json.Marshal(M{"data": msg.Message, "event": msg.Event})
	if err != nil {
		return errors.New("message err: " + err.Error())
	}

	return client.Push(TextData, messageJson)
}

func (client *WebSocketClient) ProtoBuf(msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return errors.New("protobuf err: " + err.Error())
	}

	return client.Push(BinData, messageProtoBuf)

}

func (client *WebSocketClient) JsonEmit(msg JsonPackage) error {

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

	return client.Push(TextData, Pack([]byte(msg.Event), data, TextData, Json))

}

func (client *WebSocketClient) ProtoBufEmit(msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return errors.New("protobuf err: " + err.Error())
	}

	return client.Push(BinData, Pack([]byte(msg.Event), messageProtoBuf, BinData, ProtoBuf))

}

// Push 发送消息
func (client *WebSocketClient) Push(messageType int, message []byte) error {

	if client.Status == false {
		return errors.New("client is close")
	}

	client.mux.Lock()
	err := client.Conn.WriteMessage(messageType, message)
	client.mux.Unlock()
	return err
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
			return client.Push(BinData, Pack(nil, nil, PingData, BinData))
		}
	}

	if client.PingHandler == nil {
		client.PingHandler = func(connection *WebSocketClient) func(appData string) error {
			return func(appData string) error {
				return nil
			}
		}
	}

	if client.PongHandler == nil {
		client.PongHandler = func(connection *WebSocketClient) func(appData string) error {
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

	// 连接服务器
	handler, response, err := dialer.Dial(client.Protocol+"://"+client.Host+":"+strconv.Itoa(client.Port)+client.Path, nil)
	if err != nil {
		go client.OnError(NewError(err))
		return
	}

	// 设置PING处理函数
	handler.SetPingHandler(client.PingHandler(client))

	// 设置PONG处理函数
	handler.SetPongHandler(client.PongHandler(client))

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
			messageFrame, message, err := client.Conn.ReadMessage()
			if err != nil {
				closeChan <- false
				return
			}

			// unpack
			version, messageType, protoType, route, body := UnPack(message)

			if client.OnMessage != nil {
				go client.OnMessage(client, messageFrame, message)
			}

			// check version
			if version != Version {
				route, body := ParseMessage(message)
				if route != nil && client.tire != nil {
					go client.router(client, &ReceivePackage{MessageType: messageFrame, Event: string(route), Message: body, ProtoType: Json})
				}
				continue
			}

			// Ping
			if messageType == PingData {
				err := client.PingHandler(client)("")
				if err != nil {
					closeChan <- false
					return
				}
				continue
			}

			// Pong
			if messageType == PongData {
				err := client.PongHandler(client)("")
				if err != nil {
					closeChan <- false
					return
				}
				continue
			}

			// on router
			if client.tire != nil {
				go client.router(client, &ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType})
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
