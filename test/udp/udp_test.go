/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2020-09-18 16:40
**/

package udp

import (
	"fmt"
	"math/rand"
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
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/async"
	"github.com/lemonyxk/kitty/v2/socket/udp/client"
	"github.com/lemonyxk/kitty/v2/socket/udp/server"
	"github.com/stretchr/testify/assert"
)

var stop = make(chan bool)

func shutdown() {
	stop <- true
}

var udpServer *server.Server

var udpServerRouter *router.Router[*socket.Stream[server.Conn]]

var udpClient *client.Client

var clientRouter *router.Router[*socket.Stream[client.Conn]]

var addr = "127.0.0.1:8668"

var fd int64 = 0

func initServer() {

	var ready = make(chan bool)

	// create server
	udpServer = kitty.NewUdpServer(addr)

	// event
	udpServer.OnOpen = func(conn server.Conn) { fd++ }
	udpServer.OnClose = func(conn server.Conn) {}
	udpServer.OnError = func(err error) {}
	udpServer.OnMessage = func(conn server.Conn, msg []byte) {}

	// middleware
	udpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *socket.Stream[server.Conn]) {
			next(stream)
		}
	})

	// create router
	udpServerRouter = &router.Router[*socket.Stream[server.Conn]]{StrictMode: true}

	// set group route
	udpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return stream.JsonEmit(stream.Event, "i am server")
		})
	})

	udpServerRouter.Route("/asyncClient").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Emit(stream.Event, stream.Data)
	})

	var udpRouter = udpServerRouter.Create()
	udpRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		return stream.JsonEmit(stream.Event, res)
	})

	udpRouter.Route("/Emit").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Emit(stream.Event, stream.Data)
	})

	udpRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res awesomepackage.AwesomeMessage
		_ = proto.Unmarshal(stream.Data, &res)
		return stream.ProtoBufEmit(stream.Event, &res)
	})

	go udpServer.SetRouter(udpServerRouter).Start()

	udpServer.OnSuccess = func() {
		ready <- true
	}

	<-ready
}

func initClient() {

	var ready = make(chan bool)
	var isRun = false

	// create client
	udpClient = kitty.NewUdpClient(addr)
	udpClient.ReconnectInterval = time.Second
	udpClient.HeartBeatInterval = time.Second

	// event
	udpClient.OnClose = func(c client.Conn) {}
	udpClient.OnOpen = func(c client.Conn) {}
	udpClient.OnError = func(err error) {}
	udpClient.OnMessage = func(client client.Conn, msg []byte) {}

	// create router
	clientRouter = &router.Router[*socket.Stream[client.Conn]]{StrictMode: true}

	clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
		return stream.JsonEmit(stream.Event, string(stream.Data))
	})

	go udpClient.SetRouter(clientRouter).Connect()

	udpClient.OnSuccess = func() {
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
		_ = udpClient.Close()
		_ = udpServer.Shutdown()

	}()

	t.Run()
}

func Test_UDP_Client(t *testing.T) {

	var count = 10000

	var flag = true

	var mux = new(sync.WaitGroup)

	mux.Add(count)

	var total int64 = 0
	var messageIDTotal int64 = 0

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			defer mux.Add(-1)
			messageIDTotal += stream.MessageID()
			assert.True(t, string(stream.Data) == `"i am server"`, "stream is nil")
			return nil
		})
	})

	for i := 0; i < count; i++ {
		total += int64(i + 1)
		go func() {
			_ = udpClient.JsonEmit("/hello/world", strings.Repeat("hello world!", 1))
		}()
	}

	go func() {
		<-time.After(100 * time.Second)
		mux.Done()
		flag = false
	}()

	mux.Wait()

	assert.True(t, total == messageIDTotal, "message id not equal", total, messageIDTotal)

	if !flag {
		t.Fatal("timeout")
	}
}

// TODO
// need test more
func Test_UDP_Client_Async(t *testing.T) {

	udpClient.DailTimeout = time.Second * 10

	var asyncClient = async.NewClient[client.Conn](udpClient)

	var wait = sync.WaitGroup{}

	var random = rand.Intn(13) + 50

	wait.Add(random)

	var count = 0

	for i := 0; i < random; i++ {
		go func(index int) {
			var str = fmt.Sprintf("%d", index)
			var stream, err = asyncClient.Emit("/asyncClient", []byte(str))
			count++

			assert.True(t, err == nil, err)

			assert.True(t, stream != nil, "stream is nil")

			assert.True(t, string(stream.Data) == str, fmt.Sprintf("`%+v` not equal `%+v`", string(stream.Data), str))

			wait.Done()

		}(i)
	}

	wait.Wait()

	assert.True(t, count == random, "count not equal", count, random)
}

func Test_UDP_JsonEmit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var udpRouter = clientRouter.Create()

	udpRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[client.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		assert.True(t, res["name"] == "kitty", res)
		assert.True(t, res["age"] == "18", res)
		mux.Done()
		return nil
	})

	var err = udpClient.JsonEmit("/JsonFormat", kitty2.M{
		"name": "kitty",
		"age":  "18",
	})

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_UDP_Emit(t *testing.T) {

	var mux = sync.WaitGroup{}

	mux.Add(1)

	var udpRouter = clientRouter.Create()

	udpRouter.Route("/Emit").Handler(func(stream *socket.Stream[client.Conn]) error {
		assert.True(t, string(stream.Data) == `{"name":"kitty","age":18}`, string(stream.Data))
		mux.Done()
		return nil
	})

	var err = udpClient.Emit("/Emit", []byte(`{"name":"kitty","age":18}`))

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_UDP_ProtobufEmit(t *testing.T) {
	var mux = sync.WaitGroup{}

	mux.Add(1)

	var udpRouter = clientRouter.Create()

	udpRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[client.Conn]) error {
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

	var err = udpClient.ProtoBufEmit("/ProtoBufEmit", &buf)

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_UDP_Server_Async(t *testing.T) {

	var asyncServer = async.NewServer[server.Conn](udpServer)

	var wait = sync.WaitGroup{}

	var random = rand.Intn(10122) + 5000

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

func Test_UDP_Ping_Pong(t *testing.T) {

	var pingCount int32 = 0
	var pongCount int32 = 0

	udpServer.HeartBeatTimeout = time.Second * 3
	udpServer.PingHandler = func(conn server.Conn) func(data string) error {
		return func(data string) error {
			atomic.AddInt32(&pingCount, 1)
			var err error
			var t = time.Now()
			conn.SetLastPing(t)
			if udpServer.HeartBeatTimeout != 0 {
				err = conn.SetReadDeadline(t.Add(udpServer.HeartBeatTimeout))
			}
			err = conn.Pong()
			return err
		}
	}

	udpClient.ReconnectInterval = time.Millisecond * 1

	udpClient.HeartBeatTimeout = time.Second * 3
	udpClient.HeartBeatInterval = time.Millisecond * 10
	udpClient.PongHandler = func(conn client.Conn) func(data string) error {
		return func(data string) error {
			atomic.AddInt32(&pongCount, 1)
			var t = time.Now()
			conn.SetLastPong(t)
			if udpClient.HeartBeatTimeout != 0 {
				return conn.SetReadDeadline(t.Add(udpClient.HeartBeatTimeout))
			}
			return nil
		}
	}

	// reconnect make the config effective
	_ = udpClient.Close()

	var ready = make(chan bool)

	time.AfterFunc(time.Millisecond*1234, func() {
		ready <- true
	})

	<-ready

	assert.True(t, pingCount == pongCount, fmt.Sprintf("pingCount:%d, pongCount:%d", pingCount, pongCount))
}

func Test_UDP_Shutdown(t *testing.T) {
	shutdown()
}
