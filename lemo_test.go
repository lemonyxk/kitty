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
	"encoding/json"
	"testing"
)

var data = M{"data": "word", "event": "hello"}

var res, _ = json.Marshal(data)

func BenchmarkParseMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s, bytes := ParseMessage(res)
		if s != "hello" {
			b.Error("error")
		}

		if string(bytes) != "word" {
			b.Error("word", string(bytes))
		}
	}
}
