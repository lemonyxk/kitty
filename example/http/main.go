/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-24 03:15
**/

package main

import (
	"embed"
	"log"
	http2 "net/http"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/errors"
	awesomepackage "github.com/lemonyxk/kitty/v2/example/protobuf"
	"github.com/lemonyxk/kitty/v2/http"
	"github.com/lemonyxk/kitty/v2/http/client"
	server3 "github.com/lemonyxk/kitty/v2/http/server"
	router2 "github.com/lemonyxk/kitty/v2/router"
)

//go:embed public/**
var fileSystem embed.FS

func runHttpServer() {

	var ready = make(chan struct{})

	var httpServer = kitty.NewHttpServer("127.0.0.1:8666")
	// use ssl for https
	// httpServer.CertFile = "/Users/lemon/test/go/localhost+2.pem"
	// httpServer.KeyFile = "/Users/lemon/test/go/localhost+2-key.pem"

	var httpServerRouter = kitty.NewHttpServerRouter()

	var httpStaticServerRouter = kitty.NewHttpServerStaticRouter()

	// middleware
	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
			log.Println("middleware1 start")
			next(stream)
			log.Println("middleware1 end")
		}
	})

	httpServer.Use(func(next server3.Middle) server3.Middle {
		return func(stream *http.Stream) {
			log.Println("middleware2 start")
			next(stream)
			log.Println("middleware2 end")
		}
	})

	var before = func(stream *http.Stream) error {
		log.Println("before start")
		// you could return error to stop the stream
		return nil
	}

	var after = func(stream *http.Stream) error {
		log.Println("after start")
		// handle this error by set OnError
		return errors.NewWithStack("after error")
	}

	httpServer.OnError = func(stream *http.Stream, err error) {
		// %+v print stack
		// log.Printf("%+v", err)
		log.Println(err)
	}

	// you cloud create your own router to use get and post or other method easily
	var router = httpServerRouter.Create()
	router.Get("/hello").Before(before).After(after).Handler(func(stream *http.Stream) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		return stream.EndString("hello world!")
	})

	// or you can just use original router
	httpServerRouter.RouteMethod("POST", "/proto").Handler(func(stream *http.Stream) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		log.Println(stream.AutoGet("name").String())
		var res awesomepackage.AwesomeMessage
		var msg = stream.Protobuf.Bytes()
		var err = proto.Unmarshal(msg, &res)
		if err != nil {
			return stream.EndString(err.Error())
		}
		log.Printf("%+v", res)
		return stream.EndString("hello proto!")
	})

	// create group router
	var group = httpServerRouter.Group("/hello").Create()
	group.Get("/world").Handler(func(t *http.Stream) error {
		return t.JsonFormat("SUCCESS", 200, os.Getpid())
	})

	// another way to use group router
	httpServerRouter.Group("/hello").Handler(func(handler *router2.Handler[*http.Stream]) {
		handler.Get("/hello").Handler(func(t *http.Stream) error {
			return t.JsonFormat("SUCCESS", 200, os.Getpid())
		})
	})

	httpServer.OnSuccess = func() {
		ready <- struct{}{}
		log.Println(httpServer.LocalAddr())
	}

	// first file system
	httpStaticServerRouter.SetStaticPath("/", "public", http2.FS(fileSystem))
	// second file system
	httpStaticServerRouter.SetStaticPath("/protobuf", "protobuf", http2.Dir("./example"))

	// allow access directory in first and second file system
	httpStaticServerRouter.SetOpenDir(0, 1)

	go httpServer.
		SetStaticRouter(httpStaticServerRouter).
		SetRouter(httpServerRouter).
		Start()

	<-ready
}

func main() {
	runHttpServer()

	var msg = awesomepackage.AwesomeMessage{
		AwesomeField: "1",
		AwesomeKey:   "2",
	}

	var res = client.Post("http://127.0.0.1:8666/proto").Protobuf(&msg).Send()
	if res.LastError() != nil {
		log.Println(res.LastError())
	}

	select {}
}
