package main

import (
	"os"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/logger"
)

func main() {

	HttpServer()

	lemo.ListenSignal(func(sig os.Signal) {
		logger.Log(sig)
	})

}

func HttpServer() {

	var server = lemo.Server{Host: "127.0.0.1", Port: 8666}

	var httpServer = lemo.HttpServer{}

	httpServer.OnError = func(err func() *lemo.Error) {
		logger.Log(err())
	}

	httpServer.Group("/hello").Handler(func(this *lemo.HttpServer) {
		this.Get("/world").Handler(func(t *lemo.Stream) func() *lemo.Error {
			return lemo.NewError(t.End("hello world!"))
		})
	})

	httpServer.SetStaticPath("/dir", "./example/public")

	server.Start(nil, &httpServer)

}
