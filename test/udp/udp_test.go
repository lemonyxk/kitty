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
	"strings"
	"sync"
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
			return stream.Conn.JsonEmit(socket.JsonPack{
				Event: "/hello/world",
				Data:  "i am server",
			})
		})
	})

	udpServerRouter.Route("/asyncClient").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Conn.JsonEmit(socket.JsonPack{
			Event: "/asyncClient",
			Data:  string(stream.Data),
		})
	})

	var udpRouter = udpServerRouter.Create()
	udpRouter.Route("/JsonFormat").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res kitty2.M
		_ = jsoniter.Unmarshal(stream.Data, &res)
		return stream.Conn.JsonEmit(socket.JsonPack{
			Event: stream.Event,
			Data:  res,
		})
	})

	udpRouter.Route("/Emit").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Conn.Emit(socket.Pack{
			Event: stream.Event,
			Data:  stream.Data,
		})
	})

	udpRouter.Route("/ProtoBufEmit").Handler(func(stream *socket.Stream[server.Conn]) error {
		var res awesomepackage.AwesomeMessage
		_ = proto.Unmarshal(stream.Data, &res)
		return stream.Conn.ProtoBufEmit(socket.ProtoBufPack{
			Event: stream.Event,
			Data:  &res,
		})
	})

	go udpServer.SetRouter(udpServerRouter).Start()

	udpServer.OnSuccess = func() {
		ready <- true
	}

	<-ready
}

func initClient() {

	var ready = make(chan bool)

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
		return stream.Conn.Client().JsonEmit(socket.JsonPack{
			Event: "/asyncServer",
			Data:  string(stream.Data),
		})
	})

	go udpClient.SetRouter(clientRouter).Connect()

	udpClient.OnSuccess = func() {
		ready <- true
	}

	<-ready
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

func Test_UDP_Client_Async(t *testing.T) {

	var asyncClient = async.NewClient[client.Conn](udpClient)

	var wait = sync.WaitGroup{}

	wait.Add(100)

	for i := 0; i < 100; i++ {
		var index = i
		go func() {
			stream, err := asyncClient.JsonEmit(socket.JsonPack{
				Event: "/asyncClient",
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

func Test_UDP_Client(t *testing.T) {

	var count = 10000

	var flag = true

	var mux = new(sync.WaitGroup)

	mux.Add(count)

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == `"i am server"`, "stream is nil")
			return nil
		})
	})

	for i := 0; i < count; i++ {
		_ = udpClient.JsonEmit(socket.JsonPack{
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

	var err = udpClient.JsonEmit(socket.JsonPack{
		Event: "/JsonFormat",
		Data: kitty2.M{
			"name": "kitty",
			"age":  "18",
		},
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

	var err = udpClient.Emit(socket.Pack{
		Event: "/Emit",
		Data:  []byte(`{"name":"kitty","age":18}`),
	})

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

	var err = udpClient.ProtoBufEmit(socket.ProtoBufPack{
		Event: "/ProtoBufEmit",
		Data:  &buf,
	})

	assert.True(t, err == nil, err)

	mux.Wait()
}

func Test_UDP_Server_Async(t *testing.T) {

	var asyncServer = async.NewServer[server.Conn](udpServer)

	var wait = sync.WaitGroup{}

	wait.Add(100)

	for i := 0; i < 100; i++ {
		var index = i
		go func() {
			stream, err := asyncServer.JsonEmit(fd, socket.JsonPack{
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

func Test_UDP_Shutdown(t *testing.T) {
	shutdown()
}
