package main

import (
	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
)

func main() {

	var server = lemo.Server{Host: "0.0.0.0", Port: 8666}

	var webSocketServer = &lemo.WebSocketServer{}

	var webSocketServerRouter = &lemo.WebSocketServerRouter{}

	webSocketServerRouter.Group("/hello").Handler(func(handler *lemo.WebSocketServerRouteHandler) {
		handler.Route("/world").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) func() *exception.Error {
			console.Debug("hello world")
			return nil
		})
	})

	var httpServer = &lemo.HttpServer{}

	var httpServerRouter = &lemo.HttpServerRouter{}

	httpServerRouter.Group("/hello").Handler(func(handler *lemo.HttpServerRouteHandler) {
		handler.Get("/world").Handler(func(t *lemo.Stream) func() *exception.Error {
			return exception.New(t.End("hello world!"))
		})
	})

	server.Start(webSocketServer.Router(webSocketServerRouter), httpServer.Router(httpServerRouter))

}
