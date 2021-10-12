/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2020-09-18 16:40
**/

package client

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/udp/server"
	"github.com/stretchr/testify/assert"
)

var stop = make(chan bool)

var mux = new(sync.WaitGroup)

func shutdown() {
	stop <- true
}

var udpServer *server.Server

var udpServerRouter *server.Router

var udpClient *Client

var clientRouter *Router

var addr = "127.0.0.1:8668"

func initServer(fn func()) {

	// create server
	udpServer = server.NewUdpServer(addr)

	// event
	udpServer.OnOpen = func(conn *server.Conn) {}
	udpServer.OnClose = func(conn *server.Conn) {}
	udpServer.OnError = func(err error) {}
	udpServer.OnMessage = func(conn *server.Conn, msg []byte) {}

	// middleware
	udpServer.Use(func(next server.Middle) server.Middle {
		return func(conn *server.Conn, stream *socket.Stream) {
			next(conn, stream)
		}
	})

	// create router
	udpServerRouter = &server.Router{StrictMode: true}

	// set group route
	udpServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Conn, stream *socket.Stream) error {
			return conn.JsonEmit(socket.JsonPack{
				Event: "/hello/world",
				Data:  "i am server",
				ID:    stream.ID,
			})
		})
	})

	udpServerRouter.Route("/async").Handler(func(conn *server.Conn, stream *socket.Stream) error {
		return conn.JsonEmit(socket.JsonPack{
			Event: "/async",
			Data:  "async test",
			ID:    stream.ID,
		})
	})

	go udpServer.SetRouter(udpServerRouter).Start()

	udpServer.OnSuccess = func() {
		fn()
	}
}

func initClient(fn func()) {
	// create client
	udpClient = NewUdpClient(addr)
	udpClient.ReconnectInterval = time.Second
	udpClient.HeartBeatInterval = time.Second

	// event
	udpClient.OnClose = func(c *Client) {}
	udpClient.OnOpen = func(c *Client) {}
	udpClient.OnError = func(err error) {}
	udpClient.OnMessage = func(client *Client, msg []byte) {}

	// create router
	clientRouter = &Router{StrictMode: true}

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

func Test_Client_Async(t *testing.T) {
	stream, err := udpClient.Async().JsonEmit(socket.JsonPack{
		Event: "/async",
		Data:  strings.Repeat("hello world!", 1),
	})

	assert.True(t, err == nil, err)

	assert.True(t, stream != nil, "stream is nil")

	assert.True(t, string(stream.Data) == `"async test"`, "stream is nil")
}

func Test_Client(t *testing.T) {

	var id int64 = 123456789

	var count = 10000

	var flag = true

	clientRouter.Group("/hello").Handler(func(handler *RouteHandler) {
		handler.Route("/world").Handler(func(c *Client, stream *socket.Stream) error {
			defer mux.Add(-1)
			assert.True(t, string(stream.Data) == `"i am server"`, "stream is nil")
			assert.True(t, stream.ID == id, "id not match", stream.ID)
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = udpClient.JsonEmit(socket.JsonPack{
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

func Test_Shutdown(t *testing.T) {
	shutdown()
}
