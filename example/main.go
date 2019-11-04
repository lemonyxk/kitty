package main

import (
	"log"
	"time"

	"github.com/Lemo-yxk/lemo/container"
	"github.com/Lemo-yxk/lemo/utils"
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

	log.Println(utils.OrderID())
	time.Sleep(time.Millisecond * 10)
	log.Println(utils.OrderID())

}
