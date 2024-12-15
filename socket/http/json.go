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
	"github.com/lemonyxk/kitty/errors"
	json "github.com/lemonyxk/kitty/json"
)

type Json struct {
	buf *bytes.Buffer
}

func (j *Json) Reset(data any) error {
	if j.buf == nil {
		return errors.New("header is not application/json")
	}
	j.buf.Reset()
	var bts, err = json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = j.buf.Write(bts)
	return err
}

func (j *Json) Bytes() []byte {
	if j.buf == nil {
		return nil
	}
	return j.buf.Bytes()
}

func (j *Json) String() string {
	if j.buf == nil {
		return ""
	}
	return string(j.buf.Bytes())
}

func (j *Json) Decode(v any) error {
	if j.buf == nil {
		return errors.New("header is not application/json")
	}
	return json.Unmarshal(j.buf.Bytes(), v)
}

func (j *Json) Read(p []byte) (n int, err error) {
	if j.buf == nil {
		return 0, errors.New("header is not application/json")
	}
	return j.buf.Read(p)
}

func (j *Json) Write(p []byte) (n int, err error) {
	if j.buf == nil {
		return 0, errors.New("header is not application/json")
	}
	return j.buf.Write(p)
}

func (j *Json) Validate(t any) error {
	if j.buf == nil {
		return errors.New("header is not application/json")
	}
	return NewValidator[any]().From(j.Bytes()).Bind(t)
}
