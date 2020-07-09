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
	"net/http"
	"testing"

	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
)

var t2 http.Server

func init() {
	console.Info(exception.IsNil(t2))
}

func BenchmarkStdAppLogs_normal_jsoniter(b *testing.B) {
	for j := 0; j <= b.N; j++ {
		exception.IsNil(t2)
	}
}
