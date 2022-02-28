package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	http2 "net/http"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lemoyxk/kitty"
	awesomepackage "github.com/lemoyxk/kitty/example/protobuf"
	"github.com/lemoyxk/kitty/http"
	"github.com/lemoyxk/kitty/http/client"
	server3 "github.com/lemoyxk/kitty/http/server"
	"github.com/lemoyxk/kitty/socket"
	"github.com/lemoyxk/kitty/socket/tcp/server"
	udpClient2 "github.com/lemoyxk/kitty/socket/udp/client"
	server4 "github.com/lemoyxk/kitty/socket/udp/server"
	server2 "github.com/lemoyxk/kitty/socket/websocket/server"
)

//go:embed public/**
var fileSystem embed.FS

func main() {

	runHttpServer()
	runWebSocketServer()
	runTcpServer()
	runUdpServer()

	runHttpClientWithProcess()
	runHttpClient()
	runTcpClient()
	runUdpClient()
	runWebSocketClient()

	select {}
}

func runTcpServer() {
	var tcpServer = kitty.NewTcpServer("127.0.0.1:8888")

	var tcpServerRouter = kitty.NewTcpServerRouter()

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
	var webSocketServer = kitty.NewWebSocketServer("127.0.0.1:8667")

	var webSocketServerRouter = kitty.NewWebSocketServerRouter()

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
	var udpServer = kitty.NewUdpServer("127.0.0.1:5000")

	var udpServerRouter = kitty.NewUdpServerRouter()

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
	var httpServer = kitty.NewHttpServer("127.0.0.1:8666")
	httpServer.CertFile = "/Users/lemo/test/go/localhost+2.pem"
	httpServer.KeyFile = "/Users/lemo/test/go/localhost+2-key.pem"

	var httpServerRouter = kitty.NewHttpServerRouter()

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
			next(stream)
		}
	})

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		return stream.EndString("hello world!")
	})

	httpServerRouter.Route("POST", "/proto").Handler(func(stream *http.Stream) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		log.Println(stream.AutoGet("name").String())
		log.Println(stream.Files.First("file"))
		var res awesomepackage.AwesomeMessage
		var msg = stream.Protobuf.Bytes()
		var err = proto.Unmarshal(msg, &res)
		if err != nil {
			return stream.EndString(err.Error())
		}
		return stream.EndString("hello proto!")
	})

	httpServerRouter.Group("/hello").Handler(func(handler *server3.RouteHandler) {
		handler.Get("/world").Handler(func(t *http.Stream) error {
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	httpServer.OnSuccess = func() {
		log.Println(httpServer.LocalAddr())
	}

	// httpServerRouter.SetStaticPath("/", "", http2.Dir("./example/public"))

	httpServerRouter.SetStaticPath("/", "public", http2.FS(fileSystem))
	// httpServerRouter.SetDefaultIndex("index.html", "index.htm")
	httpServerRouter.SetStaticMiddle(".md", func(bts []byte) ([]byte, string) {
		return bts, "text/markdown"
	})
	httpServerRouter.SetOpenDir(true)

	go httpServer.SetRouter(httpServerRouter).Start()
}

func runHttpClientWithProcess() {
	time.AfterFunc(time.Second, func() {
		var progress = kitty.NewHttpClientProgress()
		progress.Rate(0.01).OnProgress(func(p []byte, current int64, total int64) {
			fmt.Printf("\rDownloading... %d %d B complete", current, total)
			if current == total {
				fmt.Println()
			}
		})

		client.Get("https://code.jquery.com/jquery-3.6.0.js").Progress(progress).Query().Send()
		// client.Get("https://127.0.0.1:8666/1.png").Progress(progress).Query().Send()
	})
}

func runHttpClient() {

	time.AfterFunc(time.Second, func() {
		var res = client.Get("https://127.0.0.1:8666/hello").Query().Send()
		if res.LastError() == nil {
			log.Println("http OK!")
		}

		var msg = awesomepackage.AwesomeMessage{
			AwesomeField: "1",
			AwesomeKey:   "2",
		}

		res = client.Post("https://127.0.0.1:8666/proto").Protobuf(&msg).Send()
		if res.LastError() == nil {
			log.Println("http OK!")
		}
	})
}

func runTcpClient() {
	time.AfterFunc(time.Second, func() {
		var tcpClient = kitty.NewTcpClient("127.0.0.1:8888")
		var clientRouter = kitty.NewTcpClientRouter()
		tcpClient.SetRouter(clientRouter).Connect()
	})
}

func runUdpClient() {
	time.AfterFunc(time.Second, func() {
		var udpClient = kitty.NewUdpClient("127.0.0.1:5000")
		var clientRouter = kitty.NewUdpClientRouter()

		udpClient.OnOpen = func(client *udpClient2.Client) {
			log.Println(client.RemoteAddr())
		}

		go udpClient.SetRouter(clientRouter).Connect()

		select {}
	})
}

func runWebSocketClient() {
	time.AfterFunc(time.Second, func() {
		var webSocketClient = kitty.NewWebSocketClient("ws://127.0.0.1:8667")
		var clientRouter = kitty.NewWebSocketClientRouter()
		webSocketClient.SetRouter(clientRouter).Connect()
	})
}
