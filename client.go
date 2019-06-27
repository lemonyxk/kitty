package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketClientFunction func(c *Client, messageType int, message []byte)

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
	AutoHeartBeat     bool
	HeartBeatInterval int
	HeartBeat         func(c *Client)
	Reconnect         bool
	ReconnectInterval int
	WriteBufferSize   int
	ReadBufferSize    int
	HandshakeTimeout  int

	// 消息处理
	OnOpen    func(c *Client)
	OnClose   func(c *Client)
	OnMessage func(c *Client, messageType int, message []byte)
	OnError   func(err interface{})
	Status    bool

	WebSocketRouter map[string]WebSocketClientFunction

	mux sync.RWMutex

	TsProto int

	Context interface{}
}

// Json 发送JSON字符
func (c *Client) Json(messageType int, message M) error {

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return c.Push(messageType, data)
}

func (c *Client) ProtoBuf(messageType int, message []byte) error {
	return nil
}

func (c *Client) Emit(messageType int, event string, message M) error {
	switch c.TsProto {
	case Json:
		return c.jsonEmit(messageType, event, message)
	case ProtoBuf:
		return c.protoBufEmit(messageType, event, message)
	}

	return fmt.Errorf("unknown ts ptoto")
}

func (c *Client) protoBufEmit(messageType int, event string, message M) error {
	return nil
}

func (c *Client) jsonEmit(messageType int, event string, message M) error {

	var data = dataPackage{Event: event, Data: message}

	messageJson, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return c.Push(messageType, messageJson)
}

// Push 发送消息
func (c *Client) Push(messageType int, message []byte) error {

	if c.Status == false {
		return fmt.Errorf("client is close")
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	return c.Conn.WriteMessage(messageType, message)
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
		log.Println(err)
		c.OnError(err)
		c.reconnecting()
	}
}

// Connect 连接服务器
func (c *Client) Connect() {
	// 设置LOG信息

	defer c.catchError()

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
		log.Panicln("OnOpen must set")
	}

	if c.OnClose == nil {
		log.Panicln("OnClose must set")
	}

	if c.OnMessage == nil {
		log.Panicln("OnMessage must set")
	}

	if c.OnError == nil {
		log.Panicln("OnError must set")
	}

	var dialer websocket.Dialer

	// 握手
	if c.HandshakeTimeout == 0 {
		c.HandshakeTimeout = 2
	}
	dialer.HandshakeTimeout = time.Duration(c.HandshakeTimeout) * time.Second

	// 写入BUF大小
	if c.WriteBufferSize == 0 {
		dialer.WriteBufferSize = 1024 * 1024 * 2
	}

	// 读出BUF大小
	if c.ReadBufferSize == 0 {
		dialer.ReadBufferSize = 1024 * 1024 * 2
	}

	// 定时心跳间隔
	if c.HeartBeatInterval == 0 {
		c.HeartBeatInterval = 15
	}

	// 自动重连间隔
	if c.ReconnectInterval == 0 {
		c.ReconnectInterval = 1
	}

	// 连接服务器
	client, _, err := dialer.Dial(fmt.Sprintf("%s://%s:%d%s", c.Protocol, c.Host, c.Port, c.Path), nil)
	if err != nil {
		log.Panicln(err)
	}

	c.Conn = client

	c.Status = true

	// 连接成功
	c.OnOpen(c)

	// 接收消息
	ch := make(chan []byte)

	var messageType int
	var message []byte

	// 定时器 心跳
	ticker := time.NewTicker(time.Duration(c.HeartBeatInterval) * time.Second)

	// 如果有心跳设置
	if c.AutoHeartBeat == true {
		if c.HeartBeat == nil {
			c.HeartBeat = func(c *Client) {
				_ = c.Push(websocket.PingMessage, nil)
			}
		}
	} else {
		ticker.Stop()
	}

	go func() {

		defer c.catchError()

		for {
			select {
			case <-ticker.C:
				c.HeartBeat(c)
			case message := <-ch:
				if c.OnMessage != nil {
					c.OnMessage(c, messageType, message)
				}
				if c.WebSocketRouter != nil {
					c.router(c, messageType, message)
				}
			}
		}

	}()

	for {

		messageType, message, err = client.ReadMessage()

		if err != nil {
			log.Println(err)
			break
		}

		ch <- message

	}

	// 关闭定时器
	ticker.Stop()
	// 更改状态
	c.Status = false
	// 关闭连接
	_ = client.Close()
	// 触发回调
	c.OnClose(c)
	// 触发重连设置
	c.reconnecting()
}
