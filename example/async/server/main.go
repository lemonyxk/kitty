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

	"github.com/lemonyxk/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/tcp/client"
	"github.com/lemonyxk/kitty/socket/tcp/server"
)

// the same as ws and udp

var tcpServer *server.Server[any]

var tcpClient *client.Client[any]

var fd int64 = 0

func asyncTcpServer() {

	var ready = make(chan struct{})

	tcpServer = kitty.NewTcpServer[any]("127.0.0.1:8888")

	tcpServer.OnOpen = func(conn server.Conn) {
		fd++
	}

	var tcpServerRouter = kitty.NewTcpServerRouter[any]()

	tcpServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go tcpServer.SetRouter(tcpServerRouter).Start()

	<-ready
}

func asyncTcpClient() {

	var ready = make(chan struct{})
	var isRun = false

	tcpClient = kitty.NewTcpClient[any]("127.0.0.1:8888")

	var clientRouter = kitty.NewTcpClientRouter[any]()

	tcpClient.OnSuccess = func() {
		if isRun {
			return
		}
		ready <- struct{}{}
	}

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			return stream.Emit(stream.Event(), stream.Data())
		})
	})

	go tcpClient.SetRouter(clientRouter).Connect()

	<-ready
	isRun = true
}

func main() {
	asyncTcpServer()
	asyncTcpClient()

	var asyncServer = socket.NewAsyncServer[server.Conn, any](tcpServer)
	sender, err := asyncServer.Sender(fd)
	if err != nil {
		panic(err)
	}

	stream, err := sender.Emit("/hello/world", []byte("hello world"))
	if err != nil {
		panic(err)
	}

	log.Println(string(stream.Data()), err)

	select {}
}
