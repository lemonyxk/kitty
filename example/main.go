package main

import (
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	_ "github.com/mkevac/debugcharts"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	utils.Process.Fork(run, 1)

	utils.Signal.ListenKill().Done(func(sig os.Signal) {
		console.Log(sig)
	})
}

func run() {

	var webSocketServer = &lemo.WebSocketServer{Host: "0.0.0.0", Port: 8667, Path: "/", AutoBind: true}

	var webSocketServerRouter = &lemo.WebSocketServerRouter{}

	webSocketServer.Use(func(next lemo.WebSocketServerMiddle) lemo.WebSocketServerMiddle {
		return func(conn *lemo.WebSocket, receive *lemo.ReceivePackage) {
			console.Log(1)
			next(conn, receive)
			console.Log(2)
		}
	})

	webSocketServer.OnMessage = func(conn *lemo.WebSocket, messageType int, msg []byte) {
		console.Log(os.Getpid())
	}

	webSocketServerRouter.Group("/hello").Handler(func(handler *lemo.WebSocketServerRouteHandler) {
		handler.Route("/world").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) exception.ErrorFunc {
			return conn.JsonFormat(lemo.JsonPackage{
				Event: "/hello/world",
				Message: &lemo.JsonMessage{
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

	httpServerRouter.Group("/hello").Handler(func(handler *lemo.HttpServerRouteHandler) {
		handler.Get("/world").Handler(func(t *lemo.Stream) exception.ErrorFunc {
			time.Sleep(time.Second * 5)
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	go httpServer.SetRouter(httpServerRouter).Start()

	console.Log("start success")
}
