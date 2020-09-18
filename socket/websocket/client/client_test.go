/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2020-09-18 16:40
**/

package client

import (
	"strings"
	"sync"
	"testing"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/websocket/server"
)

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var webSocketServer *server.Server

var webSocketServerRouter *server.Router

var client *Client

var clientRouter *Router

func initServer(fn func()) {

	// create server
	webSocketServer = &server.Server{Host: "127.0.0.1:8667", Path: "/"}

	// event
	webSocketServer.OnOpen = func(conn *server.Conn) {}
	webSocketServer.OnClose = func(conn *server.Conn) {}
	webSocketServer.OnError = func(err error) {}
	webSocketServer.OnMessage = func(conn *server.Conn, messageType int, msg []byte) {}

	// middleware
	webSocketServer.Use(func(next server.Middle) server.Middle {
		return func(conn *server.Conn, stream *socket.Stream) {
			next(conn, stream)
		}
	})

	// create router
	webSocketServerRouter = &server.Router{IgnoreCase: true}

	// set group route
	webSocketServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Conn, stream *socket.Stream) error {
			return conn.Json(socket.JsonPackage{
				Event: "/hello/world",
				Data:  "i am server",
			})
		})
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	webSocketServer.OnSuccess = func() {
		fn()
	}
}

func initClient(fn func()) {
	// create client
	client = &Client{Scheme: "ws", Host: "127.0.0.1:8667", Reconnect: true, AutoHeartBeat: true}

	// event
	client.OnClose = func(c *Client) {}
	client.OnOpen = func(c *Client) {}
	client.OnError = func(err error) {}
	client.OnMessage = func(c *Client, messageType int, msg []byte) {}

	// create router
	clientRouter = &Router{IgnoreCase: true}

	go client.SetRouter(clientRouter).Connect()

	client.OnSuccess = func() {
		fn()
	}
}

func TestMain(t *testing.M) {

	var sucServer = false
	var sucClient = false

	var serverFn = func() {
		sucServer = true
	}

	var clientFn = func() {
		sucClient = true
	}

	initServer(serverFn)

	initClient(clientFn)

	go func() {
		for {
			if sucServer && sucClient {
				t.Run()
				break
			}
		}

	}()

	<-stop

	_ = client.Close()
	_ = webSocketServer.Shutdown()
}

func Test_Client_Async(t *testing.T) {
	stream, err := client.AsyncJson(socket.JsonPackage{
		Event: "/hello/world",
		Data:  strings.Repeat("hello world!", 1),
	})

	kitty.AssertEqual(t, err == nil, err)

	kitty.AssertEqual(t, stream != nil, "stream is nil")

	kitty.AssertEqual(t, string(stream.Message) == `"i am server"`)
}

func Test_Client(t *testing.T) {
	clientRouter.Group("/hello").Handler(func(handler *RouteHandler) {
		handler.Route("/world").Handler(func(c *Client, stream *socket.Stream) error {
			defer mux.Add(-1)
			kitty.AssertEqual(t, string(stream.Message) == `"i am server"`, "stream is nil")
			return nil
		})
	})

	_ = client.Json(socket.JsonPackage{
		Event: "/hello/world",
		Data:  strings.Repeat("hello world!", 1),
	})

	mux.Add(1)

	mux.Wait()
}

func Test_Shutdown(t *testing.T) {
	shutdown()
}
