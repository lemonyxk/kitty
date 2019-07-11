package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketClientFunction func(c *Client, fte *Fte, message []byte)

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
	OnMessage func(c *Client, fte *Fte, message []byte)
	OnError   func(err interface{})
	Status    bool

	WebSocketRouter map[string]WebSocketClientFunction

	mux sync.RWMutex

	TsProto int

	Context interface{}
}

// Json 发送JSON字符
func (c *Client) Json(fte *Fte, msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("message error: %v", err)
	}

	return c.Push(fte.Type, messageJson)
}

func (c *Client) ProtoBuf(fte *Fte, msg interface{}) error {
	return nil
}

func (c *Client) Emit(fte *Fte, msg interface{}) error {

	if fte.Type == BinaryMessage {
		if j, b := msg.([]byte); b {
			return c.Push(fte.Type, j)
		}

		return fmt.Errorf("message type is bin that message must be []byte")
	}

	switch c.TsProto {
	case Json:
		return c.jsonEmit(fte.Type, fte.Event, msg)
	case ProtoBuf:
		return c.protoBufEmit(fte.Type, fte.Event, msg)
	}

	return fmt.Errorf("unknown ts ptoto")
}

func (c *Client) protoBufEmit(messageType int, event string, msg interface{}) error {
	return nil
}

func (c *Client) jsonEmit(messageType int, event string, msg interface{}) error {

	var messageJson = M{"event": event, "data": msg}

	if j, b := msg.([]byte); b {
		messageJson["data"] = string(j)
	}

	return c.Json(&Fte{Type: messageType}, messageJson)
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
		log.Println(err)
		log.Println(string(debug.Stack()))
		c.OnError(err)
		c.reconnecting()
	}
}

// Connect 连接服务器
func (c *Client) Connect() {
	// 设置LOG信息

	defer c.catchError()

	if c.TsProto == 0 {
		c.TsProto = Json
	}

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
			}
		}

	}()

	for {

		messageType, message, err := client.ReadMessage()

		if err != nil {
			c.OnError(err)
			break
		}

		if c.OnMessage != nil {
			c.OnMessage(c, &Fte{Type: messageType}, message)
		}

		if c.WebSocketRouter != nil {
			c.router(c, &Fte{Type: messageType}, message)
		}

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
