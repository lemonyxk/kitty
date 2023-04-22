/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2020-09-18 16:40
**/

package tcp

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty"
	hello "github.com/lemonyxk/kitty/example/protobuf"
	kitty2 "github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/async"
	"github.com/lemonyxk/kitty/socket/tcp/client"
	"github.com/lemonyxk/kitty/socket/tcp/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var stop = make(chan bool)

func shutdown() {
	stop <- true
}

var tcpServer *server.Server

var tcpServerRouter *router.Router[*socket.Stream[server.Conn]]

var tcpClient *client.Client

var clientRouter *router.Router[*socket.Stream[client.Conn]]

var addr = "127.0.0.1:8667"

var fd int64 = 1

func initServer() {

	var ready = make(chan bool)

	// create server
	tcpServer = kitty.NewTcpServer(addr)
	// tcpServer.HeartBeatTimeout = 5 * time.Second

	// event
	tcpServer.OnOpen = func(conn server.Conn) {}
	tcpServer.OnClose = func(conn server.Conn) {}
	tcpServer.OnError = func(stream *socket.Stream[server.Conn], err error) {}
	tcpServer.OnMessage = func(conn server.Conn, msg []byte) {}

	// middleware
	tcpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *socket.Stream[server.Conn]) {
			next(stream)
		}
	})

	// create router
	tcpServerRouter = &router.Router[*socket.Stream[server.Conn]]{StrictMode: true}

	// set group route
	tcpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return stream.JsonEmit(stream.Event, "i am server")
		})
	})

	tcpServerRouter.Route("/asyncClient").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Emit(stream.Event, stream.Data)
	})

	var tcpRouter = tcpServerRouter.Create()
	tcpRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		return stream.JsonEmit(stream.Event, res)
	})

	tcpRouter.Route("/Emit").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Emit(stream.Event, stream.Data)
	})

	tcpRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res hello.AwesomeMessage
		_ = proto.Unmarshal(stream.Data, &res)
		return stream.ProtoBufEmit(stream.Event, &res)
	})

	go tcpServer.SetRouter(tcpServerRouter).Start()

	tcpServer.OnSuccess = func() {
		ready <- true
	}

	<-ready
}

func initClient() {

	var ready = make(chan bool)

	var isRun = false

	// create client
	tcpClient = kitty.NewTcpClient(addr)
	tcpClient.ReconnectInterval = time.Second
	// tcpClient.HeartBeatInterval = time.Second

	// event
	tcpClient.OnClose = func(c client.Conn) {}
	tcpClient.OnOpen = func(c client.Conn) {}
	tcpClient.OnError = func(stream *socket.Stream[client.Conn], err error) {}
	tcpClient.OnMessage = func(client client.Conn, msg []byte) {}

	// create router
	clientRouter = &router.Router[*socket.Stream[client.Conn]]{StrictMode: true}

	clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
		return stream.JsonEmit(stream.Event, string(stream.Data))
	})

	go tcpClient.SetRouter(clientRouter).Connect()

	tcpClient.OnSuccess = func() {
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

		_ = tcpClient.Close()
		_ = tcpServer.Shutdown()
	}()

	t.Run()
}

func Test_TCP_Client(t *testing.T) {

	var count = 100000

	var flag = true

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var total uint64 = 0
	var messageIDTotal uint64 = 0
	var countTotal uint64 = 0

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			if atomic.AddUint64(&countTotal, 1) == uint64(count) {
				mux.Done()
			}
			messageIDTotal += stream.MessageID()
			assert.True(t, string(stream.Data) == `"i am server"`, "stream is nil")
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
			_ = tcpClient.Sender().JsonEmit("/hello/world", strings.Repeat("hello world!", 1))
		}()
	}

	mux.Wait()

	assert.True(t, total == messageIDTotal, "message id not equal", total, messageIDTotal, countTotal)

	if !flag {
		t.Fatal("timeout")
	}
}

func Test_TCP_Client_Async(t *testing.T) {

	var asyncClient = async.NewClient[client.Conn](tcpClient)

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
			assert.True(t, string(stream.Data) == str, fmt.Sprintf("`%+v` not equal `%+v`", string(stream.Data), str))

			wait.Done()
		}()
	}

	wait.Wait()

	assert.True(t, int(count) == random, "count not equal", count, random)
}

func Test_TCP_JsonEmit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var tcpRouter = clientRouter.Create()

	tcpRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		assert.True(t, res["name"] == "kitty", res)
		assert.True(t, res["age"] == "18", res)
		mux.Done()
		return nil
	})

	var err = tcpClient.Sender().JsonEmit("/JsonFormat", kitty2.M{
		"name": "kitty",
		"age":  "18",
	})

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_TCP_Emit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var tcpRouter = clientRouter.Create()

	tcpRouter.Route("/Emit").Handler(func(stream *socket.Stream[client.Conn]) error {
		assert.True(t, string(stream.Data) == `{"name":"kitty","age":18}`, string(stream.Data))
		mux.Done()
		return nil
	})

	var err = tcpClient.Sender().Emit("/Emit", []byte(`{"name":"kitty","age":18}`))

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_TCP_ProtobufEmit(t *testing.T) {
	var mux = sync.WaitGroup{}

	mux.Add(1)

	var tcpRouter = clientRouter.Create()

	tcpRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res hello.AwesomeMessage
		_ = proto.Unmarshal(stream.Data, &res)
		assert.True(t, res.AwesomeField == "1", res.String())
		assert.True(t, res.AwesomeKey == "2", res.String())
		mux.Done()
		return nil
	})

	var buf = hello.AwesomeMessage{
		AwesomeField: "1",
		AwesomeKey:   "2",
	}

	var err = tcpClient.Sender().ProtoBufEmit("/ProtoBufEmit", &buf)

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_TCP_Server_Async(t *testing.T) {

	var asyncServer = async.NewServer[server.Conn](tcpServer)

	var wait = sync.WaitGroup{}

	var random = rand.Intn(1001) + 100000

	wait.Add(random)

	for i := 0; i < random; i++ {
		var index = i
		go func() {

			stream, err := asyncServer.JsonEmit(fd, "/asyncServer", index)

			assert.True(t, err == nil, err)

			assert.True(t, stream != nil, "stream is nil")

			assert.True(t, string(stream.Data) == fmt.Sprintf("\"%d\"", index), "stream is nil")

			wait.Done()
		}()
	}

	wait.Wait()
}

func Test_TCP_Ping_Pong(t *testing.T) {

	var pingCount int32 = 0
	var pongCount int32 = 0

	tcpServer.HeartBeatTimeout = time.Second * 3
	tcpServer.PingHandler = func(conn server.Conn) func(data string) error {
		return func(data string) error {
			atomic.AddInt32(&pingCount, 1)
			var err error
			var t = time.Now()
			conn.SetLastPing(t)
			if tcpServer.HeartBeatTimeout != 0 {
				err = conn.SetDeadline(t.Add(tcpServer.HeartBeatTimeout))
			}
			err = conn.Pong()
			return err
		}
	}

	tcpClient.ReconnectInterval = time.Millisecond * 1

	tcpClient.HeartBeatTimeout = time.Second * 3
	tcpClient.HeartBeatInterval = time.Millisecond * 10
	tcpClient.PongHandler = func(conn client.Conn) func(data string) error {
		return func(data string) error {
			atomic.AddInt32(&pongCount, 1)
			var t = time.Now()
			conn.SetLastPong(t)
			if tcpClient.HeartBeatTimeout != 0 {
				return conn.SetDeadline(t.Add(tcpClient.HeartBeatTimeout))
			}
			return nil
		}
	}

	// reconnect make the config effective
	_ = tcpClient.Close()

	var ready = make(chan bool)

	time.AfterFunc(time.Millisecond*1234, func() {
		ready <- true
	})

	<-ready

	assert.True(t, pingCount == pongCount, fmt.Sprintf("pingCount:%d, pongCount:%d", pingCount, pongCount))
}

func Test_TCP_Multi_Client(t *testing.T) {

	var count int32 = 0

	for i := 0; i < 100; i++ {
		go func() {
			// create client
			var tClient = kitty.NewTcpClient(addr)
			tClient.ReconnectInterval = time.Second
			tClient.HeartBeatInterval = time.Millisecond * 1000 / 60

			// event
			tClient.OnClose = func(c client.Conn) {}
			tClient.OnOpen = func(c client.Conn) {}
			tClient.OnError = func(stream *socket.Stream[client.Conn], err error) {}
			tClient.OnMessage = func(client client.Conn, msg []byte) {}

			// create router
			var clientRouter = &router.Router[*socket.Stream[client.Conn]]{StrictMode: true}

			clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
				return stream.JsonEmit(stream.Event, string(stream.Data))
			})

			go tClient.SetRouter(clientRouter).Connect()

			tClient.OnSuccess = func() {
				atomic.AddInt32(&count, 1)
			}
		}()
	}

	time.Sleep(time.Second * 3)
	assert.True(t, count == 100, fmt.Sprintf("count:%d", count))
}

func Test_TCP_Shutdown(t *testing.T) {
	shutdown()
}
