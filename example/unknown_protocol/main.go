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
	"strings"
	"time"

	"github.com/lemonyxk/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/websocket/client"
	"github.com/lemonyxk/kitty/socket/websocket/server"
)

// websocket has been subcontracted and udp is not streaming data,
// so there is an unknown possibility.

var wsServer *server.Server[any]

var wsClient *client.Client[any]

func asyncWsServer() {

	var ready = make(chan struct{})

	wsServer = kitty.NewWebSocketServer[any]("127.0.0.1:8888")

	var wsServerRouter = kitty.NewWebSocketServerRouter[any]()

	// route:message
	wsServer.OnUnknown = func(conn server.Conn, message []byte, next server.Middle) {
		var index = strings.IndexByte(string(message), ':')
		if index == -1 {
			return
		}

		var route = message[:index]
		var data = message[index+1:]

		next(socket.NewStream(conn, 0, 0, 0, 0, route, data))
	}

	wsServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data()))
			return stream.Conn().Push(packMessage(stream.Event(), string(stream.Data())))
		})
	})

	wsServer.OnSuccess = func() {
		ready <- struct{}{}
	}

	go wsServer.SetRouter(wsServerRouter).Start()

	<-ready
}

func asyncWsClient() {

	var ready = make(chan struct{})
	var isRun = false

	wsClient = kitty.NewWebSocketClient[any]("ws://127.0.0.1:8888")

	var clientRouter = kitty.NewWebSocketClientRouter[any]()

	wsClient.OnError = func(stream *socket.Stream[client.Conn], err error) {
		log.Println(err)
	}

	wsClient.OnUnknown = func(conn client.Conn, message []byte, next client.Middle) {
		var index = strings.IndexByte(string(message), ':')
		if index == -1 {
			return
		}

		var route = message[:index]
		var data = message[index+1:]

		next(socket.NewStream(conn, 0, 0, 0, 0, route, data))
	}

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			time.Sleep(time.Second)
			return stream.Conn().Push(packMessage(stream.Event(), string(stream.Data())))
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

func packMessage(a, b string) []byte {
	return []byte(a + ":" + b)
}

func main() {
	asyncWsServer()
	asyncWsClient()

	var err = wsClient.Conn().Push(packMessage("/hello/world", "hello world"))

	log.Println(err)

	select {}
}
