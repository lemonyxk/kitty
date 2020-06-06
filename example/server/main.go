package main

import (
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/http"
	server3 "github.com/Lemo-yxk/lemo/http/server"
	"github.com/Lemo-yxk/lemo/tcp/server"
	"github.com/Lemo-yxk/lemo/utils"
	server2 "github.com/Lemo-yxk/lemo/websocket/server"
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

	// var progress = utils.HttpClient.NewProgress()
	// progress.Rate(0.01).OnProgress(func(p []byte, current int64, total int64) {
	// 	console.OneLine("Downloading... %d %d B complete", current, total)
	// })
	//
	// utils.HttpClient.Get("https://www.twle.cn/static/js/jquery.min.js").Progress(progress).Send()
}

func run() {

	var webSocketServer = &server2.Server{Host: "0.0.0.0", Port: 8667, Path: "/", AutoBind: true}

	var webSocketServerRouter = &server2.Router{IgnoreCase: true}

	webSocketServer.Use(func(next server2.Middle) server2.Middle {
		return func(conn *server2.WebSocket, receive *lemo.ReceivePackage) {
			next(conn, receive)
		}
	})

	webSocketServer.OnMessage = func(conn *server2.WebSocket, messageType int, msg []byte) {
		console.Log(len(msg))
	}

	webSocketServerRouter.Group("/hello").Handler(func(handler *server2.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server2.WebSocket, receive *lemo.Receive) exception.Error {
			console.Log(string(receive.Body.Message))
			return conn.Json(lemo.JsonPackage{
				Event: "/hello/world",
				Data:  "i am server",
			})
		})
	})

	go webSocketServer.SetRouter(webSocketServerRouter).Start()

	var httpServer = server3.Server{Host: "0.0.0.0", Port: 8666, AutoBind: true}

	var httpServerRouter = &server3.Router{}

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			if stream.Request.Header.Get("Upgrade") == "websocket" {
				httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "0.0.0.0:8667"}).ServeHTTP(stream.Response, stream.Request)
			} else {
				console.Log(1, "start")
				next(stream)
				console.Log(1, "end")
			}
		}
	})

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			console.Log(2, "start")
			next(stream)
			console.Log(2, "end")
		}
	})

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) exception.Error {
		console.Log("handler")
		return exception.New(stream.EndString("hello"))
	})

	httpServerRouter.Group("/hello").Handler(func(handler *server3.RouteHandler) {
		handler.Get("/world").Handler(func(t *http.Stream) exception.Error {
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	go httpServer.SetRouter(httpServerRouter).Start()

	console.Log("start success")

	var tcpServer = &server.Server{
		Host:     "127.0.0.1",
		Port:     8888,
		AutoBind: true,
	}

	tcpServer.OnMessage = func(conn *server.Socket, messageType int, msg []byte) {
		console.Log(len(msg))
	}

	var tcpServerRouter = &server.Router{IgnoreCase: true}

	tcpServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Socket, receive *lemo.Receive) exception.Error {
			console.Log(string(receive.Body.Message))
			return nil
		})
	})

	tcpServer.SetRouter(tcpServerRouter).Start()

}
