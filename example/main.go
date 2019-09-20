package main

import (
	"log"
	"strings"

	"github.com/Lemo-yxk/ws"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	var server = &ws.Server{Host: "127.0.0.1", Port: 12345, Path: "/Game-Robot"}

	var socketHandler = &ws.Socket{}

	socketHandler.SetRouter("hello1", func(conn *ws.Connection, ftd *ws.Fte, msg []byte) {
		log.Println(ftd.Fd)
	})

	socketHandler.OnClose = func(conn *ws.Connection) {
		log.Println(conn.Fd, "is close")
	}

	socketHandler.OnError = func(err error) {
		log.Println(err)
	}

	socketHandler.OnOpen = func(conn *ws.Connection) {
		log.Println(conn.Fd, "is open")
	}

	var httpHandler = &ws.Http{}

	httpHandler.Group("/hello",
		[]ws.Before{
			func(t *ws.Stream) (i interface{}, e error) {
				// log.Println("before group")
				_ = t.End("group")
				return nil, nil
			},
		},
		func() {
			httpHandler.Get("/2", []ws.After{
				func(t *ws.Stream) error {
					_ = t.End("after")
					return nil
				},
			}, func(t *ws.Stream) {
				_ = t.End(t.End(strings.Repeat("bow", 10000)))
			})
		})

	log.Println(ws.PassAfter, ws.PassBefore, ws.ForceAfter, ws.ForceBefore)

	server.Start(socketHandler, httpHandler)

}
