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
	"strconv"
)

type str int

const String str = iota

func (s str) Join(v []interface{}, sep string) string {

	var buf bytes.Buffer

	for i := 0; i < len(v); i++ {
		switch v[i].(type) {
		case int:
			buf.WriteString(strconv.Itoa(v[i].(int)))
		case float64:
			buf.WriteString(strconv.FormatFloat(v[i].(float64), 'f', -1, 64))
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
