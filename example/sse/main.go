/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-24 03:15
**/

package main

import (
	"github.com/lemonyxk/kitty"
	kitty2 "github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/socket/http"
	"github.com/lemonyxk/kitty/socket/http/server"
	"log"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func runHttpServer() {

	var ready = make(chan struct{})

	var httpServer = kitty.NewHttpServer[any]("127.0.0.1:8999")
	// use ssl for https
	httpServer.CertFile = "example/ssl/127.0.0.1+1.pem"
	httpServer.KeyFile = "example/ssl/127.0.0.1+1-key.pem"

	var httpServerRouter = kitty.NewHttpServerRouter[any]()

	var httpStaticServerRouter = kitty.NewHttpServerStaticRouter()

	// middleware
	httpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream[server.Conn]) {
			stream.Response.Header().Set("Access-Control-Allow-Origin", "*")
			stream.Parser.Auto()
			next(stream)
		}
	})

	httpServer.OnError = func(stream *http.Stream[server.Conn], err error) {
		log.Println(err)
	}

	// you cloud create your own router to use get and post or other method easily
	var httpRouter = httpServerRouter.Create()
	httpRouter.Get("/sse").Handler(func(stream *http.Stream[server.Conn]) error {

		var sse, err = stream.UpgradeSse(&http.SseConfig{Retry: time.Second * 3})
		if err != nil {
			return err
		}

		go func() {
			for !sse.IsClose() {
				time.Sleep(time.Second * 1)
				log.Println(sse.LasTEventID)
				if err := sse.Json(kitty2.M{"a": 1, "b": 2}); err != nil {
					log.Println(err)
				}
			}
		}()

		return sse.Wait()
	})

	go httpServer.
		SetStaticRouter(httpStaticServerRouter).
		SetRouter(httpServerRouter).
		Start()

	<-ready
}

func main() {
	runHttpServer()
	select {}
}
