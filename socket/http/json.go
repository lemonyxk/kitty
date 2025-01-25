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
	json "github.com/lemonyxk/kitty/json"
)

type Json struct {
	bts []byte
	t   any
}

func (j *Json) Reset(data any) error {
	var bts, err = json.Marshal(data)
	if err != nil {
		return err
	}
	j.bts = bts
	return err
}

func (j *Json) Bytes() []byte {
	return j.bts
}

func (j *Json) String() string {
	return string(j.bts)
}

func (j *Json) Decode(v any) error {
	if len(j.bts) == 0 {
		return nil
	}
	j.t = v
	return json.Unmarshal(j.bts, v)
}

func (j *Json) Validate(t any) error {
	if len(j.bts) == 0 {
		return nil
	}
	j.t = t
	return NewValidator[any]().From(j.bts).Bind(t)
}
