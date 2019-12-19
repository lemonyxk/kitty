```go
package main

import (
	"os"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	exception.Try(func() {

		utils.Goroutine.Run(func() {
			var a []int
			a[0] = 1
		})

		utils.Goroutine.Run(func() {
			var a *lemo.Server
			a = nil
			a.Host = "a"
		})

		utils.Goroutine.Run(func() {
			var a interface{}
			a = "1"
			console.Log(a.(int))
		})

		var a []int
		a[0] = 1

		exception.Assert(os.Open(""))

	}).Catch(func(errorFunc exception.ErrorFunc) exception.ErrorFunc {

		utils.Goroutine.Watch(func(errorFunc exception.ErrorFunc) {
			console.Error(errorFunc)
		})

		console.Error(errorFunc)
		return nil

	}).Finally(func(errorFunc exception.ErrorFunc) exception.ErrorFunc {

		console.Error("finally")
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
			return exception.New(t.EndString("hello world"))
			// return exception.New(t.EndBytes(utils.Captcha.New(240, 80).ToPNG()))
		})
	})

	server.Start(webSocketServer.Router(webSocketServerRouter), httpServer.Router(httpServerRouter))

}
```