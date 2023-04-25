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
	"github.com/lemonyxk/kitty/socket/websocket/client"
	"github.com/lemonyxk/kitty/socket/websocket/server"
)

var wsServer *server.Server

var wsClient *client.Client

func runWsServer() {

	var ready = make(chan struct{})

	wsServer = kitty.NewWebSocketServer("127.0.0.1:8888")
	// wsServer.CertFile = "example/ssl/localhost+2.pem"
	// wsServer.KeyFile = "example/ssl/localhost+2-key.pem"

	wsServer.HeartBeatTimeout = time.Second * 3

	var wsServerRouter = kitty.NewWebSocketServerRouter()

	wsServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data()), stream.MessageID())
			return stream.Emit(stream.Event(), stream.Data())
		})
	})

	wsServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go wsServer.SetRouter(wsServerRouter).Start()

	<-ready
}

func runWsClient() {

	var ready = make(chan struct{})
	var isRun = false

	wsClient = kitty.NewWebSocketClient("ws://127.0.0.1:8888")
	// wsClient.CertFile = "example/ssl/localhost+2.pem"
	// wsClient.KeyFile = "example/ssl/localhost+2-key.pem"

	wsClient.HeartBeatTimeout = time.Second * 5
	wsClient.HeartBeatInterval = time.Second * 2

	var clientRouter = kitty.NewWebSocketClientRouter()

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			return nil
			// return stream.Emit(stream.Event, stream.Data)
		})
	})

	wsClient.OnSuccess = func() {
		if isRun {
			return
		}
		ready <- struct{}{}
	}

	go wsClient.SetRouter(clientRouter).Connect()

	<-ready
	isRun = true
}

func main() {
	runWsServer()
	runWsClient()

	for {
		var err = wsClient.Sender().Emit("/hello/world", []byte("hello world"))
		if err != nil {
			log.Println(err)
			break
		}
		time.Sleep(time.Second)
	}

	select {}
}
