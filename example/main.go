package main

import (
	"net/http"
	"os"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	var httpServer = lemo.HttpServer{Host: "0.0.0.0", Port: 8666}

	var httpServerRouter = &lemo.HttpServerRouter{}

	httpServer.Use(func(next lemo.HttpServerMiddle) lemo.HttpServerMiddle {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	})

	httpServerRouter.Group("/hello").Handler(func(handler *lemo.HttpServerRouteHandler) {
		handler.Get("/world").Handler(func(t *lemo.Stream) exception.ErrorFunc {
			return exception.New(t.EndString("hello world"))
			// return exception.New(t.EndBytes(utils.Captcha.New(240, 80).ToPNG()))
		})
	})

	go httpServer.SetRouter(httpServerRouter).Start()

	var webSocketServer = &lemo.WebSocketServer{Host: "0.0.0.0", Port: 8667, Path: "/"}

	var webSocketServerRouter = &lemo.WebSocketServerRouter{}

	webSocketServer.Use(func(next lemo.WebSocketServerMiddle) lemo.WebSocketServerMiddle {
		return func(conn *lemo.WebSocket, receive *lemo.ReceivePackage) {
			console.Log(1)
			next(conn, receive)
			console.Log(2)
		}
	})

	webSocketServerRouter.Group("/hello").Handler(func(handler *lemo.WebSocketServerRouteHandler) {
		handler.Route("/world").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) exception.ErrorFunc {
			console.Log(3)
			return exception.New(conn.JsonFormat(lemo.JsonPackage{
				Event:   "/hello/world",
				Message: "hello world",
			}))
		})
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	utils.Signal.ListenAll().Done(func(sig os.Signal) {
		console.Log(sig)
	})
}
