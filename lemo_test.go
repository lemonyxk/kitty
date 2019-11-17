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
	"github.com/json-iterator/go"
	"testing"
)

type StdAppLogs_normal struct {
	Level   int32   `json:"l"`
	Message string  `json:"m"`
	Time    float32 `json:"t"`
	//Context1 map[string]jsonpkg.RawMessage `json:"ctx1"`
	//Context2 jsonpkg.RawMessage            `json:"ctx2"`
	//Context3 []byte                        `json:"ctx3"`
	//Context4 Context                       `json:"ctx4"`
}

var (
	// 一般类型的测试，包括int32, string, float32, 其他格式均不关心的格式
	datas_normal = []*StdAppLogs_normal{
		// 所有字段都存在且争取
		{Level: 1, Message: "message", Time: 1.0},
		// 缺少int32类型
		{Message: "message", Time: 1.0},
		// 缺少string类型
		{Level: 1, Time: 1.0},
		// 缺少float32类型
		{Level: 1, Message: "message"},
		// 缺少所有一般类型
		{},
	}
)

func BenchmarkStdAppLogs_normal_jsoniter(b *testing.B) {

	for j := 0; j <= b.N; j++ {
		for _, data := range datas_normal {
			// 解析成json格式字符串
			dataByte, _ := jsoniter.Marshal(data)

			// 再反解回结构体
			appLog := StdAppLogs_normal{}
			_ = jsoniter.Unmarshal(dataByte, &appLog)
		}
	}
}

func BenchmarkStdAppLogs_normal_json(b *testing.B) {

	for j := 0; j <= b.N; j++ {
		for _, data := range datas_normal {
			// 解析成json格式字符串
			dataByte, _ := json.Marshal(data)

			// 再反解回结构体
			appLog := StdAppLogs_normal{}
			_ = json.Unmarshal(dataByte, &appLog)
		}
	}
}
