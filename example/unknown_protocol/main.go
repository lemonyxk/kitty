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

	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/websocket/client"
	"github.com/lemonyxk/kitty/v2/socket/websocket/server"
)

// websocket has been subcontracted and udp is not streaming data,
// so there is an unknown possibility.

var wsServer *server.Server

var wsClient *client.Client

func asyncWsServer() {

	var ready = make(chan struct{})

	wsServer = kitty.NewWebSocketServer("127.0.0.1:8888")

	var wsServerRouter = kitty.NewWebSocketServerRouter()

	// route:message
	wsServer.OnUnknown = func(conn server.Conn, message []byte, next server.Middle) {
		var index = strings.IndexByte(string(message), ':')
		if index == -1 {
			return
		}

		var route = message[:index]
		var data = message[index+1:]

		next(&socket.Stream[server.Conn]{Conn: conn, Pack: socket.Pack{Event: string(route), Data: data}})
	}

	wsServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			log.Println(string(stream.Data))
			return stream.Conn.Push(packMessage(stream.Event, string(stream.Data)))
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

	wsClient = kitty.NewWebSocketClient("ws://127.0.0.1:8888")

	var clientRouter = kitty.NewWebSocketClientRouter()

	wsClient.OnError = func(err error) {
		log.Println(err)
	}

	wsClient.OnUnknown = func(conn client.Conn, message []byte, next client.Middle) {
		var index = strings.IndexByte(string(message), ':')
		if index == -1 {
			return
		}

		var route = message[:index]
		var data = message[index+1:]

		next(&socket.Stream[client.Conn]{Conn: conn, Pack: socket.Pack{Event: string(route), Data: data}})
	}

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			time.Sleep(time.Second)
			return stream.Conn.Push(packMessage(stream.Event, string(stream.Data)))
		})
	})

	wsClient.OnSuccess = func() {
		ready <- struct{}{}
	}

	go wsClient.SetRouter(clientRouter).Connect()

	<-ready
}

func packMessage(a, b string) []byte {
	return []byte(a + ":" + b)
}

func main() {
	asyncWsServer()
	asyncWsClient()

	var err = wsClient.Push(packMessage("/hello/world", "hello world"))

	log.Println(err)

	select {}
}
