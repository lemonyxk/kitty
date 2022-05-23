/**
* @program: kitty
*
* @description:
*
* @author: lemo
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
	client2 "github.com/lemonyxk/kitty/v2/socket/tcp/client"
	"github.com/lemonyxk/kitty/v2/socket/tcp/server"
)

// the same as ws and udp

var tcpServer *server.Server

var tcpClient *client2.Client

func asyncTcpServer() {

	var ready = make(chan struct{})

	tcpServer = kitty.NewTcpServer("127.0.0.1:8888")

	var tcpServerRouter = kitty.NewTcpServerRouter()

	tcpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return stream.Conn.Emit(socket.Pack{
				Event: "/hello/world",
				Data:  stream.Data,
			})
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

	tcpClient = kitty.NewTcpClient("127.0.0.1:8888")

	var clientRouter = kitty.NewTcpClientRouter()

	tcpClient.OnSuccess = func() {
		ready <- struct{}{}
	}

	go tcpClient.SetRouter(clientRouter).Connect()

	<-ready
}

func main() {
	asyncTcpServer()
	asyncTcpClient()

	var asyncClient = async.NewClient[client2.Conn](tcpClient)

	var stream, err = asyncClient.Emit(socket.Pack{
		Event: "/hello/world",
		Data:  []byte("hello world"),
	})

	log.Println(string(stream.Data), err)

	select {}
}
