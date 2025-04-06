/**
* @program: kitty
*
* @create: 2024-12-14 15:03
**/

package json

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"io"
)

func NewDecoder(reader io.Reader) sonic.Decoder {
	return sonic.ConfigFastest.NewDecoder(reader)
}

func Unmarshal(data []byte, v interface{}) error {
	return sonic.ConfigFastest.Unmarshal(data, v)
}

func Get(data []byte, path ...interface{}) (ast.Node, error) {
	return sonic.Get(data, path...)
}
