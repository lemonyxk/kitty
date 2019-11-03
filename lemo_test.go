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
	"testing"

	"github.com/Lemo-yxk/lemo/container"
)

var q = container.NewLastPool(container.LastPoolConfig{
	Max: 4,
	Min: 4,
	New: func() interface{} {
		return 1
	},
})

func BenchmarkParseMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q.Put(1)
		q.Get()
	}
}
