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

var wsServer *server.Server[any]

var wsClient *client.Client[any]

func runWsServer() {

	var ready = make(chan struct{})

	wsServer = kitty.NewWebSocketServer[any]("127.0.0.1:8888")
	// wsServer.CertFile = "example/ssl/localhost+2.pem"
	// wsServer.KeyFile = "example/ssl/localhost+2-key.pem"

	wsServer.HeartBeatTimeout = time.Second * 3

	var wsServerRouter = kitty.NewWebSocketServerRouter[any]()

	wsServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn], any]) {
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

	wsClient = kitty.NewWebSocketClient[any]("ws://127.0.0.1:8888")
	// wsClient.CertFile = "example/ssl/localhost+2.pem"
	// wsClient.KeyFile = "example/ssl/localhost+2-key.pem"

	wsClient.HeartBeatTimeout = time.Second * 5
	wsClient.HeartBeatInterval = time.Second * 2

	var clientRouter = kitty.NewWebSocketClientRouter[any]()

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn], any]) {
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
