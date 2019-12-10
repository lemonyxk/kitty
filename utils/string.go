/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-10 21:23
**/

package utils

import (
	"bytes"
	"fmt"
)

type str int

const String str = iota

func (s str) Join(v []interface{}, sep string) string {

	var buf bytes.Buffer

	for i := 0; i < len(v); i++ {
		switch v[i].(type) {
		case string:
			buf.WriteString(v[i].(string))
		default:
			buf.WriteString(fmt.Sprintf("%s", v[i]))
		}
		if i != len(v)-1 {
			buf.WriteString(sep)
		}
	}

	return buf.String()
}
