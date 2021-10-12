package main

import (
	"log"
	"os"
	"time"

	"github.com/lemoyxk/kitty/http"
	"github.com/lemoyxk/kitty/http/client"
	server3 "github.com/lemoyxk/kitty/http/server"
	"github.com/lemoyxk/kitty/socket"
	client2 "github.com/lemoyxk/kitty/socket/tcp/client"
	"github.com/lemoyxk/kitty/socket/tcp/server"
	client3 "github.com/lemoyxk/kitty/socket/udp/client"
	server4 "github.com/lemoyxk/kitty/socket/udp/server"
	client4 "github.com/lemoyxk/kitty/socket/websocket/client"
	server2 "github.com/lemoyxk/kitty/socket/websocket/server"
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

	runHttpServer()
	runWebSocketServer()
	runTcpServer()
	runUdpServer()

	runHttpClient()
	runTcpClient()
	runUdpClient()
	runWebSocketClient()

	select {}
	// utils.Signal.ListenKill().Done(func(sig os.Signal) {
	// 	console.Info(sig)
	// })

	// var progress = utils.HttpClient.NewProgress()
	// progress.Rate(0.01).OnProgress(func(p []byte, current int64, total int64) {
	// 	console.OneLine("Downloading... %d %d B complete", current, total)
	// })
	//
	// utils.HttpClient.Get("https://www.twle.cn/static/js/jquery.min.js").Progress(progress).Send()

	// console.SetFormatter(console.NewJsonFormatter())

}

func runTcpServer() {
	var tcpServer = server.NewTcpServer("127.0.0.1:8888")

	var tcpServerRouter = &server.Router{StrictMode: true}

	tcpServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Conn, stream *socket.Stream) error {
			log.Println(string(stream.Data))
			return nil
		})
	})

	tcpServer.OnSuccess = func() {
		log.Println(tcpServer.LocalAddr())
	}

	go tcpServer.SetRouter(tcpServerRouter).Start()
}

func runWebSocketServer() {
	var webSocketServer = server2.NewWebSocketServer("127.0.0.1:8667")

	var webSocketServerRouter = &server2.Router{StrictMode: true}

	webSocketServerRouter.Group("/hello").Handler(func(handler *server2.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server2.Conn, stream *socket.Stream) error {
			log.Println(string(stream.Data))
			return nil
		})
	})

	webSocketServer.OnSuccess = func() {
		log.Println(webSocketServer.LocalAddr())
	}

	go webSocketServer.SetRouter(webSocketServerRouter).Start()
}

func runUdpServer() {
	var udpServer = server4.NewUdpServer("127.0.0.1:5000")

	var udpServerRouter = server4.NewUdpServerRouter()

	udpServerRouter.Group("/hello").Handler(func(handler *server4.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server4.Conn, stream *socket.Stream) error {
			log.Println(string(stream.Data))
			return nil
		})
	})

	udpServer.OnSuccess = func() {
		log.Println(udpServer.LocalAddr())
	}

	go udpServer.SetRouter(udpServerRouter).Start()
}

func runHttpServer() {
	var httpServer = server3.NewHttpServer("127.0.0.1:8666")
	httpServer.CertFile = "/Users/lemo/test/go/localhost+2.pem"
	httpServer.KeyFile = "/Users/lemo/test/go/localhost+2-key.pem"

	var httpServerRouter = &server3.Router{}

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) error {
		stream.AutoParse()
		return stream.EndString("hello world!")
	})

	httpServerRouter.Group("/hello").Handler(func(handler *server3.RouteHandler) {
		handler.Get("/world").Handler(func(t *http.Stream) error {
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	httpServer.OnSuccess = func() {
		log.Println(httpServer.LocalAddr())
	}

	httpServerRouter.SetStaticPath("/", "./example/server/public")

	go httpServer.SetRouter(httpServerRouter).Start()
}

func runHttpClient() {
	time.AfterFunc(time.Second, func() {
		var res = client.Get("http://127.0.0.1:8666/hello").Query().Send()
		if res.LastError() == nil {
			log.Println("http OK!")
		}
	})
}

func runTcpClient() {
	time.AfterFunc(time.Second, func() {
		var tcpClient = client2.NewTcpClient("127.0.0.1:8888")
		var clientRouter = client2.NewTcpClientRouter()
		go tcpClient.SetRouter(clientRouter).Connect()
	})
}

func runUdpClient() {
	time.AfterFunc(time.Second, func() {
		var udpClient = client3.NewUdpClient("127.0.0.1:5000")
		var clientRouter = client3.NewUdpClientRouter()
		go udpClient.SetRouter(clientRouter).Connect()
	})
}

func runWebSocketClient() {
	time.AfterFunc(time.Second, func() {
		var webSocketClient = client4.NewWebSocketClient("ws://127.0.0.1:8667")
		var clientRouter = client4.NewWebSocketClientRouter()
		go webSocketClient.SetRouter(clientRouter).Connect()
	})
}
