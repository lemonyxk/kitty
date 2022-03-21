/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2020-09-18 16:40
**/

package websocket

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/socket/websocket/client"
	"github.com/stretchr/testify/assert"

	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/websocket/server"
)

type JsonPack struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
	ID    int64  `json:"id"`
}

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var webSocketServer *server.Server

var webSocketServerRouter *server.Router

var webSocketClient *client.Client

var clientRouter *client.Router

var addr = "127.0.0.1:8669"

func initServer(fn func()) {

	// create server
	webSocketServer = kitty.NewWebSocketServer(addr)

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
	webSocketServerRouter = kitty.NewWebSocketServerRouter()

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
	webSocketClient = kitty.NewWebSocketClient("ws://" + addr)
	webSocketClient.ReconnectInterval = time.Second
	webSocketClient.HeartBeatInterval = time.Second

	// event
	webSocketClient.OnClose = func(c *client.Client) {}
	webSocketClient.OnOpen = func(c *client.Client) {}
	webSocketClient.OnError = func(err error) {}
	webSocketClient.OnMessage = func(c *client.Client, messageType int, msg []byte) {}

	// handle unknown proto
	webSocketClient.OnUnknown = func(conn *client.Client, message []byte, next client.Middle) {
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
	clientRouter = kitty.NewWebSocketClientRouter()

	go webSocketClient.SetRouter(clientRouter).Connect()

	webSocketClient.OnSuccess = func() {
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

		_ = webSocketClient.Close()
		_ = webSocketServer.Shutdown()
	}()

	for {
		if sucServer && sucClient {
			t.Run()
			break
		}
	}

}

func Test_WS_Client_Async(t *testing.T) {
	stream, err := webSocketClient.Async().JsonEmit(socket.JsonPack{
		Event: "/async",
		Data:  strings.Repeat("hello world!", 1),
	})

	assert.True(t, err == nil, err)

	assert.True(t, stream != nil, "stream is nil")

	assert.True(t, string(stream.Data) == "async test", "stream is nil")
}

func Test_WS_Client(t *testing.T) {

	var id int64 = 123456

	var count = 10000

	var flag = true

	clientRouter.Group("/hello").Handler(func(handler *client.RouteHandler) {
		handler.Route("/world").Handler(func(c *client.Client, stream *socket.Stream) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == "i am server", "stream is nil")
			assert.True(t, stream.ID == id, "id not match", stream.ID)
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = ClientJson(webSocketClient, JsonPack{
			Event: "/hello/world",
			Data:  strings.Repeat("hello world!", 1),
			ID:    id,
		})
	}

	go func() {
		<-time.After(3 * time.Second)
		mux.Done()
		flag = false
	}()

	mux.Wait()

	if !flag {
		t.Fatal("timeout")
	}
}

func Test_WS_Shutdown(t *testing.T) {
	shutdown()
}

func ClientJson(c *client.Client, pack JsonPack) error {
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
	return conn.Server.Push(conn.FD, data)
}
