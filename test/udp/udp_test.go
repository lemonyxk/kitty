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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	"github.com/lemonyxk/kitty/v2/socket/udp/client"
	"github.com/lemonyxk/kitty/v2/socket/udp/server"
	"github.com/stretchr/testify/assert"
)

var stop = make(chan bool)

var mux = new(sync.WaitGroup)

func shutdown() {
	stop <- true
}

var udpServer *server.Server

var udpServerRouter *router.Router[*socket.Stream[server.Conn]]

var udpClient *client.Client

var clientRouter *router.Router[*socket.Stream[client.Conn]]

var addr = "127.0.0.1:8668"

func initServer(fn func()) {

	// create server
	udpServer = kitty.NewUdpServer(addr)

	// event
	udpServer.OnOpen = func(conn server.Conn) {}
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

	udpServerRouter.Route("/async").Handler(func(stream *socket.Stream[server.Conn]) error {
		return stream.Conn.JsonEmit(socket.JsonPack{
			Event: "/async",
			Data:  "async test",
		})
	})

	go udpServer.SetRouter(udpServerRouter).Start()

	udpServer.OnSuccess = func() {
		fn()
	}
}

func initClient(fn func()) {
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

	go udpClient.SetRouter(clientRouter).Connect()

	udpClient.OnSuccess = func() {
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

		_ = udpClient.Close()
		_ = udpServer.Shutdown()

	}()

	for {
		if sucServer && sucClient {
			t.Run()
			break
		}
	}

}

func Test_UDP_Client_Async(t *testing.T) {
	stream, err := udpClient.Async().JsonEmit(socket.JsonPack{
		Event: "/async",
		Data:  strings.Repeat("hello world!", 1),
	})

	assert.True(t, err == nil, err)

	assert.True(t, stream != nil, "stream is nil")

	assert.True(t, string(stream.Data) == `"async test"`, "stream is nil")
}

func Test_UDP_Client(t *testing.T) {

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

func Test_UDP_Shutdown(t *testing.T) {
	shutdown()
}
