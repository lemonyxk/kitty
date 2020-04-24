package main

import (
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	// utils.Process.Fork(run, 1)
	//
	// go func() {
	// 	http.HandleFunc("/reload", func(writer http.ResponseWriter, request *http.Request) {
	// 		utils.Process.Reload()
	// 	})
	// 	console.Log(http.ListenAndServe(":12345", nil))
	// }()
	run()
	utils.Signal.ListenKill().Done(func(sig os.Signal) {
		console.Log(sig)
	})
}

func run() {

	var webSocketServer = &lemo.WebSocketServer{Host: "0.0.0.0", Port: 8667, Path: "/", AutoBind: true}

	var webSocketServerRouter = &lemo.WebSocketServerRouter{}

	webSocketServer.Use(func(next lemo.WebSocketServerMiddle) lemo.WebSocketServerMiddle {
		return func(conn *lemo.WebSocket, receive *lemo.ReceivePackage) {
			next(conn, receive)
		}
	})

	webSocketServer.OnMessage = func(conn *lemo.WebSocket, messageType int, msg []byte) {
		console.Log(len(msg))
	}

	webSocketServerRouter.Group("/hello").Handler(func(handler *lemo.WebSocketServerRouteHandler) {
		handler.Route("/world").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) exception.ErrorFunc {
			return conn.JsonFormat(lemo.JsonPackage{
				Event: "/hello/world",
				Message: &lemo.JsonFormat{
					Status: "",
					Code:   0,
					Msg:    nil,
				},
			})
		})
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	var httpServer = lemo.HttpServer{Host: "0.0.0.0", Port: 8666, AutoBind: true}

	var httpServerRouter = &lemo.HttpServerRouter{}

	httpServer.Use(func(next lemo.HttpServerMiddle) lemo.HttpServerMiddle {
		return func(stream *lemo.Stream) {
			if stream.Request.Header.Get("Upgrade") == "websocket" {
				httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "0.0.0.0:8667"}).ServeHTTP(stream.Response, stream.Request)
			} else {
				next(stream)
			}
		}
	})

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *lemo.Stream) exception.ErrorFunc {
		return exception.New(stream.EndString("hello"))
	})

	httpServerRouter.Group("/hello").Handler(func(handler *lemo.HttpServerRouteHandler) {
		handler.Get("/world").Handler(func(t *lemo.Stream) exception.ErrorFunc {
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	go httpServer.SetRouter(httpServerRouter).Start()

	console.Log("start success")

	var tcpServer = &lemo.SocketServer{
		Host:     "127.0.0.1",
		Port:     8888,
		AutoBind: true,
	}

	tcpServer.OnMessage = func(conn *lemo.Socket, messageType int, msg []byte) {
		console.Log(len(msg))
	}

	var tcpServerRouter = &lemo.SocketServerRouter{IgnoreCase: true}

	tcpServerRouter.Group("/hello").Handler(func(handler *lemo.SocketServerRouteHandler) {
		handler.Route("/world").Handler(func(conn *lemo.Socket, receive *lemo.Receive) exception.ErrorFunc {
			console.Log(len(receive.Body.Message))
			return nil
		})
	})

	tcpServer.SetRouter(tcpServerRouter).Start()

}
