package main

import (
	"log"
	"os"
	"time"

	"github.com/Lemo-yxk/lemo/container"
	"github.com/Lemo-yxk/lemo/logger"
	"github.com/Lemo-yxk/lemo/utils"
)

func main() {

	var q = container.NewBlockQueue()

	go func() {
		for {
			logger.Console(q.Get())
		}
	}()

	for {
		time.Sleep(time.Second)
		q.Put(time.Now())
	}

	utils.ListenSignal(func(sig os.Signal) {
		log.Println(sig)
	})
}
