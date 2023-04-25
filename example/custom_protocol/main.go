/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-23 21:58
**/

package main

import (
	"log"
	"time"

	"github.com/lemonyxk/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/tcp/client"
	"github.com/lemonyxk/kitty/socket/tcp/server"
)

// The same as websocket and udp.
//
// Implement the Protocol.
//
// 	type Protocol interface {
// 		Decode(message []byte) (messageType byte, id int64, route []byte, body []byte)
// 		Encode(messageType byte, id int64, route []byte, body []byte) []byte
// 		Reader() func(n int, buf []byte, fn func(bytes []byte)) error
// 		HeadLen() int
// 		Ping() []byte
// 		Pong() []byte
// 		IsPong(messageType byte) bool
// 		IsPing(messageType byte) bool
// 		IsUnknown(messageType byte) bool
// 	}

var tcpServer *server.Server

var tcpClient *client.Client

func asyncTcpServer() {

	var ready = make(chan struct{})

	tcpServer = kitty.NewTcpServer("127.0.0.1:8888")

	tcpServer.Protocol = &CustomTcp{}

	var tcpServerRouter = kitty.NewTcpServerRouter()

	tcpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data()))
			return stream.Emit(stream.Event(), stream.Data())
		})
	})

	tcpServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go tcpServer.SetRouter(tcpServerRouter).Start()

	<-ready
}

func asyncTcpClient() {

	var ready = make(chan struct{})
	var isRun = false

	tcpClient = kitty.NewTcpClient("127.0.0.1:8888")

	tcpClient.Protocol = &CustomTcp{}

	var clientRouter = kitty.NewTcpClientRouter()

	tcpClient.OnError = func(stream *socket.Stream[client.Conn], err error) {
		log.Println(err)
	}

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			time.Sleep(time.Second)
			return stream.Emit(stream.Event(), stream.Data())
		})
	})

	tcpClient.OnSuccess = func() {
		if isRun {
			return
		}
		ready <- struct{}{}
	}

	go tcpClient.SetRouter(clientRouter).Connect()

	<-ready
	isRun = true
}

func main() {
	asyncTcpServer()
	asyncTcpClient()

	var err = tcpClient.Sender().Emit("/hello/world", []byte("hello world"))

	log.Println(err)

	select {}
}
