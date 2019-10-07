package main

import (
	"log"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/Lemo-yxk/lemo"
	awesomepackage "github.com/Lemo-yxk/lemo/protobuf-origin"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	go Server()

	time.Sleep(time.Second)

	Client()

	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞
	<-signalChan

}

func Server() {
	var server = &lemo.Server{Host: "0.0.0.0", Port: 12345}

	var socketHandler = &lemo.Socket{Path: "/"}

	// var socketBefore = []lemo.WebSocketBefore{
	// 	func(conn *lemo.Connection, msg *lemo.MessagePackage) (lemo.Context, func() *lemo.Error) {
	// 		return "hello111111111", nil
	// 	},
	// }

	socketHandler.SetRouter("/:hello", func(conn *lemo.Connection, receive *lemo.Receive) func() *lemo.Error {

		var awesome = &awesomepackage.AwesomeMessage{}
		err := proto.Unmarshal(receive.Message.Message, awesome)

		if err != nil {
			return lemo.NewError(err)
		}

		log.Println(receive.Message.Event, receive.Message.MessageType, receive.Message.FormatType == lemo.ProtoBuf, awesome.AwesomeField, awesome.AwesomeKey)
		_ = conn.ProtoBufEmit(conn.Fd, lemo.ProtoBufPackage{Event: "/haha", Message: &awesomepackage.AwesomeMessage{AwesomeKey: "????", AwesomeField: "!!!!"}})
		return nil
	})

	socketHandler.OnMessage = func(conn *lemo.Connection, messageType int, msg []byte) {
		if len(msg) == 0 {
			return
		}

		// var awesome = &awesomepackage.AwesomeMessage{}
		// err := proto.Unmarshal(msg[9:], awesome)
		//
		// if err != nil {
		// 	log.Println("marshaling error: ", err)
		// 	return
		// }

		log.Println(msg)
	}

	socketHandler.OnClose = func(fd uint32) {
		log.Println(fd, "is close")
	}

	socketHandler.OnError = func(err func() *lemo.Error) {
		log.Println(err())
	}

	socketHandler.OnOpen = func(conn *lemo.Connection) {
		log.Println(conn.Fd, "is open")
	}

	var httpHandler = &lemo.Http{}

	var before = []lemo.HttpBefore{
		func(t *lemo.Stream) (lemo.Context, func() *lemo.Error) {
			_ = t.End("before")
			return nil, nil
		},
	}

	var after = []lemo.HttpAfter{
		func(t *lemo.Stream) func() *lemo.Error {
			_ = t.End("after")
			return nil
		},
	}

	httpHandler.Get("/debug/pprof/", pprof.Index)
	httpHandler.Get("/debug/pprof/:tip", pprof.Index)
	httpHandler.Get("/debug/pprof/cmdline", pprof.Cmdline)
	httpHandler.Get("/debug/pprof/profile", pprof.Profile)
	httpHandler.Get("/debug/pprof/symbol", pprof.Symbol)
	httpHandler.Get("/debug/pprof/trace", pprof.Trace)

	httpHandler.Group("/:hello", func() {
		httpHandler.Get("/:12", before, after, func(t *lemo.Stream) func() *lemo.Error {

			var params = t.Params.ByName("hello")

			err := t.Json(lemo.M{"hello": params})

			return lemo.NewError(err)
		})
	})

	httpHandler.OnError = func(ef func() *lemo.Error) {
		var err = ef()
		if err == nil {
			return
		}
		log.Println(err)
	}

	server.Start(socketHandler, httpHandler)
}

func Client() {
	var client = &lemo.Client{Host: "127.0.0.1", Port: 12345, Path: "/", HandshakeTimeout: 10, AutoHeartBeat: true}

	client.InitRouter()

	client.SetRouter("/haha", func(c *lemo.Client, receive *lemo.ReceivePackage) func() *lemo.Error {

		var awesome = &awesomepackage.AwesomeMessage{}
		err := proto.Unmarshal(receive.Message, awesome)

		if err != nil {
			return lemo.NewError(err)
		}

		log.Println(receive.Event, receive.MessageType, receive.FormatType == lemo.ProtoBuf, awesome.AwesomeField, awesome.AwesomeKey)

		return nil
	})

	client.OnOpen = func(c *lemo.Client) {
		log.Println("open")

		var data = &awesomepackage.AwesomeMessage{AwesomeField: "尼玛的", AwesomeKey: "他妈的"}

		log.Println(c.ProtoBufEmit(lemo.ProtoBufPackage{Event: "/xixi", Message: data}))
	}

	client.OnMessage = func(c *lemo.Client, messageType int, msg []byte) {
		log.Println(string(msg))
	}

	client.OnError = func(err func() *lemo.Error) {
		log.Println(err())
	}

	client.OnClose = func(c *lemo.Client) {
		log.Println("close")
	}

	client.Connect()
}
