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
	"github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lemonyxk/kitty"
	hello "github.com/lemonyxk/kitty/example/protobuf"
	kitty2 "github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/websocket/client"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/websocket/server"
)

type JsonPack struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

var stop = make(chan bool)

func shutdown() {
	stop <- true
}

var webSocketServer *server.Server[any]

var webSocketServerRouter *router.Router[*socket.Stream[server.Conn], any]

var webSocketClient *client.Client[any]

var clientRouter *router.Router[*socket.Stream[client.Conn], any]

var addr = "127.0.0.1:8669"

var fd int64 = 1

func initServer() {

	var ready = make(chan bool)

	// create server
	webSocketServer = kitty.NewWebSocketServer[any](addr)
	webSocketServer.HeartBeatTimeout = 5 * time.Second

	// event
	webSocketServer.OnOpen = func(conn server.Conn) {}
	webSocketServer.OnClose = func(conn server.Conn) {}
	webSocketServer.OnError = func(stream *socket.Stream[server.Conn], err error) {}
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
		var messageID = j.Get("message_id").ToUint64()
		var code = j.Get("code").ToUint32()

		if route == "" {
			return
		}
		next(socket.NewStream(conn, 0, 0, code, messageID, []byte(route), []byte(data)))
	}

	// create router
	webSocketServerRouter = kitty.NewWebSocketServerRouter[any]()

	// set group route
	webSocketServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return stream.JsonEmit("/hello/world", "i am server")
		})
	})

	webSocketServerRouter.Route("/unknown").Handler(func(stream *socket.Stream[server.Conn]) error {
		return ServerJson(stream, JsonPack{
			Event: stream.Event(),
			Data:  string(stream.Data()),
		})
	})

	webSocketServerRouter.Route("/asyncClient").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Emit(stream.Event(), stream.Data())
	})

	var wsRouter = webSocketServerRouter.Create()
	wsRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res kitty2.M
		_ = json.Unmarshal(stream.Data(), &res)
		return stream.JsonEmit(stream.Event(), res)
	})

	wsRouter.Route("/Emit").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Emit(stream.Event(), stream.Data())
	})

	wsRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res hello.AwesomeMessage
		_ = proto.Unmarshal(stream.Data(), &res)
		return stream.ProtoBufEmit(stream.Event(), &res)
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
	webSocketClient = kitty.NewWebSocketClient[any]("ws://" + addr)
	webSocketClient.ReconnectInterval = time.Second
	webSocketClient.HeartBeatInterval = time.Second

	// event
	webSocketClient.OnClose = func(c client.Conn) {}
	webSocketClient.OnOpen = func(c client.Conn) {}
	webSocketClient.OnError = func(stream *socket.Stream[client.Conn], err error) {}
	webSocketClient.OnMessage = func(c client.Conn, messageType int, msg []byte) {}

	// handle unknown proto
	webSocketClient.OnUnknown = func(c client.Conn, message []byte, next client.Middle) {
		var j = jsoniter.Get(message)
		var route = j.Get("event").ToString()
		var data = j.Get("data").ToString()
		var messageID = j.Get("message_id").ToUint64()
		var code = j.Get("code").ToUint32()
		if route == "" {
			return
		}
		next(socket.NewStream(c, 0, 0, code, messageID, []byte(route), []byte(data)))
	}

	// create router
	clientRouter = kitty.NewWebSocketClientRouter[any]()

	clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
		return stream.JsonEmit(stream.Event(), string(stream.Data()))
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

func Test_WS_Client(t *testing.T) {

	var count = 100000

	var flag = true

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var total uint64 = 0
	var messageIDTotal uint64 = 0
	var countTotal uint64 = 0

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn], any]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			if atomic.AddUint64(&countTotal, 1) == uint64(count) {
				mux.Done()
			}
			messageIDTotal += stream.MessageID()
			assert.True(t, string(stream.Data()) == `"i am server"`, string(stream.Data()))
			return nil
		})
	})

	go func() {
		<-time.After(100 * time.Second)
		mux.Done()
		flag = false
	}()

	for i := 0; i < count; i++ {
		total += uint64(i + 1)
		go func() {
			_ = webSocketClient.Sender().JsonEmit("/hello/world", strings.Repeat("hello world!", 1))
		}()
	}

	mux.Wait()

	assert.True(t, total == messageIDTotal, "message id not equal", total, messageIDTotal, countTotal)

	if !flag {
		t.Fatal("timeout")
	}
}

func Test_WS_Client_Async(t *testing.T) {

	var asyncClient = socket.NewAsyncClient[client.Conn, any](webSocketClient)

	var wait = sync.WaitGroup{}

	var random = rand.Intn(1300) + 100000

	wait.Add(random)

	var count int32 = 0

	for i := 0; i < random; i++ {
		var index = i
		go func() {
			var str = fmt.Sprintf("%d", index)

			var stream, err = asyncClient.Emit("/asyncClient", []byte(str))

			atomic.AddInt32(&count, 1)

			assert.True(t, err == nil, err)
			assert.True(t, stream != nil, "stream is nil")
			assert.True(t, string(stream.Data()) == str, fmt.Sprintf("`%+v` not equal `%+v`", string(stream.Data()), str))

			wait.Done()
		}()
	}

	wait.Wait()

	assert.True(t, int(count) == random, "count not equal", count, random)
}

func Test_WS_JsonEmit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res kitty2.M
		_ = json.Unmarshal(stream.Data(), &res)
		assert.True(t, res["name"] == "kitty", res)
		assert.True(t, res["age"] == "18", res)
		mux.Done()
		return nil
	})

	var err = webSocketClient.Sender().JsonEmit("/JsonFormat", kitty2.M{"name": "kitty", "age": "18"})

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_WS_Emit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/Emit").Handler(func(stream *socket.Stream[client.Conn]) error {
		assert.True(t, string(stream.Data()) == `{"name":"kitty","age":18}`, string(stream.Data()))
		mux.Done()
		return nil
	})

	var err = webSocketClient.Sender().Emit("/Emit", []byte(`{"name":"kitty","age":18}`))

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_WS_ProtobufEmit(t *testing.T) {
	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res hello.AwesomeMessage
		_ = proto.Unmarshal(stream.Data(), &res)
		assert.True(t, res.AwesomeField == "1", res.String())
		assert.True(t, res.AwesomeKey == "2", res.String())
		mux.Done()
		return nil
	})

	var buf = hello.AwesomeMessage{
		AwesomeField: "1",
		AwesomeKey:   "2",
	}

	var err = webSocketClient.Sender().ProtoBufEmit("/ProtoBufEmit", &buf)

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_WS_Server_Async(t *testing.T) {

	var asyncServer = socket.NewAsyncServer[server.Conn, any](webSocketServer)

	var wait = sync.WaitGroup{}

	var random = rand.Intn(1001) + 100000

	wait.Add(random)

	for i := 0; i < random; i++ {
		var index = i
		go func() {
			var sender, err = asyncServer.Sender(fd)

			assert.True(t, err == nil, err)

			stream, err := sender.JsonEmit("/asyncServer", index)

			assert.True(t, err == nil, err)

			assert.True(t, stream != nil, "stream is nil")

			assert.True(t, string(stream.Data()) == fmt.Sprintf("\"%d\"", index), string(stream.Data()))

			wait.Done()
		}()
	}

	wait.Wait()
}

func Test_WS_Unknown(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var wsRouter = clientRouter.Create()

	wsRouter.Route("/unknown").Handler(func(stream *socket.Stream[client.Conn]) error {
		assert.True(t, string(stream.Data()) == `{"name":"unknown","age":18}`, string(stream.Data()))
		mux.Done()
		return nil
	})

	var err = ClientJson(webSocketClient, JsonPack{
		Event: "/unknown",
		Data:  `{"name":"unknown","age":18}`,
	})
	assert.True(t, err == nil, err)

	mux.Wait()
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
				err = conn.SetDeadline(t.Add(webSocketServer.HeartBeatTimeout))
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
				return conn.SetDeadline(t.Add(webSocketClient.HeartBeatTimeout))
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
}

func Test_WS_Multi_Client(t *testing.T) {

	var count int32 = 0

	for i := 0; i < 100; i++ {
		go func() {
			// create client
			var wClient = kitty.NewWebSocketClient[any]("ws://" + addr)
			wClient.ReconnectInterval = time.Second
			wClient.HeartBeatInterval = time.Millisecond * 1000 / 60

			// event
			wClient.OnClose = func(c client.Conn) {}
			wClient.OnOpen = func(c client.Conn) {}
			wClient.OnError = func(stream *socket.Stream[client.Conn], err error) {}
			wClient.OnMessage = func(c client.Conn, messageType int, msg []byte) {}

			// handle unknown proto
			wClient.OnUnknown = func(c client.Conn, message []byte, next client.Middle) {
				var j = jsoniter.Get(message)
				var route = j.Get("event").ToString()
				var data = j.Get("data").ToString()
				var messageID = j.Get("message_id").ToUint64()
				var code = j.Get("code").ToUint32()
				if route == "" {
					return
				}
				next(socket.NewStream(c, 0, 0, code, messageID, []byte(route), []byte(data)))
			}

			// create router
			var clientRouter = kitty.NewWebSocketClientRouter[any]()

			clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
				return stream.JsonEmit(stream.Event(), string(stream.Data()))
			})

			go wClient.SetRouter(clientRouter).Connect()

			wClient.OnSuccess = func() {
				atomic.AddInt32(&count, 1)
			}
		}()
	}

	time.Sleep(time.Second * 3)
	assert.True(t, count == 100, fmt.Sprintf("count:%d", count))
}

func Test_WS_Shutdown(t *testing.T) {
	shutdown()
}

func ClientJson(c *client.Client[any], pack JsonPack) error {
	data, err := json.Marshal(pack)
	if err != nil {
		return err
	}
	return c.Conn().Push(data)
}

func ServerJson(stream *socket.Stream[server.Conn], pack JsonPack) error {
	data, err := json.Marshal(pack)
	if err != nil {
		return err
	}
	return stream.Conn().Push(data)
}
