/**
* @program: kitty
*
* @create: 2024-12-14 15:03
**/

package json

import (
	jsoniter "github.com/json-iterator/go"
	"io"
)

func NewDecoder(reader io.Reader) *jsoniter.Decoder {
	return jsoniter.ConfigFastest.NewDecoder(reader)
}

func Unmarshal(data []byte, v interface{}) error {
	return jsoniter.ConfigFastest.Unmarshal(data, v)
}
