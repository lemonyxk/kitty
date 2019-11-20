package main

import (
	"fmt"
	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/exception"
)

func main() {

	fmt.Println(test())

	var server = lemo.Server{Host: "0.0.0.0", Port: 8666}

	var httpServer = lemo.HttpServer{}

	httpServer.Group("/hello").Handler(func(this *lemo.HttpServer) {
		this.Get("/world").Handler(func(t *lemo.Stream) func() *exception.Error {
			return exception.New(t.Json("hello"))
		})
	})

	for key, value := range httpServer.GetAllRouters() {
		fmt.Println(key, value.Info, string(value.Route))
	}

	server.Start(nil, &httpServer)

}

func test() error {
	return exception.New("hello")()
}
