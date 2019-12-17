/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-17 14:26
**/

package String

import (
	"bytes"
	"strings"
)

type String struct {
	byt bytes.Buffer
}

func New(str ...string) *String {
	var s = &String{}
	for i := 0; i < len(str); i++ {
		s.byt.WriteString(str[i])
	}
	return s
}

func (s *String) String() string {
	return s.byt.String()
}

func (s *String) Bytes() []byte {
	return s.Bytes()
}

func (s *String) Split(sep string) []string {
	return strings.Split(s.byt.String(), sep)
}

func (s *String) Add(str string) *String {
	s.byt.WriteString(str)
	return s
}
