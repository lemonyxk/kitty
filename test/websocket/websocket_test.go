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
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	async2 "github.com/lemonyxk/kitty/v2/socket/async"
	"github.com/lemonyxk/kitty/v2/socket/websocket/client"
	"github.com/stretchr/testify/assert"

	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/websocket/server"
)

type JsonPack struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var webSocketServer *server.Server

var webSocketServerRouter *router.Router[*socket.Stream[server.Conn]]

var webSocketClient *client.Client

var clientRouter *router.Router[*socket.Stream[client.Conn]]

var addr = "127.0.0.1:8669"

func initServer(fn func()) {

	// create server
	webSocketServer = kitty.NewWebSocketServer(addr)

	// event
	webSocketServer.OnOpen = func(conn server.Conn) {}
	webSocketServer.OnClose = func(conn server.Conn) {}
	webSocketServer.OnError = func(err error) {}
	webSocketServer.OnMessage = func(conn server.Conn, msg []byte) {}

	// middleware
	webSocketServer.Use(func(next server.Middle) server.Middle {
		return func(stream *socket.Stream[server.Conn]) {
			next(stream)
		}
	})

	// handle unknown proto
	webSocketServer.OnUnknown = func(conn server.Conn, message []byte, next server.Middle) {
		var j = jsoniter.Get(message)
		var route = j.Get("event").ToString()
		var data = j.Get("data").ToString()
		if route == "" {
			return
		}
		next(&socket.Stream[server.Conn]{Conn: conn, Pack: socket.Pack{Event: route, Data: []byte(data)}})
	}

	// create router
	webSocketServerRouter = kitty.NewWebSocketServerRouter()

	// set group route
	webSocketServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return ServerJson(stream.Conn, JsonPack{
				Event: "/hello/world",
				Data:  "i am server",
			})
		})
	})

	webSocketServerRouter.Route("/asyncClient").Handler(func(stream *socket.Stream[server.Conn]) error {
		return ServerJson(stream.Conn, JsonPack{
			Event: "/asyncClient",
			Data:  string(stream.Data),
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
	webSocketClient.OnClose = func(c client.Conn) {}
	webSocketClient.OnOpen = func(c client.Conn) {}
	webSocketClient.OnError = func(err error) {}
	webSocketClient.OnMessage = func(c client.Conn, messageType int, msg []byte) {}

	// handle unknown proto
	webSocketClient.OnUnknown = func(c client.Conn, message []byte, next client.Middle) {
		var j = jsoniter.Get(message)
		var route = j.Get("event").ToString()
		var data = j.Get("data").ToString()
		if route == "" {
			return
		}
		next(&socket.Stream[client.Conn]{Conn: c, Pack: socket.Pack{Event: route, Data: []byte(data)}})
	}

	// create router
	clientRouter = kitty.NewWebSocketClientRouter()

	clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
		return stream.Conn.Client().JsonEmit(socket.JsonPack{
			Event: "/asyncServer",
			Data:  string(stream.Data),
		})
	})

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

	var asyncClient = async2.NewClient[client.Conn](webSocketClient)

	var wait = sync.WaitGroup{}

	wait.Add(100)

	for i := 0; i < 100; i++ {
		var index = i
		go func() {
			stream, err := asyncClient.JsonEmit(socket.JsonPack{
				Event: "/asyncClient",
				Data:  fmt.Sprintf("%d", index),
			})

			assert.True(t, err == nil, err)

			assert.True(t, stream != nil, "stream is nil")

			assert.True(t, string(stream.Data) == fmt.Sprintf("\"%d\"", index), "stream is nil")

			wait.Done()
		}()
	}

	wait.Wait()
}

func Test_WS_Client(t *testing.T) {

	var count = 10000

	var flag = true

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == "i am server", "stream is nil")
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = ClientJson(webSocketClient, JsonPack{
			Event: "/hello/world",
			Data:  strings.Repeat("hello world!", 1),
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

func Test_TCP_Server_Async(t *testing.T) {

	var asyncServer = async2.NewServer[server.Conn](webSocketServer)

	var wait = sync.WaitGroup{}

	wait.Add(100)

	for i := 0; i < 100; i++ {
		var index = i
		go func() {
			stream, err := asyncServer.JsonEmit(1, socket.JsonPack{
				Event: "/asyncServer",
				Data:  index,
			})

			assert.True(t, err == nil, err)

			assert.True(t, stream != nil, "stream is nil")

			assert.True(t, string(stream.Data) == fmt.Sprintf("\"%d\"", index), "stream is nil")

			wait.Done()
		}()
	}

	wait.Wait()
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

func ServerJson(conn server.Conn, pack JsonPack) error {
	data, err := jsoniter.Marshal(pack)
	if err != nil {
		return err
	}
	return conn.Server().Push(conn.FD(), data)
}
