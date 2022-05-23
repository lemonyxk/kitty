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
	"time"

	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	client2 "github.com/lemonyxk/kitty/v2/socket/udp/client"
	"github.com/lemonyxk/kitty/v2/socket/udp/server"
)

var udpServer *server.Server

var udpClient *client2.Client

func runUdpServer() {

	var ready = make(chan struct{})

	udpServer = kitty.NewUdpServer("127.0.0.1:8888")

	var udpServerRouter = kitty.NewUdpServerRouter()

	udpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data))
			return stream.Conn.Emit(socket.Pack{
				Event: "/hello/world",
				Data:  stream.Data,
			})
		})
	})

	udpServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go udpServer.SetRouter(udpServerRouter).Start()

	<-ready
}

func runUdpClient() {

	var ready = make(chan struct{})

	udpClient = kitty.NewUdpClient("127.0.0.1:8888")

	var clientRouter = kitty.NewUdpClientRouter()

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client2.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client2.Conn]) error {
			time.Sleep(time.Second)
			return stream.Conn.Emit(socket.Pack{
				Event: stream.Event,
				Data:  stream.Data,
			})
		})
	})

	udpClient.OnSuccess = func() {
		ready <- struct{}{}
	}

	go udpClient.SetRouter(clientRouter).Connect()

	<-ready
}

func main() {
	runUdpServer()
	runUdpClient()

	var err = udpClient.Emit(socket.Pack{
		Event: "/hello/world",
		Data:  []byte("hello world"),
	})

	log.Println(err)

	select {}
}
