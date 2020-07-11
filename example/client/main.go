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
	"strings"
	"time"

	kitty "github.com/lemoyxk/kitty"
	client3 "github.com/lemoyxk/kitty/websocket/client"

	client2 "github.com/lemoyxk/kitty/tcp/client"
)

func main() {
	run()
}

func run() {

	_ = client3.Client{}

	var client = &client2.Client{Host: "0.0.0.0:", Reconnect: true, AutoHeartBeat: true}

	client.OnClose = func(c *client2.Client) {
		log.Println("close")
	}

	client.OnOpen = func(c *client2.Client) {
		log.Println("open")
	}

	client.OnMessage = func(c *client2.Client, messageType int, msg []byte) {
		// log.Println(string(msg))
	}

	client.OnError = func(err error) {
		log.Println(err)
	}

	var router = &client2.Router{IgnoreCase: true}

	router.Group("/hello").Handler(func(handler *client2.RouteHandler) {
		handler.Route("/world").Handler(func(c *client2.Client, receive *kitty.Receive) error {
			log.Println(string(receive.Body.Message))
			return nil
		})
	})

	go func() {
		var ticker = time.NewTicker(time.Second)
		for range ticker.C {
			_ = client.JsonEmit(kitty.JsonPackage{
				Event: "/hello/world",
				Data:  strings.Repeat("hello world!", 1),
			})
		}
	}()

	go client.SetRouter(router).Connect()

	select {}

}
