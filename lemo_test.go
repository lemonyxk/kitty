/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-05 14:19
**/

package lemo

import (
	head2 "github.com/Lemo-yxk/lemo/container/head"
	"testing"
)

type People struct {
	value int
}

func (p *People) Value() int {
	return p.value
}

var head = head2.NewMinHead(&People{0}, &People{1}, &People{2})

func BenchmarkParseMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//head.Push(1)
		head.Pop()
	}
}
