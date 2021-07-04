/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-06-06 15:44
**/

package main

import (
	"log"
	"time"

	"github.com/lemoyxk/kitty/socket"
	client3 "github.com/lemoyxk/kitty/socket/websocket/client"
)

func main() {
	run()
}

func run() {

	var client = &client3.Client{
		Scheme:            "ws",
		Addr:              "127.0.0.1:8667",
		HeartBeatTimeout:  time.Second * 2,
		HeartBeatInterval: time.Second,
	}

	client.OnClose = func(c *client3.Client) {
		log.Println("close")
	}

	client.OnOpen = func(c *client3.Client) {
		log.Println("open")
	}

	client.OnMessage = func(c *client3.Client, messageType int, msg []byte) {
		// log.Println(string(msg))
	}

	client.OnError = func(err error) {
		log.Println(err)
	}

	var router = &client3.Router{IgnoreCase: true}

	router.Group("/hello").Handler(func(handler *client3.RouteHandler) {
		handler.Route("/world").Handler(func(c *client3.Client, stream *socket.Stream) error {
			log.Println(string(stream.Data))
			return nil
		})
	})

	// go func() {
	// 	var ticker = time.NewTicker(time.Second)
	// 	for range ticker.C {
	// 		stream, err := client.AsyncJson(socket.JsonPackage{
	// 			Event: "/hello/world",
	// 			Data:  strings.Repeat("hello world!", 1),
	// 		})
	//
	// 		log.Println(string(stream.Message), err)
	// 	}
	// }()

	go client.SetRouter(router).Connect()

	select {}

}
