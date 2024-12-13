/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:33
**/

package http

import (
	"bytes"
	json "github.com/bytedance/sonic"
)

type Json struct {
	buf *bytes.Buffer
}

func (j *Json) Reset(data any) error {
	j.buf.Reset()
	var bts, err = json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = j.buf.Write(bts)
	return err
}

func (j *Json) Bytes() []byte {
	return j.buf.Bytes()
}

func (j *Json) String() string {
	return string(j.buf.Bytes())
}

func (j *Json) Decode(v any) error {
	return json.Unmarshal(j.buf.Bytes(), v)
}

func (j *Json) Read(p []byte) (n int, err error) {
	return j.buf.Read(p)
}

func (j *Json) Write(p []byte) (n int, err error) {
	return j.buf.Write(p)
}

func (j *Json) Validate(t any) error {
	return NewValidator[any]().From(j.Bytes()).Bind(t)
}
