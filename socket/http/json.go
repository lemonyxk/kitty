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
	"github.com/goccy/go-json"
)

type Json struct {
	buf *bytes.Buffer
}

func (j *Json) Reset(data any) error {
	bts, err := json.Marshal(data)
	if err != nil {
		return err
	}
	j.buf.Reset()
	_, err = j.buf.Write(bts)
	if err != nil {
		return err
	}
	return nil
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
