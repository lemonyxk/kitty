package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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

	// PingMessage PING
	PingMessage int
	// PongMessage PONG
	PongMessage int
	// TextMessage 文本
	TextMessage int
	// BinaryMessage 二进制
	BinaryMessage int

	MessageRouter map[string]func(c *Client, messageType int, message []byte)

	GlobalConfig M

	BeforeSend func(route string, message M) error

	mux sync.RWMutex
}

func (c *Client) SetGlobalConfig(key string, value interface{}) {
	c.GlobalConfig[key] = value
}

func (c *Client) SetRoute(route string, f func(c *Client, messageType int, message []byte)) {
	c.MessageRouter[route] = f
}

// Json 发送JSON字符
func (c *Client) Json(route string, message M) error {

	if c.BeforeSend != nil {
		err := c.BeforeSend(route, message)
		if err != nil {
			return err
		}
	}

	if message == nil {
		message = make(M)
	}

	jsonMessage := M{
		"event": route,
		"data":  message,
	}

	if c.GlobalConfig != nil {
		for k, v := range c.GlobalConfig {
			jsonMessage["data"].(M)[k] = v
		}
	}

	data, err := json.Marshal(jsonMessage)
	if err != nil {
		return err
	}

	return c.Push(c.TextMessage, data)
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

func (c *Client) Reconnecting() {
	if c.Reconnect == true {
		time.AfterFunc(time.Duration(c.ReconnectInterval)*time.Second, func() {
			c.Connect()
		})
	}
}

func (c *Client) CatchError() {
	if err := recover(); err != nil {
		log.Println(err)
		c.OnError(err)
		c.Reconnecting()
	}
}

// Connect 连接服务器
func (c *Client) Connect() {
	// 设置LOG信息

	defer c.CatchError()

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

	c.PingMessage = websocket.PingMessage

	c.PongMessage = websocket.PongMessage

	c.TextMessage = websocket.TextMessage

	c.BinaryMessage = websocket.BinaryMessage

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

	c.GlobalConfig = make(M)

	c.MessageRouter = make(map[string]func(c *Client, messageType int, message []byte))

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
				c.Push(websocket.PingMessage, nil)
			}
		}
	} else {
		ticker.Stop()
	}

	go func() {

		defer c.CatchError()

		for {
			select {
			case <-ticker.C:
				c.HeartBeat(c)
			case message := <-ch:
				c.OnMessage(c, messageType, message)
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
	client.Close()
	// 触发回调
	c.OnClose(c)
	// 触发重连设置
	c.Reconnecting()
}
