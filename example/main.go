package main

import (
	"fmt"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	console.Log(fmt.Sprintf("%t", "a"))

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
			var captcha = utils.Captcha.New(240, 80)
			console.Log(captcha.Digits())
			return exception.New(t.EndString(`
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<meta http-equiv="X-UA-Compatible" content="ie=edge" />
		<title>Document</title>
	</head>
	<body>
		<img src="data:image/png;base64,` + string(captcha.ToBase64()) + `"/>
	</body>
</html>
`))
		})
	})

	server.Start(webSocketServer.Router(webSocketServerRouter), httpServer.Router(httpServerRouter))

}
