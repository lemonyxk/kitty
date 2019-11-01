package main

import (
	"os"
	"time"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/logger"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	logger.SetFlag(logger.DEBUG | logger.LOG)

	logger.SetLogHook(func(status string, t time.Time, file string, line int, v ...interface{}) {
		println(status)
	})

	var httpClient = utils.NewHttpClient()

	var url = "http://127.0.0.1/Proxy/User/login"

	res, err := httpClient.Post(url).Form(lemo.M{"account_name": 571413495, "password": 123456.0111}).Send()
	if err != nil {
		panic(err)
	}

	logger.Console(string(res))

	HttpServer()

	utils.ListenSignal(func(sig os.Signal) {
		logger.Console(sig)
	})

}

func HttpServer() {

	var server = lemo.Server{Host: "127.0.0.1", Port: 8666}

	var httpServer = lemo.HttpServer{}

	httpServer.OnError = func(err func() *lemo.Error) {
		logger.Console(err())
	}

	httpServer.Group("/hello").Handler(func(this *lemo.HttpServer) {
		this.Get("/world").Handler(func(t *lemo.Stream) func() *lemo.Error {
			logger.Console("ha")
			return lemo.NewError(t.End("hello world!"))
		})
	})

	httpServer.SetStaticPath("/dir", "./example/public")

	// var websocketServer = lemo.WebSocketServer{IgnoreCase: true, Path: "/"}
	//
	// websocketServer.Group("/hello").Handler(func(this *lemo.WebSocketServer) {
	// 	this.Route("/world").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) func() *lemo.Error {
	// 		conn.JsonFormat(conn.Fd,lemo.JsonPackage{
	// 			Event:   re,
	// 			Message: nil,
	// 		})
	// 		return nil
	// 	})
	// })

	server.Start(nil, &httpServer)

}
