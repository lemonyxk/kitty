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

	"github.com/json-iterator/go"
)

var a = []byte(`{"a":1}`)
var b M

func BenchmarkStdAppLogs_normal_jsoniter(b *testing.B) {
	for j := 0; j <= b.N; j++ {
		jsoniter.Unmarshal(a, &b)
	}
}

func BenchmarkStdAppLogs_normal_json(b *testing.B) {

	for j := 0; j <= b.N; j++ {
		json.Unmarshal(a, &b)
	}
}
