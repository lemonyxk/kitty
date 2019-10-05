package main

import (
	"log"
	"net/http/pprof"

	"github.com/Lemo-yxk/lemo"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	var server = &lemo.Server{Host: "0.0.0.0", Port: 12345}

	var socketHandler = &lemo.Socket{Path: "/"}

	// var socketBefore = []lemo.WebSocketBefore{
	// 	func(conn *lemo.Connection, msg *lemo.MessagePackage) (lemo.Context, func() *lemo.Error) {
	// 		return "hello111111111", nil
	// 	},
	// }

	// socketHandler.SetRouter("/:hello", socketBefore, func(conn *lemo.Connection, receive *lemo.Receive) func() *lemo.Error {
	// 	log.Println(receive.Context)
	// 	return nil
	// })

	log.Println([]byte("\r\n"))

	socketHandler.OnMessage = func(conn *lemo.Connection, messageType int, msg []byte) {
		if len(msg) == 0 {
			return
		}
		// var awesome = &awesomepackage.AwesomeMessage{}
		// err := proto.Unmarshal(msg, awesome)
		// if err != nil {
		// 	log.Println("marshaling error: ", err)
		// 	return
		// }
		log.Println(conn.Fd, messageType, msg)
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
