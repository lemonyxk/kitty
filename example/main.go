package main

import (
	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
)

func main() {

	exception.Try(func() {

		exception.Assert(1)

	}).Catch(func(err error, trace *caller.Trace) exception.ErrorFunc {
		console.Log(2)
		console.Error(err)
		console.Println(trace)
		return exception.New(err)
	}).Finally(func(err error) exception.ErrorFunc {
		return nil
	})

	var server = lemo.Server{Host: "0.0.0.0", Port: 8666}

	var webSocketServer = &lemo.WebSocketServer{}

	var webSocketServerRouter = &lemo.WebSocketServerRouter{}

	webSocketServerRouter.Group("/hello").Handler(func(handler *lemo.WebSocketServerRouteHandler) {
		handler.Route("/world").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) exception.ErrorFunc {
			console.Debug("hello world")
			return nil
		})
	})

	var httpServer = &lemo.HttpServer{}

	var httpServerRouter = &lemo.HttpServerRouter{}

	httpServerRouter.Group("/hello").Handler(func(handler *lemo.HttpServerRouteHandler) {
		handler.Get("/world").Handler(func(t *lemo.Stream) exception.ErrorFunc {
			return exception.New(t.EndJson("hello world"))
			// return exception.New(t.EndBytes(utils.Captcha.New(240, 80).ToPNG()))
		})
	})

	server.Start(webSocketServer.Router(webSocketServerRouter), httpServer.Router(httpServerRouter))

}
