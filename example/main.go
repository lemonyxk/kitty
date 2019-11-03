package main

import (
	"log"

	"github.com/Lemo-yxk/lemo/container"
)

func main() {

	var q = container.NewLastPool(container.LastPoolConfig{
		Max: 0,
		Min: 0,
		New: func() interface{} {
			return 1
		},
	})

	q.Put(1)
	log.Println(q.Get())
}
