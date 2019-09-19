package main

import (
	"github.com/Lemo-yxk/ws"
	"log"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	var server = &ws.Server{Host: "127.0.0.1", Port: 5858, Path: "/Game-Robot"}

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

	httpHandler.Group("/hello", []ws.Before{
		func(t *ws.Stream) (i interface{}, e error) {
			log.Println("before1")
			return nil, nil
		},
		func(t *ws.Stream) (i interface{}, e error) {
			log.Println("before2")
			return nil, nil
		},
	}, func() {
		httpHandler.Get("/1", func(t *ws.Stream) {
			log.Println("now1")
			_ = t.Json("hello1")
		})
		httpHandler.Get("/2", func(t *ws.Stream) {
			log.Println("now2")
			_ = t.End("hello2")
		})
	}, []ws.After{
		func(t *ws.Stream) error {
			log.Println("after1")
			return nil
		},
		func(t *ws.Stream) error {
			log.Println("after2")
			return nil
		},
	})

	server.Start(ws.WebSocket(socketHandler), httpHandler)

}
