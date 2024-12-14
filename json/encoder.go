/**
* @program: kitty
*
* @create: 2024-12-14 15:00
**/

package json

import (
	"github.com/bytedance/sonic"
	"io"
)

func NewEncoder(writer io.Writer) sonic.Encoder {
	return sonic.ConfigFastest.NewEncoder(writer)
}

func Marshal(v interface{}) ([]byte, error) {
	return sonic.ConfigFastest.Marshal(v)
}
