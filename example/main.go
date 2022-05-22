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
	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/errors"
	awesomepackage "github.com/lemonyxk/kitty/v2/example/protobuf"
	"github.com/lemonyxk/kitty/v2/http"
	"github.com/lemonyxk/kitty/v2/http/client"
	server3 "github.com/lemonyxk/kitty/v2/http/server"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
	client2 "github.com/lemonyxk/kitty/v2/socket/tcp/client"
	"github.com/lemonyxk/kitty/v2/socket/tcp/server"
	udpClient2 "github.com/lemonyxk/kitty/v2/socket/udp/client"
	server4 "github.com/lemonyxk/kitty/v2/socket/udp/server"
	server2 "github.com/lemonyxk/kitty/v2/socket/websocket/server"
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

	tcpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *socket.Stream[server.Conn]) {
			next(stream)
		}
	})

	var before = func(stream *socket.Stream[server.Conn]) error {
		return errors.NewWithStack("before")
	}

	tcpServerRouter.Group("/hello").Before(before).Handler(func(handler *router.Handler[*socket.Stream[server.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server.Conn]) error {
			return stream.Conn.JsonEmit(socket.JsonPack{
				Event: "/hello/world",
				Data:  nil,
			})
		})
	})

	tcpServer.OnError = func(err error) {
		log.Printf("[tcpServer] %+v", err)
	}

	tcpServer.OnSuccess = func() {
		log.Println(tcpServer.LocalAddr())
	}

	go tcpServer.SetRouter(tcpServerRouter).Start()
}

func runWebSocketServer() {
	var webSocketServer = kitty.NewWebSocketServer("127.0.0.1:8667")

	var webSocketServerRouter = kitty.NewWebSocketServerRouter()

	webSocketServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server2.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server2.Conn]) error {
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

	udpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[server4.Conn]]) {
		handler.Route("/world").Handler(func(stream *socket.Stream[server4.Conn]) error {
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
	// httpServer.CertFile = "/Users/lemon/test/go/localhost+2.pem"
	// httpServer.KeyFile = "/Users/lemon/test/go/localhost+2-key.pem"

	var httpServerRouter = kitty.NewHttpServerRouter()

	var httpStaticServerRouter = kitty.NewHttpServerStaticRouter()

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
			next(stream)
		}
	})

	httpServerRouter.RouteMethod("GET", "/hello").Handler(func(stream *http.Stream) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		return stream.EndString("hello world!")
	})

	httpServerRouter.RouteMethod("POST", "/proto").Handler(func(stream *http.Stream) error {
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

	var group = httpServerRouter.Group("/hello").Create()
	group.Get("/world").Handler(func(t *http.Stream) error {
		return t.JsonFormat("SUCCESS", 200, os.Getpid())
	})

	httpServer.OnSuccess = func() {
		log.Println(httpServer.LocalAddr())
	}

	// httpServerRouter.SetStaticPath("/", "", http2.Dir("./example/public"))

	httpStaticServerRouter.SetStaticPath("/", "public", http2.FS(fileSystem))
	httpStaticServerRouter.SetStaticPath("/js", "test", http2.Dir("./example"))
	// httpServerRouter.SetDefaultIndex("index.html", "index.htm")
	// httpServerRouter.SetStaticFileMiddle(".md").Handler(func(w http2.ResponseWriter, r *http2.Request, f fs.File, i fs.FileInfo) error {
	// 	return nil
	// })
	httpStaticServerRouter.SetOpenDir(0)

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

		clientRouter.Group("/hello").Handler(func(handler *router.Handler[*socket.Stream[client2.Conn]]) {
			handler.Route("/world").Handler(func(stream *socket.Stream[client2.Conn]) error {
				log.Println("tcp OK!")
				return nil
			})
		})

		go tcpClient.SetRouter(clientRouter).Connect()

		tcpClient.OnSuccess = func() {
			_ = tcpClient.JsonEmit(socket.JsonPack{
				Event: "/hello/world",
				Data:  nil,
			})
		}
	})
}

func runUdpClient() {
	time.AfterFunc(time.Second, func() {
		var udpClient = kitty.NewUdpClient("127.0.0.1:5000")
		var clientRouter = kitty.NewUdpClientRouter()

		udpClient.OnOpen = func(client udpClient2.Conn) {
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
