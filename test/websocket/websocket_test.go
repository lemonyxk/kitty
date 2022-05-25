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
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/example/protobuf"
	kitty2 "github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket/async"
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

func shutdown() {
	stop <- true
}

var webSocketServer *server.Server

var webSocketServerRouter *router.Router[*socket.Stream[server.Conn]]

var webSocketClient *client.Client

var clientRouter *router.Router[*socket.Stream[client.Conn]]

var addr = "127.0.0.1:8669"

var fd int64 = 0

func initServer() {

	var ready = make(chan bool)

	// create server
	webSocketServer = kitty.NewWebSocketServer(addr)

	// event
	webSocketServer.OnOpen = func(conn server.Conn) { fd++ }
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
		next(&socket.Stream[server.Conn]{Conn: conn, Event: route, Data: []byte(data)})
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

	var wsRouter = webSocketServerRouter.Create()
	wsRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		return stream.Conn.JsonEmit(stream.Event, res)
	})

	wsRouter.Route("/Emit").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Conn.Emit(stream.Event, stream.Data)
	})

	wsRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res awesomepackage.AwesomeMessage
		_ = proto.Unmarshal(stream.Data, &res)
		return stream.Conn.ProtoBufEmit(stream.Event, &res)
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	webSocketServer.OnSuccess = func() {
		ready <- true
	}

	<-ready
}

func initClient() {
	// do not use chan in on success event
	// because when the client is closed and reconnected, the on success event will be called again.
	// or you can use a global variable to record the client status.
	// just for test.
	var ready = make(chan bool)
	var isRun = false

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
		next(&socket.Stream[client.Conn]{Conn: c, Event: route, Data: []byte(data)})
	}

	// create router
	clientRouter = kitty.NewWebSocketClientRouter()

	clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
		return stream.Conn.Client().JsonEmit(stream.Event, string(stream.Data))
	})

	go webSocketClient.SetRouter(clientRouter).Connect()

	webSocketClient.OnSuccess = func() {
		if isRun {
			return
		}
		ready <- true
	}

	<-ready

	isRun = true
}

func TestMain(t *testing.M) {

	initServer()

	initClient()

	go func() {
		<-stop

		_ = webSocketClient.Close()
		_ = webSocketServer.Shutdown()
	}()

	t.Run()
}

func Test_WS_Client_Async(t *testing.T) {

	var asyncClient = async.NewClient[client.Conn](webSocketClient)

	var wait = sync.WaitGroup{}

	wait.Add(100)

	for i := 0; i < 100; i++ {
		var index = i
		go func() {
			stream, err := asyncClient.JsonEmit("/asyncClient", fmt.Sprintf("%d", index))

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

	var mux sync.WaitGroup

	mux.Add(count)

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == "i am server", "stream is nil")
			return nil
		})
	})

	for i := 0; i < count; i++ {
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

func Test_WS_JsonEmit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		assert.True(t, res["name"] == "kitty", res)
		assert.True(t, res["age"] == "18", res)
		mux.Done()
		return nil
	})

	var err = webSocketClient.JsonEmit("/JsonFormat", kitty2.M{"name": "kitty", "age": "18"})

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_WS_Emit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/Emit").Handler(func(stream *socket.Stream[client.Conn]) error {
		assert.True(t, string(stream.Data) == `{"name":"kitty","age":18}`, string(stream.Data))
		mux.Done()
		return nil
	})

	var err = webSocketClient.Emit("/Emit", []byte(`{"name":"kitty","age":18}`))

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_WS_ProtobufEmit(t *testing.T) {
	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res awesomepackage.AwesomeMessage
		_ = proto.Unmarshal(stream.Data, &res)
		assert.True(t, res.AwesomeField == "1", res)
		assert.True(t, res.AwesomeKey == "2", res)
		mux.Done()
		return nil
	})

	var buf = awesomepackage.AwesomeMessage{
		AwesomeField: "1",
		AwesomeKey:   "2",
	}

	var err = webSocketClient.ProtoBufEmit("/ProtoBufEmit", &buf)

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_WS_Server_Async(t *testing.T) {

	var asyncServer = async.NewServer[server.Conn](webSocketServer)

	var wait = sync.WaitGroup{}

	wait.Add(100)

	for i := 0; i < 100; i++ {
		var index = i
		go func() {
			stream, err := asyncServer.JsonEmit(fd, "/asyncServer", index)

			assert.True(t, err == nil, err)

			assert.True(t, stream != nil, "stream is nil")

			assert.True(t, string(stream.Data) == fmt.Sprintf("\"%d\"", index), string(stream.Data))

			wait.Done()
		}()
	}

	wait.Wait()
}

func Test_WS_Ping_Pong(t *testing.T) {

	var pingCount int32 = 0
	var pongCount int32 = 0

	webSocketServer.HeartBeatTimeout = time.Second * 3
	webSocketServer.PingHandler = func(conn server.Conn) func(data string) error {
		return func(data string) error {
			atomic.AddInt32(&pingCount, 1)
			var err error
			var t = time.Now()
			conn.SetLastPing(t)
			if webSocketServer.HeartBeatTimeout != 0 {
				err = conn.Conn().SetReadDeadline(t.Add(webSocketServer.HeartBeatTimeout))
			}
			err = conn.Pong()
			return err
		}
	}

	webSocketClient.ReconnectInterval = time.Millisecond * 1

	webSocketClient.HeartBeatTimeout = time.Second * 3
	webSocketClient.HeartBeatInterval = time.Millisecond * 10
	webSocketClient.PongHandler = func(conn client.Conn) func(data string) error {
		return func(data string) error {
			atomic.AddInt32(&pongCount, 1)
			var t = time.Now()
			conn.SetLastPong(t)
			if webSocketClient.HeartBeatTimeout != 0 {
				return conn.Conn().SetReadDeadline(t.Add(webSocketClient.HeartBeatTimeout))
			}
			return nil
		}
	}

	// reconnect make the config effective
	_ = webSocketClient.Close()

	var ready = make(chan bool)

	time.AfterFunc(time.Millisecond*1234, func() {
		ready <- true
	})

	<-ready

	assert.True(t, pingCount == pongCount, fmt.Sprintf("pingCount:%d, pongCount:%d", pingCount, pongCount))

	assert.True(t, pingCount == 123, pingCount)
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
