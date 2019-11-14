package main

import (
	"github.com/Lemo-yxk/lemo/container/head"
	"github.com/Lemo-yxk/lemo/logger"
)

type People struct {
	value int
}

func (p *People) Value() int {
	return p.value
}

func main() {

	var minHead = head.NewMinHead(&People{0}, &People{1}, &People{2})

	logger.Println(minHead.Pop())
	logger.Println(minHead.Pop())
	logger.Println(minHead.Pop())

}
