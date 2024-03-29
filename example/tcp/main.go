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

var tcpServer *server.Server[any]

var tcpClient *client.Client[any]

func runTcpServer() {

	var ready = make(chan struct{})

	tcpServer = kitty.NewTcpServer[any]("127.0.0.1:8888")

	tcpServer.HeartBeatTimeout = time.Second * 5

	// tcpServer.CertFile = "example/ssl/localhost+2.pem"
	// tcpServer.KeyFile = "example/ssl/localhost+2-key.pem"

	var tcpServerRouter = kitty.NewTcpServerRouter[any]()

	tcpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data()))
			var sender, _ = tcpServer.Sender(100)
			if sender != nil {
				sender.Emit("/hello/world", []byte("hello world"))
			}

			return stream.Emit(stream.Event(), stream.Data())
		})
	})

	tcpServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go tcpServer.SetRouter(tcpServerRouter).Start()

	<-ready
}

func runTcpClient() {

	var ready = make(chan struct{})
	var isRun = false

	tcpClient = kitty.NewTcpClient[any]("127.0.0.1:8888")

	tcpClient.HeartBeatTimeout = time.Second * 3
	tcpClient.HeartBeatInterval = time.Second * 1

	// tcpClient.CertFile = "example/ssl/localhost+2.pem"
	// tcpClient.KeyFile = "example/ssl/localhost+2-key.pem"

	var clientRouter = kitty.NewTcpClientRouter[any]()

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			time.Sleep(time.Second)
			return stream.Emit(stream.Event(), stream.Data())
		})
	})

	// make sure the event run only once
	// because when the client reconnect, the event will be run again and chan will be blocked.
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

	runTcpServer()
	runTcpClient()

	var err = tcpClient.Sender().Emit("/hello/world", []byte("hello world"))

	log.Println(err)

	select {}
}
