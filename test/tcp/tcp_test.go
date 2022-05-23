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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	async2 "github.com/lemonyxk/kitty/v2/socket/async"
	"github.com/lemonyxk/kitty/v2/socket/tcp/client"
	"github.com/lemonyxk/kitty/v2/socket/tcp/server"
	"github.com/stretchr/testify/assert"
)

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var tcpServer *server.Server

var tcpServerRouter *router.Router[*socket.Stream[server.Conn]]

var tcpClient *client.Client

var clientRouter *router.Router[*socket.Stream[client.Conn]]

var addr = "127.0.0.1:8667"

func initServer(fn func()) {

	// create server
	tcpServer = kitty.NewTcpServer(addr)

	// event
	tcpServer.OnOpen = func(conn server.Conn) {}
	tcpServer.OnClose = func(conn server.Conn) {}
	tcpServer.OnError = func(err error) {}
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
			return stream.Conn.JsonEmit(socket.JsonPack{
				Event: "/hello/world",
				Data:  "i am server",
			})
		})
	})

	tcpServerRouter.Route("/asyncClient").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Conn.JsonEmit(socket.JsonPack{
			Event: "/asyncClient",
			Data:  string(stream.Data),
		})
	})

	go tcpServer.SetRouter(tcpServerRouter).Start()

	tcpServer.OnSuccess = func() {
		fn()
	}
}

func initClient(fn func()) {
	// create client
	tcpClient = kitty.NewTcpClient(addr)
	tcpClient.ReconnectInterval = time.Second
	tcpClient.HeartBeatInterval = time.Second

	// event
	tcpClient.OnClose = func(c client.Conn) {}
	tcpClient.OnOpen = func(c client.Conn) {}
	tcpClient.OnError = func(err error) {}
	tcpClient.OnMessage = func(client client.Conn, msg []byte) {}

	// create router
	clientRouter = &router.Router[*socket.Stream[client.Conn]]{StrictMode: true}

	clientRouter.Route("/asyncServer").Handler(func(stream *socket.Stream[client.Conn]) error {
		return stream.Conn.JsonEmit(socket.JsonPack{
			Event: "/asyncServer",
			Data:  string(stream.Data),
		})
	})

	go tcpClient.SetRouter(clientRouter).Connect()

	tcpClient.OnSuccess = func() {
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

		_ = tcpClient.Close()
		_ = tcpServer.Shutdown()
	}()

	for {
		if sucServer && sucClient {
			t.Run()
			break
		}
	}

}

func Test_TCP_Client_Async(t *testing.T) {

	var asyncClient = async2.NewClient[client.Conn](tcpClient)

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

func Test_TCP_Client(t *testing.T) {

	var count = 10000

	var flag = true

	clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[client.Conn]) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == `"i am server"`, "stream is nil")
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = tcpClient.JsonEmit(socket.JsonPack{
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

	var asyncServer = async2.NewServer[server.Conn](tcpServer)

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

func Test_TCP_Shutdown(t *testing.T) {
	shutdown()
}
