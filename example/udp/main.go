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

	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/udp/client"
	"github.com/lemonyxk/kitty/v2/socket/udp/server"
)

var udpServer *server.Server

var udpClient *client.Client

func runUdpServer() {

	var ready = make(chan struct{})

	udpServer = kitty.NewUdpServer("127.0.0.1:8888")

	udpServer.HeartBeatTimeout = time.Second * 5

	var udpServerRouter = kitty.NewUdpServerRouter()

	udpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data))
			return stream.Conn.Emit(stream.Event, stream.Data)
		})
	})

	udpServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go udpServer.SetRouter(udpServerRouter).Start()

	<-ready
}

func runUdpClient() {

	var ready = make(chan struct{}, 100)
	var isRun = false

	udpClient = kitty.NewUdpClient("127.0.0.1:8888")

	udpClient.HeartBeatTimeout = time.Second * 2
	udpClient.HeartBeatInterval = time.Second * 3

	var clientRouter = kitty.NewUdpClientRouter()

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			time.Sleep(time.Second)
			return stream.Conn.Emit(stream.Event, stream.Data)
		})
	})

	udpClient.OnSuccess = func() {
		if isRun {
			return
		}
		ready <- struct{}{}
	}

	go udpClient.SetRouter(clientRouter).Connect()

	<-ready
	isRun = true
}

func main() {

	runUdpServer()
	runUdpClient()

	var err = udpClient.Emit("/hello/world", []byte("hello world"))

	log.Println(err)

	select {}
}
