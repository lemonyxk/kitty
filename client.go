package lemo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type WebSocketClientFunction func(c *Client, receive *ReceivePackage) func() *Error

// Client 客户端
type Client struct {
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
	OnError   func(err func() *Error)
	Status    bool

	Router map[string]WebSocketClientFunction

	mux sync.RWMutex

	TsProto int

	Context interface{}
}

// Json 发送JSON字符
func (c *Client) Json(msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return c.Push(TextMessage, messageJson)
}

func (c *Client) ProtoBuf(msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("protobuf error: %v", err)
	}

	return c.Push(BinaryMessage, messageProtoBuf)

}

func (c *Client) JsonEmit(msg JsonPackage) error {

	var data = []byte{13, 10}

	if msg.Event == "" {
		msg.Event = "/"
	}

	data = append(data, byte(len(msg.Event)), Json)
	data = append(data, []byte(msg.Event)...)

	if mb, ok := msg.Message.([]byte); ok {
		msg.Message = string(mb)
	}

	messageJson, err := json.Marshal(msg.Message)
	if err != nil {
		return fmt.Errorf("protobuf error: %v", err)
	}

	data = append(data, messageJson...)

	return c.Push(TextMessage, data)

}

func (c *Client) ProtoBufEmit(msg ProtoBufPackage) error {

	var data = []byte{13, 10}

	if msg.Event == "" {
		msg.Event = "/"
	}

	data = append(data, byte(len(msg.Event)), ProtoBuf)
	data = append(data, []byte(msg.Event)...)

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return fmt.Errorf("protobuf error: %v", err)
	}

	data = append(data, messageProtoBuf...)

	return c.Push(BinaryMessage, data)

}

// Push 发送消息
func (c *Client) Push(messageType int, message []byte) error {

	if c.Status == false {
		return fmt.Errorf("client is close")
	}

	// 默认为文本
	if messageType == 0 {
		messageType = TextMessage
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.Conn.WriteMessage(messageType, message)
}

func (c *Client) Close() error {
	c.Reconnect = false
	return c.Conn.Close()
}

func (c *Client) reconnecting() {
	if c.Reconnect == true {
		time.AfterFunc(time.Duration(c.ReconnectInterval)*time.Second, func() {
			c.Connect()
		})
	}
}

func (c *Client) catchError() {
	if err := recover(); err != nil {
		if c.OnError != nil {
			go c.OnError(NewErrorFromDeep(err, 2))
		}
		c.reconnecting()
	}
}

// Connect 连接服务器
func (c *Client) Connect() {
	// 设置LOG信息

	defer c.catchError()

	var closeChan = make(chan bool)

	if c.Host == "" {
		c.Host = "127.0.0.1"
	}

	if c.Port == 0 {
		c.Port = 1207
	}

	if c.Protocol == "" {
		c.Protocol = "ws"
	}

	if c.Path == "" {
		c.Path = "/"
	}

	if c.OnOpen == nil {
		panic("OnOpen must set")
	}

	if c.OnClose == nil {
		panic("OnClose must set")
	}

	if c.OnError == nil {
		panic("OnError must set")
	}

	// 握手
	if c.HandshakeTimeout == 0 {
		c.HandshakeTimeout = 2
	}

	// 写入BUF大小
	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 1024 * 1024 * 2
	}

	// 读出BUF大小
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 1024 * 1024 * 2
	}

	// 定时心跳间隔
	if c.HeartBeatInterval == 0 {
		c.HeartBeatInterval = 15
	}

	// 自动重连间隔
	if c.ReconnectInterval == 0 {
		c.ReconnectInterval = 1
	}

	var dialer = websocket.Dialer{
		HandshakeTimeout: time.Duration(c.HandshakeTimeout) * time.Second,
		WriteBufferSize:  c.WriteBufferSize,
		ReadBufferSize:   c.ReadBufferSize,
	}

	// 连接服务器
	client, response, err := dialer.Dial(fmt.Sprintf("%s://%s:%d%s", c.Protocol, c.Host, c.Port, c.Path), nil)
	if err != nil {
		panic(err)
	}

	c.Response = response

	c.Conn = client

	c.Status = true

	// 连接成功
	go c.OnOpen(c)

	// 定时器 心跳
	ticker := time.NewTicker(time.Duration(c.HeartBeatInterval) * time.Second)

	// 如果有心跳设置
	if c.AutoHeartBeat == true {
		if c.HeartBeat == nil {
			c.HeartBeat = func(c *Client) error {
				return c.Push(websocket.PingMessage, nil)
			}
		}
	} else {
		ticker.Stop()
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := c.HeartBeat(c); err != nil {
					closeChan <- false
					break
				}
			}
		}

	}()

	go func() {
		for {

			messageType, message, err := client.ReadMessage()

			if err != nil {
				closeChan <- false
				break
			}

			go func() {

				var mLen = len(message)
				var event string
				var data []byte

				// 空消息
				if mLen == 0 {
					return
				}

				//
				if mLen < 4 {
					if c.OnMessage != nil {
						c.OnMessage(c, messageType, message)
					}
					return
				}

				// not proto or json type
				if message[0] != 13 || message[1] != 10 || (message[3] != Json && message[3] != ProtoBuf) {
					if c.OnMessage != nil {
						c.OnMessage(c, messageType, message)
					}
					return
				}

				if message[2] == 0 {
					event = "/"
					data = nil
				} else {
					if mLen < int(message[2])+4 {
						if c.OnMessage != nil {
							c.OnMessage(c, messageType, message)
						}
						return
					}
					event = string(message[4 : 4+message[2]])
					data = message[message[2]+4:]
				}

				if c.Router != nil {
					var receivePackage = &ReceivePackage{MessageType: messageType, Event: event, Message: data, FormatType: message[3]}
					c.router(c, receivePackage)
					return
				}
			}()

		}
	}()

	<-closeChan

	// 关闭定时器
	ticker.Stop()
	// 更改状态
	c.Status = false
	// 关闭连接
	_ = client.Close()
	// 触发回调
	go c.OnClose(c)
	// 触发重连设置
	c.reconnecting()
}
