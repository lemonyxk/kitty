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
	"strings"
	"time"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
	client2 "github.com/Lemo-yxk/lemo/tcp/client"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {
	run()
}

func run() {

	var client = &client2.Client{Host: "0.0.0.0", Port: 8888, Reconnect: true, AutoHeartBeat: true}

	client.OnClose = func(c *client2.Client) {
		console.Log("close")
	}

	client.OnOpen = func(c *client2.Client) {
		console.Log("open")
	}

	client.OnMessage = func(c *client2.Client, messageType int, msg []byte) {
		// console.Log(string(msg))
	}

	client.OnError = func(err exception.Error) {
		console.Error(err)
	}

	var router = &client2.Router{IgnoreCase: true}

	router.Group("/hello").Handler(func(handler *client2.RouteHandler) {
		handler.Route("/world").Handler(func(c *client2.Client, receive *lemo.Receive) exception.Error {
			console.Log(string(receive.Body.Message))
			return nil
		})
	})

	go func() {
		utils.Time.Ticker(time.Second, func() {
			_ = client.JsonEmit(lemo.JsonPackage{
				Event: "/hello/world",
				Data:  strings.Repeat("hello world!", 1),
			})
		}).Start()
	}()

	client.SetRouter(router).Connect()

}
