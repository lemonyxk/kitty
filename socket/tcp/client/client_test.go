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

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/tcp/server"
)

var stop = make(chan bool)

var mux sync.WaitGroup

func shutdown() {
	stop <- true
}

var tcpServer *server.Server

var tcpServerRouter *server.Router

var client *Client

var clientRouter *Router

var host = "127.0.0.1:8667"

func initServer(fn func()) {

	// create server
	tcpServer = &server.Server{Host: host}

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
	tcpServerRouter = &server.Router{IgnoreCase: true}

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

	go tcpServer.SetRouter(tcpServerRouter).Start()

	tcpServer.OnSuccess = func() {
		fn()
	}
}

func initClient(fn func()) {
	// create client
	client = &Client{Host: host, Reconnect: true, AutoHeartBeat: true}

	// event
	client.OnClose = func(c *Client) {}
	client.OnOpen = func(c *Client) {}
	client.OnError = func(err error) {}
	client.OnMessage = func(client *Client, msg []byte) {}

	// create router
	clientRouter = &Router{IgnoreCase: true}

	go client.SetRouter(clientRouter).Connect()

	client.OnSuccess = func() {
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

		_ = client.Close()
		_ = tcpServer.Shutdown()
	}()

	for {
		if sucServer && sucClient {
			t.Run()
			break
		}
	}

}

//
// func Test_Client_Async(t *testing.T) {
// 	stream, err := client.AsyncJsonEmit(socket.JsonPackage{
// 		Event: "/hello/world",
// 		Data:  strings.Repeat("hello world!", 1),
// 	})
//
// 	kitty.AssertEqual(t, err == nil, err)
//
// 	kitty.AssertEqual(t, stream != nil, "stream is nil")
//
// 	kitty.AssertEqual(t, string(stream.Message) == `"i am server"`, "stream is nil")
// }

func Test_Client(t *testing.T) {

	var id int64 = 123456789

	var count = 10000

	clientRouter.Group("/hello").Handler(func(handler *RouteHandler) {
		handler.Route("/world").Handler(func(c *Client, stream *socket.Stream) error {
			defer mux.Add(-1)
			kitty.AssertEqual(t, string(stream.Data) == `"i am server"`, "stream is nil")
			kitty.AssertEqual(t, stream.ID == id, "id not match", stream.ID)
			return nil
		})
	})

	for i := 0; i < count; i++ {
		mux.Add(1)
		_ = client.JsonEmit(socket.JsonPack{
			Event: "/hello/world",
			Data:  strings.Repeat("hello world!", 1),
			ID:    id,
		})

	}

	go func() {
		<-time.After(3 * time.Second)
		mux.Done()
		t.Fatal("timeout")
	}()

	mux.Wait()
}

func Test_Shutdown(t *testing.T) {
	shutdown()
}
