/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2020-09-18 16:40
**/

package tcp

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/tcp/client"
	"github.com/lemoyxk/kitty/socket/tcp/server"
	"github.com/stretchr/testify/assert"
)

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var tcpServer *server.Server

var tcpServerRouter *server.Router

var tcpClient *client.Client

var clientRouter *client.Router

var addr = "127.0.0.1:8667"

func initServer(fn func()) {

	// create server
	tcpServer = kitty.NewTcpServer(addr)

	// event
	tcpServer.OnOpen = func(conn *server.Conn) {}
	tcpServer.OnClose = func(conn *server.Conn) {}
	tcpServer.OnError = func(err error) {}
	tcpServer.OnMessage = func(conn *server.Conn, msg []byte) {}

	// middleware
	tcpServer.Use(func(next server.Middle) server.Middle {
		return func(conn *server.Conn, stream *socket.Stream) {
			next(conn, stream)
		}
	})

	// create router
	tcpServerRouter = &server.Router{StrictMode: true}

	// set group route
	tcpServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Conn, stream *socket.Stream) error {
			return conn.JsonEmit(socket.JsonPack{
				Event: "/hello/world",
				Data:  "i am server",
				ID:    stream.ID,
			})
		})
	})

	tcpServerRouter.Route("/async").Handler(func(conn *server.Conn, stream *socket.Stream) error {
		return conn.JsonEmit(socket.JsonPack{
			Event: "/async",
			Data:  "async test",
			ID:    stream.ID,
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
	tcpClient.OnClose = func(c *client.Client) {}
	tcpClient.OnOpen = func(c *client.Client) {}
	tcpClient.OnError = func(err error) {}
	tcpClient.OnMessage = func(client *client.Client, msg []byte) {}

	// create router
	clientRouter = &client.Router{StrictMode: true}

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
	stream, err := tcpClient.Async().JsonEmit(socket.JsonPack{
		Event: "/async",
		Data:  strings.Repeat("hello world!", 1),
	})

	assert.True(t, err == nil, err)

	assert.True(t, stream != nil, "stream is nil")

	assert.True(t, string(stream.Data) == `"async test"`, "stream is nil")
}

func Test_TCP_Client(t *testing.T) {

	var id int64 = 123456789

	var count = 1

	var flag = true

	clientRouter.Group("/hello").Handler(func(handler *client.RouteHandler) {
		handler.Route("/world").Handler(func(c *client.Client, stream *socket.Stream) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == `"i am server"`, "stream is nil")
			assert.True(t, stream.ID == id, "id not match", stream.ID)
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = tcpClient.JsonEmit(socket.JsonPack{
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

func Test_TCP_Shutdown(t *testing.T) {
	shutdown()
}
