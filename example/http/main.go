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
	"bytes"
	"embed"
	"io"
	"log"
	http2 "net/http"
	"os"
	"time"

	"github.com/lemonyxk/kitty"
	"github.com/lemonyxk/kitty/errors"
	hello "github.com/lemonyxk/kitty/example/protobuf"
	kitty2 "github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/http"
	"github.com/lemonyxk/kitty/socket/http/client"
	"github.com/lemonyxk/kitty/socket/http/server"
	"google.golang.org/protobuf/proto"
)

//go:embed public/**
var fileSystem embed.FS

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func runHttpServer() {

	var ready = make(chan struct{})

	var httpServer = kitty.NewHttpServer[any]("127.0.0.1:8666")
	// use ssl for https
	// httpServer.CertFile = "example/ssl/localhost+2.pem"
	// httpServer.KeyFile = "example/ssl/localhost+2-key.pem"

	var httpServerRouter = kitty.NewHttpServerRouter[any]()

	var httpStaticServerRouter = kitty.NewHttpServerStaticRouter()

	// middleware
	httpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream[server.Conn]) {
			// cors
			stream.Response.Header().Set("Access-Control-Allow-Origin", "*")
			stream.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			stream.Response.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if stream.Request.Method == "OPTIONS" {
				stream.Response.WriteHeader(http2.StatusOK)
				return
			}

			stream.Parser.Auto()

			log.Println("middleware1 start")
			next(stream)
			log.Println("middleware1 end")
		}
	})

	httpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream[server.Conn]) {
			log.Println("middleware2 start")
			next(stream)
			log.Println("middleware2 end")
		}
	})

	var before = func(stream *http.Stream[server.Conn]) error {
		log.Println("before start")
		// you could return error to stop the stream
		return nil
	}

	var after = func(stream *http.Stream[server.Conn]) error {
		log.Println("after start")
		// handle this error by set OnError
		return errors.New("after error")
	}

	httpServer.OnError = func(stream *http.Stream[server.Conn], err error) {
		// %+v print stack
		// log.Printf("%+v", err)
		log.Println(err)
	}

	// you cloud create your own router to use get and post or other method easily
	var httpRouter = httpServerRouter.Create()
	httpRouter.Get("/hello").Before(before).After(after).Handler(func(stream *http.Stream[server.Conn]) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		var t = map[string]interface{}{}
		log.Println(stream.Json.Decode(&t))
		return stream.Sender.String("hello world!")
	})

	type Request struct {
		Name string `json:"name" validate:"required"`
		Addr string `json:"addr" validate:"required"`
		Age  int    `json:"age" validate:"required,gte:0,lte:100"`
	}

	httpRouter.Post("/json").Before(before).After(after).Handler(func(stream *http.Stream[server.Conn]) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)

		var request Request
		var err = stream.Json.Validate(&request)
		if err != nil {
			return stream.Sender.Json(err.Error())
		}

		return stream.Sender.Json(request)
	})

	httpRouter.Post("/post").Before(before).After(after).Handler(func(stream *http.Stream[server.Conn]) error {
		log.Println(stream.Form.String())
		return stream.Sender.String("hello world!")
	})

	httpRouter.Post("/file").Before(before).After(after).Handler(func(stream *http.Stream[server.Conn]) error {
		log.Println(stream.Files.String())
		log.Println(stream.Form.String())
		return stream.Sender.String("hello world!")
	})

	httpRouter.Post("/OctetStream").Before(before).After(after).Handler(func(stream *http.Stream[server.Conn]) error {
		var b bytes.Buffer
		var body = stream.Request.Body
		log.Println("stream.Request.ContentLength:", stream.Request.ContentLength)
		time.Sleep(time.Second * 3)
		i, err := io.Copy(&b, body)
		if err != nil {
			return stream.Sender.String(err.Error())
		}
		log.Println(i, b.Len())
		return stream.Sender.String("hello world!")
	})

	httpRouter.Post("/test").Handler(func(stream *http.Stream[server.Conn]) error {
		return stream.Sender.String("hello world!")
	})

	httpRouter.Delete("/delete").Handler(func(stream *http.Stream[server.Conn]) error {
		log.Println(stream.Form.String())
		return stream.Sender.String("delete hello world!")
	})

	// or you can just use original router
	httpServerRouter.Method("POST").Route("/proto").Handler(func(stream *http.Stream[server.Conn]) error {
		log.Println("addr:", stream.Request.RemoteAddr, stream.Request.Host)
		var res hello.AwesomeMessage
		var msg = stream.Protobuf.Bytes()
		var err = proto.Unmarshal(msg, &res)
		if err != nil {
			return stream.Sender.String(err.Error())
		}
		log.Printf("%+v", res.String())
		return stream.Sender.String("proto hello world!")
	})

	// create group router
	var group = httpServerRouter.Group("/hello").Create()
	group.Get("/world").Handler(func(t *http.Stream[server.Conn]) error {
		time.Sleep(time.Second * 3)
		return t.Sender.Any(os.Getpid())
	})

	// another way to use group router
	httpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*http.Stream[server.Conn], any]) {
		handler.Get("/hello").Handler(func(t *http.Stream[server.Conn]) error {
			return t.Sender.Any(os.Getpid())
		})
	})

	httpServerRouter.Group("/hello").Handler(func(handler *router.Handler[*http.Stream[server.Conn], any]) {
		handler.Group("/hello").Handler(func(handler *router.Handler[*http.Stream[server.Conn], any]) {
			handler.Get("/hello").Handler(func(t *http.Stream[server.Conn]) error {
				return t.Sender.Any(os.Getpid())
			})
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
	// httpStaticServerRouter.SetStaticDownload(true)
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

	var msg = hello.AwesomeMessage{
		AwesomeField: "1",
		AwesomeKey:   "2",
	}

	var res = client.Post("http://127.0.0.1:8666/proto").Protobuf(&msg).Send()
	if res.Error() != nil {
		log.Println(res.Error())
	} else {
		log.Println("res:", res.String())
	}

	res = client.Delete("http://127.0.0.1:8666/delete?b=2").Form(kitty2.M{"a": 1}).Send()
	if res.Error() != nil {
		log.Println(res.Error())
	} else {
		log.Println("res:", res.String())
	}

	var h = client.Get("http://127.0.0.1:8666/hello/world").Query(kitty2.M{"a": 1})

	time.AfterFunc(time.Second, func() {
		h.Abort()
	})

	go func() {
		res = h.Send()
		if res.Error() != nil {
			log.Println(res.Error())
		}

		log.Println("res:", res.String())
	}()

	go func() {
		var f, err = os.Open("./new.go")
		if err != nil {
			panic(err)
		}

		defer func() { _ = f.Close() }()

		var res = client.Post("http://127.0.0.1:8666/file").Multipart(kitty2.M{
			"file": f, "a": 1,
		}).Send()

		if res.Error() != nil {
			log.Println(res.Error())
		} else {
			log.Println("res:", res.String())
		}

	}()

	go func() {
		var f, err = os.Open("./new.go")
		if err != nil {
			panic(err)
		}

		defer func() { _ = f.Close() }()

		var res = client.Post("http://127.0.0.1:8666/OctetStream").OctetStream(f).Send()

		if res.Error() != nil {
			log.Println(res.Error())
		} else {
			log.Println("res:", res.String())
		}

	}()

	go func() {
		var f, err = os.Open("./new.go")
		if err != nil {
			panic(err)
		}

		defer func() { _ = f.Close() }()

		var res = client.Post("http://127.0.0.1:8666/json").Json(kitty2.M{"name": "lemon", "addr": "shanghai", "age": 18}).Send()

		if res.Error() != nil {
			log.Println(res.Error())
		} else {
			log.Println("res:", res.Bytes())
		}

	}()

	select {}
}
