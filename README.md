```go
package main

import (
	"log"
	"os"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/http"
	server3 "github.com/lemoyxk/kitty/http/server"
	"github.com/lemoyxk/kitty/tcp/server"
	server2 "github.com/lemoyxk/kitty/websocket/server"
)

func main() {
	run()
}

func run() {

	var webSocketServer = &server2.Server{Host: "127.0.0.1:8667", Path: "/"}

	var webSocketServerRouter = &server2.Router{IgnoreCase: true}

	webSocketServer.Use(func(next server2.Middle) server2.Middle {
		return func(conn *server2.WebSocket, receive *kitty.ReceivePackage) {
			next(conn, receive)
		}
	})

	webSocketServer.OnMessage = func(conn *server2.WebSocket, messageType int, msg []byte) {
		log.Println(len(msg))
	}

	webSocketServerRouter.Group("/hello").Handler(func(handler *server2.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server2.WebSocket, receive *kitty.Receive) error {
			log.Println(string(receive.Body.Message))
			return conn.Json(kitty.JsonPackage{
				Event: "/hello/world",
				Data:  "i am server",
			})
		})
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	var httpServer = server3.Server{Host: "127.0.0.1:8666"}

	var httpServerRouter = &server3.Router{}

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			// if stream.Request.Header.Get("Upgrade") == "websocket" {
			// 	httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "0.0.0.0:8667"}).ServeHTTP(stream.Response, stream.Request)
			// } else {
			// 	log.Println(1, "start")
			// 	next(stream)
			// 	log.Println(1, "end")
			// }
			next(stream)
		}
	})

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			// log.Println(2, "start")
			// next(stream)
			// log.Println(2, "end")
			next(stream)
		}
	})

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) error {
		// log.Println("handler")
		return stream.EndString("hello")
	})

	httpServerRouter.Group("/hello").Handler(func(handler *server3.RouteHandler) {
		handler.Get("/world").Handler(func(t *http.Stream) error {
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	httpServer.OnSuccess = func() {
		log.Println(httpServer.LocalAddr())
	}

	go httpServer.SetRouter(httpServerRouter).Start()

	log.Println("start success")

	var tcpServer = &server.Server{Host: "127.0.0.1:8888"}

	tcpServer.OnMessage = func(conn *server.Socket, messageType int, msg []byte) {
		log.Println(len(msg))
	}

	var tcpServerRouter = &server.Router{IgnoreCase: true}

	tcpServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Socket, receive *kitty.Receive) error {
			log.Println(string(receive.Body.Message))
			return nil
		})
	})

	tcpServer.OnSuccess = func() {
		log.Println(tcpServer.LocalAddr())
	}

	tcpServer.SetRouter(tcpServerRouter).Start()

}



```