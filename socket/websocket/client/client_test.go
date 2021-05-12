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
	"time"

	"github.com/json-iterator/go"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/websocket/server"
)

type JsonPack struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
	ID    int64       `json:"id"`
}

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var webSocketServer *server.Server

var webSocketServerRouter *server.Router

var client *Client

var clientRouter *Router

var host = "127.0.0.1:8669"

func initServer(fn func()) {

	// create server
	webSocketServer = &server.Server{Host: host, Path: "/"}

	// event
	webSocketServer.OnOpen = func(conn *server.Conn) {}
	webSocketServer.OnClose = func(conn *server.Conn) {}
	webSocketServer.OnError = func(err error) {}
	webSocketServer.OnMessage = func(conn *server.Conn, msg []byte) {}

	// middleware
	webSocketServer.Use(func(next server.Middle) server.Middle {
		return func(conn *server.Conn, stream *socket.Stream) {
			next(conn, stream)
		}
	})

	// handle unknown proto
	webSocketServer.OnUnknown = func(conn *server.Conn, message []byte, next server.Middle) {
		var j = jsoniter.Get(message)
		var id = j.Get("id").ToInt64()
		var route = j.Get("event").ToString()
		var data = j.Get("data").ToString()
		if route == "" {
			return
		}
		next(conn, &socket.Stream{Pack: socket.Pack{Event: route, Data: []byte(data), ID: id}})
	}

	// create router
	webSocketServerRouter = &server.Router{IgnoreCase: true}

	// set group route
	webSocketServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Conn, stream *socket.Stream) error {
			return ServerJson(conn, JsonPack{
				Event: "/hello/world",
				Data:  "i am server",
				ID:    stream.ID,
			})
		})
	})

	webSocketServerRouter.Route("/async").Handler(func(conn *server.Conn, stream *socket.Stream) error {
		return ServerJson(conn, JsonPack{
			Event: "/async",
			Data:  "async test",
			ID:    stream.ID,
		})
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	webSocketServer.OnSuccess = func() {
		fn()
	}
}

func initClient(fn func()) {
	// create client
	client = &Client{Scheme: "ws", Host: host, Reconnect: true, AutoHeartBeat: true}

	// event
	client.OnClose = func(c *Client) {}
	client.OnOpen = func(c *Client) {}
	client.OnError = func(err error) {}
	client.OnMessage = func(c *Client, messageType int, msg []byte) {}

	// handle unknown proto
	client.OnUnknown = func(conn *Client, message []byte, next Middle) {
		var j = jsoniter.Get(message)
		var id = j.Get("id").ToInt64()
		var route = j.Get("event").ToString()
		var data = j.Get("data").ToString()
		if route == "" {
			return
		}
		next(conn, &socket.Stream{Pack: socket.Pack{Event: route, Data: []byte(data), ID: id}})
	}

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
		<-stop

		_ = client.Close()
		_ = webSocketServer.Shutdown()
	}()

	for {
		if sucServer && sucClient {
			t.Run()
			break
		}
	}

}

func Test_Client_Async(t *testing.T) {
	stream, err := client.Async().JsonEmit(socket.JsonPack{
		Event: "/async",
		Data:  strings.Repeat("hello world!", 1),
	})

	kitty.AssertEqual(t, err == nil, err)

	kitty.AssertEqual(t, stream != nil, "stream is nil")

	kitty.AssertEqual(t, string(stream.Data) == "async test", "stream is nil")
}

func Test_Client(t *testing.T) {

	var id int64 = 123456

	var count = 10000

	clientRouter.Group("/hello").Handler(func(handler *RouteHandler) {
		handler.Route("/world").Handler(func(c *Client, stream *socket.Stream) error {
			defer mux.Add(-1)
			kitty.AssertEqual(t, string(stream.Data) == "i am server", "stream is nil")
			kitty.AssertEqual(t, stream.ID == id, "id not match", stream.ID)
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = ClientJson(client, JsonPack{
			Event: "/hello/world",
			Data:  strings.Repeat("hello world!", 1),
			ID:    id,
		})
	}

	go func() {
		<-time.After(3 * time.Second)
		mux.Done()
		t.Fatal("timeout")
	}()

	mux.Wait()
}

func Test_Shutdown(t *testing.T) {
	shutdown()
}

func ClientJson(c *Client, pack JsonPack) error {
	data, err := jsoniter.Marshal(pack)
	if err != nil {
		return err
	}
	return c.Push(data)
}

func ServerJson(conn *server.Conn, pack JsonPack) error {
	data, err := jsoniter.Marshal(pack)
	if err != nil {
		return err
	}
	return conn.Server.Push(conn.FD, int(socket.BinData), data)
}
