/**
* @program: kitty
*
* @create: 2024-12-14 15:03
**/

package json

import (
	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	"io"
)

func NewDecoder(reader io.Reader) *jsoniter.Decoder {
	return jsoniter.ConfigFastest.NewDecoder(reader)
}

func Unmarshal(data []byte, v interface{}) error {
	if len(data) < 1024*3 { // 3KB
		return sonic.ConfigFastest.Unmarshal(data, v)
	}
	return jsoniter.ConfigFastest.Unmarshal(data, v)
}
