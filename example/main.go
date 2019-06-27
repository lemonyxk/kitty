package main

import (
	"github.com/Lemo-yxk/ws"
	"log"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {

	var webSocket = &ws.Server{Host: "127.0.0.1", Port: 5858, Path: "/Game-Robot"}

	var handlerSocket = &ws.Socket{}

	handlerSocket.InitRouter()

	handlerSocket.SetRouter("hello1", func(conn *ws.Connection, message *ws.Message, context interface{}) {
		log.Println(message.Fd)
	})

	handlerSocket.OnClose = func(conn *ws.Connection) {
		log.Println(conn.Fd, "is close")
	}

	handlerSocket.OnError = func(err error) {
		log.Println(err)
	}

	handlerSocket.OnOpen = func(conn *ws.Connection) {
		log.Println(conn.Fd, "is open")
	}

	webSocket.Start(ws.WebSocket(handlerSocket), nil)

}
