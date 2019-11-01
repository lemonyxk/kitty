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
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/Lemo-yxk/tire"
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
	HandshakeTimeout  int

	// 消息处理
	OnOpen    func(c *SocketClient)
	OnClose   func(c *SocketClient)
	OnMessage func(c *SocketClient, messageType int, msg []byte)
	OnError   func(err func() *Error)
	Status    bool

	tire *tire.Tire

	mux sync.RWMutex

	Context interface{}

	IgnoreCase bool

	PingHandler func(c *SocketClient) func(appData string) error

	PongHandler func(c *SocketClient) func(appData string) error

	group *socketClientGroup
	route *socketClientRoute
}

// Json 发送JSON字符
func (client *SocketClient) Json(msg interface{}) error {

	messageJson, err := json.Marshal(msg)
	if err != nil {
		return errors.New("message err: " + err.Error())
	}

	return client.Push(messageJson)
}

func (client *SocketClient) ProtoBuf(msg proto.Message) error {

	messageProtoBuf, err := proto.Marshal(msg)
	if err != nil {
		return errors.New("protobuf err: " + err.Error())
	}

	return client.Push(messageProtoBuf)

}

func (client *SocketClient) JsonEmit(msg JsonPackage) error {

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

	return client.Push(Pack([]byte(msg.Event), data, TextData, Json))

}

func (client *SocketClient) ProtoBufEmit(msg ProtoBufPackage) error {

	messageProtoBuf, err := proto.Marshal(msg.Message)
	if err != nil {
		return errors.New("protobuf err: " + err.Error())
	}

	return client.Push(Pack([]byte(msg.Event), messageProtoBuf, BinData, ProtoBuf))

}

// Push 发送消息
func (client *SocketClient) Push(message []byte) error {

	if client.Status == false {
		return errors.New("client is close")
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	_, err := client.Conn.Write(message)
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
			return client.Push(Pack(nil, nil, PingData, BinData))
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
		go client.OnError(NewError(err))
		return
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
					if !IsHeaderInvalid(message) {
						go client.OnError(NewError("invalid header"))
						goto OUT
					}

					singleMessageLen = GetLen(message)
				}

				// jump out and read continue
				if len(message) < singleMessageLen {
					break
				}

				// a complete message
				err := client.decodeMessage(client, message[0:singleMessageLen])
				if err != nil {
					go client.OnError(NewError(err))
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
	version, messageType, protoType, route, body := UnPack(message)

	// check version
	if version != Version {
		if client.OnMessage != nil {
			go client.OnMessage(connection, messageType, message)
		}
		return nil
	}

	// Ping
	if messageType == PingData {
		err := client.PingHandler(connection)("")
		if err != nil {
			return err
		}
		return nil
	}

	// Pong
	if messageType == PongData {
		err := client.PongHandler(connection)("")
		if err != nil {
			return err
		}
		return nil
	}

	// // check message type
	// if frameType != messageType {
	// 	return errors.New("frame type not match message type")
	// }

	// on router
	if client.tire != nil {
		var receivePackage = &ReceivePackage{MessageType: messageType, Event: string(route), Message: body, ProtoType: protoType}
		go client.router(connection, receivePackage)
		return nil
	}

	// any way run check on message
	if client.OnMessage != nil {
		go client.OnMessage(connection, messageType, message)
	}
	return nil
}
