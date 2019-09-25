package main

import (
	"log"
	"net"

	"github.com/Lemo-yxk/lemo"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	var server = &lemo.Server{Host: "127.0.0.1", Port: 12345, Path: "/Game-Robot"}

	var socketHandler = &lemo.Socket{}

	socketHandler.SetRouter("hello1", func(conn *lemo.Connection, ftd *lemo.Fte, msg []byte) {
		log.Println(ftd.Fd)
	})

	socketHandler.OnClose = func(conn *lemo.Connection) {
		log.Println(conn.Fd, "is close")
	}

	socketHandler.OnError = func(err error) {
		log.Println(err)
	}

	socketHandler.OnOpen = func(conn *lemo.Connection) {
		log.Println(conn.Fd, "is open")
	}

	var httpHandler = &lemo.Http{}

	var before = []lemo.Before{
		func(t *lemo.Stream) (i interface{}, e error) {
			_ = t.End("before")
			return nil, nil
		},
	}

	var after = []lemo.After{
		func(t *lemo.Stream) (e error) {
			_ = t.End("after")
			return nil
		},
	}

	httpHandler.Group("/:hello", func() {
		httpHandler.Get("/:12/1/:22/world/:xixi/", before, after, func(t *lemo.Stream) {
			_ = t.End(t.Params.ByName("xixi"))
		})
	})

	log.Println(net.InterfaceAddrs())

	server.Start(socketHandler, httpHandler)

}
