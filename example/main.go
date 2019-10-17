package main

import (
	"io/ioutil"
	"log"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/logger"
	"github.com/Lemo-yxk/lemo/protobuf-origin"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	go SocketServer()

	time.Sleep(time.Second)

	go SocketClient()

	// go WebSocketServer()
	// //
	// time.Sleep(time.Second)
	// //
	// go WebSocketClient()

	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞
	<-signalChan

}

func SocketClient() {

	var i int32 = 0

	var socketClient = lemo.SocketClient{Host: "127.0.0.1", Port: 5000, AutoHeartBeat: true, HeartBeatInterval: 3, HeartBeatTimeout: 5}

	socketClient.OnOpen = func(c *lemo.SocketClient) {
		logger.Log("client open")
	}

	socketClient.Group("/hello").Handler(func(this *lemo.SocketClient) {
		this.Route("/world").Handler(func(c *lemo.SocketClient, receive *lemo.Receive) func() *lemo.Error {
			logger.Log(string(receive.Message.Message), atomic.AddInt32(&i, 1))
			_ = c.JsonEmit(lemo.JsonPackage{Event: "/hello/kid", Message: lemo.M{"who": "me"}})
			return nil
		})
	})

	socketClient.OnClose = func(c *lemo.SocketClient) {
		logger.Log("client close")
	}

	socketClient.OnError = func(err func() *lemo.Error) {
		logger.Log(err())
	}

	socketClient.Connect()
}

func SocketServer() {

	var j int32 = 0

	var server = lemo.SocketServer{Host: "0.0.0.0", Port: 5000, HeartBeatTimeout: 5, HeartBeatInterval: 3}

	server.OnOpen = func(conn *lemo.Socket) {
		for i := 0; i < 10000; i++ {
			go func() {
				_ = conn.JsonEmit(conn.Fd, lemo.JsonPackage{Event: "/hello/world", Message: lemo.M{"name": "1", "addr": 2}})
			}()
		}
	}

	server.OnMessage = func(conn *lemo.Socket, messageType int, msg []byte) {
		log.Println(string(msg))
	}

	server.Group("/hello").Handler(func(socket *lemo.SocketServer) {
		socket.Route("/world").Handler(func(conn *lemo.Socket, msg *lemo.Receive) func() *lemo.Error {
			log.Println(string(msg.Message.Event), string(msg.Message.Message))
			return nil
		})

		socket.Route("/kid").Handler(func(conn *lemo.Socket, msg *lemo.Receive) func() *lemo.Error {
			log.Println(string(msg.Message.Event), string(msg.Message.Message), atomic.AddInt32(&j, 1))
			return nil
		})
	})

	server.Group("/fuck").Handler(func(socket *lemo.SocketServer) {
		socket.Route("/:world").Handler(func(conn *lemo.Socket, msg *lemo.Receive) func() *lemo.Error {
			log.Println(string(msg.Message.Event), string(msg.Message.Message), msg.Params.ByName("world"))
			return nil
		})

		socket.Route("/kid").Handler(func(conn *lemo.Socket, msg *lemo.Receive) func() *lemo.Error {
			log.Println(string(msg.Message.Event), string(msg.Message.Message))
			return nil
		})
	})

	server.Start()

}

func WebSocketServer() {
	var server = &lemo.Server{Host: "0.0.0.0", Port: 12345}

	var socketHandler = &lemo.WebSocketServer{Path: "/", HeartBeatTimeout: 5, HeartBeatInterval: 3}

	// var socketBefore = []lemo.WebSocketBefore{
	// 	func(conn *lemo.WebSocket, msg *lemo.MessagePackage) (lemo.Context, func() *lemo.Error) {
	// 		return "hello111111111", nil
	// 	},
	// }

	socketHandler.Route("/:hello").Handler(func(conn *lemo.WebSocket, receive *lemo.Receive) func() *lemo.Error {
		var awesome = &awesomepackage.AwesomeMessage{}
		err := proto.Unmarshal(receive.Message.Message, awesome)

		if err != nil {
			return lemo.NewError(err)
		}

		logger.Log(receive.Message.Event, receive.Message.MessageType, receive.Message.ProtoType == lemo.ProtoBuf, awesome.AwesomeField, awesome.AwesomeKey)
		// _ = conn.JsonEmit(conn.Fd, lemo.JsonPackage{Event: "/haha", Message: "roland 这个傻吊"})
		_ = conn.Json(conn.Fd, lemo.M{"key": "roland", "type": "people"})
		return nil
	})

	socketHandler.OnMessage = func(conn *lemo.WebSocket, messageType int, msg []byte) {
		if len(msg) == 0 {
			return
		}

		// var awesome = &awesomepackage.AwesomeMessage{}
		// err := proto.Unmarshal(msg[9:], awesome)
		//
		// if err != nil {
		// 	logger.Log("marshaling error: ", err)
		// 	return
		// }

		logger.Log(string(msg))
	}

	socketHandler.OnClose = func(fd uint32) {
		logger.Log(fd, "is close")
	}

	socketHandler.OnError = func(err func() *lemo.Error) {
		logger.Log(err())
	}

	socketHandler.OnOpen = func(conn *lemo.WebSocket) {
		logger.Log(conn.Fd, "is open")

		for value := range conn.GetConnections() {
			log.Println(value.Fd)
		}

		log.Println(1)
	}

	var httpHandler = &lemo.HttpServer{}

	var before = []lemo.HttpServerBefore{
		func(t *lemo.Stream) (lemo.Context, func() *lemo.Error) {
			// _ = t.End("before")
			return nil, nil
		},
	}

	var after = []lemo.HttpServerAfter{
		func(t *lemo.Stream) func() *lemo.Error {
			// _ = t.End("after")
			return nil
		},
	}

	httpHandler.Get("/debug/pprof/").Handler(func(t *lemo.Stream) func() *lemo.Error {
		pprof.Index(t.Response, t.Request)
		return nil
	})
	// httpHandler.Get("/debug/pprof/:tip", pprof.Index)
	// httpHandler.Get("/debug/pprof/cmdline", pprof.Cmdline)
	// httpHandler.Get("/debug/pprof/profile", pprof.Profile)
	// httpHandler.Get("/debug/pprof/symbol", pprof.Symbol)
	// httpHandler.Get("/debug/pprof/trace", pprof.Trace)

	httpHandler.Group("/:hello").Handler(func(this *lemo.HttpServer) {
		this.Route("GET", "/:12").Before(before).After(after).Handler(func(t *lemo.Stream) func() *lemo.Error {
			var params = t.Params.ByName("hello")

			err := t.Json(lemo.M{"hello": params})

			return lemo.NewError(err)
		})
		this.Route("GET", "/xixi").Before(before).After(after).Handler(func(t *lemo.Stream) func() *lemo.Error {
			t.ParseFiles()

			for _, value := range t.Files.All() {
				for _, v := range value {
					f, _ := v.Open()
					d, _ := ioutil.ReadAll(f)
					log.Println(t.End(d))
					log.Println(v.Filename)
				}
			}

			// var params = t.Params.ByName("hello")
			//
			// err := t.Json(lemo.M{"hello": params})

			return nil
		})
	})

	httpHandler.OnError = func(ef func() *lemo.Error) {
		var err = ef()
		if err == nil {
			return
		}
		logger.Log(err)
	}

	server.Start(socketHandler, httpHandler)
}

func WebSocketClient() {
	var client = &lemo.WebSocketClient{Host: "127.0.0.1", Port: 12345, Path: "/", AutoHeartBeat: true, HeartBeatTimeout: 5, HeartBeatInterval: 3}

	client.Route("/haha").Handler(func(c *lemo.WebSocketClient, receive *lemo.Receive) func() *lemo.Error {
		logger.Log(receive.Message.Event, receive.Message.MessageType, receive.Message.ProtoType == lemo.Json, string(receive.Message.Message))
		return nil
	})

	client.OnOpen = func(c *lemo.WebSocketClient) {
		logger.Log("open")
		var data = &awesomepackage.AwesomeMessage{AwesomeField: "尼玛的", AwesomeKey: "他妈的"}
		logger.Log(c.ProtoBufEmit(lemo.ProtoBufPackage{Event: "/xixi", Message: data}))
	}

	client.OnMessage = func(c *lemo.WebSocketClient, messageType int, msg []byte) {
		logger.Log(string(msg))
	}

	client.OnError = func(err func() *lemo.Error) {
		logger.Err(err)
	}

	client.OnClose = func(c *lemo.WebSocketClient) {
		logger.Log("close")
	}

	client.Connect()
}
