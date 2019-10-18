package main

import (
	"log"
	"os"
	"time"

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

	logger.SetWrite(true)
	logger.SetWriteHook(func(t time.Time, file string, line int, v ...interface{}) {
		log.Println(t, file, line, v)
	})

	var server = lemo.Server{Host: "127.0.0.1", Port: 8666}

	var httpServer = lemo.HttpServer{}

	httpServer.OnError = func(err func() *lemo.Error) {
		logger.Log(err())
	}

	httpServer.Group("/hello").Handler(func(this *lemo.HttpServer) {
		this.Get("/world").Handler(func(t *lemo.Stream) func() *lemo.Error {
			logger.Log("ha")
			return lemo.NewError(t.End("hello world!"))
		})
	})

	httpServer.SetStaticPath("/dir", "./example/public")

	server.Start(nil, &httpServer)

}
