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

	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/async"
	"github.com/lemonyxk/kitty/v2/socket/tcp/client"
	"github.com/lemonyxk/kitty/v2/socket/tcp/server"
)

// the same as ws and udp

var tcpServer *server.Server

var tcpClient *client.Client

func asyncTcpServer() {

	var ready = make(chan struct{})

	tcpServer = kitty.NewTcpServer("127.0.0.1:8888")

	var tcpServerRouter = kitty.NewTcpServerRouter()

	tcpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return stream.Emit(stream.Event, stream.Data)
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

	var clientRouter = kitty.NewTcpClientRouter()

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

	var asyncClient = async.NewClient[client.Conn](tcpClient)

	var stream, err = asyncClient.Emit("/hello/world", []byte("hello world"))

	log.Println(string(stream.Data), err)

	select {}
}
