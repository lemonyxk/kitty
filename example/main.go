package main

import (
	"fmt"

	"github.com/Lemo-yxk/lemo"
)

func main() {

	fmt.Println(test())

	var server = lemo.Server{Host: "0.0.0.0", Port: 8666}

	var httpServer = lemo.HttpServer{}

	httpServer.Group("/hello").Handler(func(this *lemo.HttpServer) {
		this.Get("/world").Handler(func(t *lemo.Stream) func() *lemo.Error {
			return lemo.NewError(t.Json("hello"))
		})
	})

	for key, value := range httpServer.GetAllRouters() {
		fmt.Println(key, value.Info, string(value.Route))
	}

	server.Start(nil, &httpServer)

}

func test() error {
	return lemo.NewError("hello")()
}
