/**
* @program: kitty
*
* @create: 2024-12-14 15:03
**/

package json

import (
	"github.com/bytedance/sonic"
	"io"
)

func NewDecoder(reader io.Reader) sonic.Decoder {
	return sonic.ConfigFastest.NewDecoder(reader)
}

func Unmarshal(data []byte, v interface{}) error {
	return sonic.ConfigFastest.Unmarshal(data, v)
}
