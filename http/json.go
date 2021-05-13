/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-13 19:33
**/

package http

import "github.com/json-iterator/go"

type Json struct {
	any jsoniter.Any
}

func (j *Json) Reset(data interface{}) jsoniter.Any {
	bts, _ := jsoniter.Marshal(data)
	j.any = jsoniter.Get(bts)
	return j.any
}

func (j *Json) getAny() jsoniter.Any {
	if j.any != nil {
		return j.any
	}
	j.any = jsoniter.Get(nil)
	return j.any
}

func (j *Json) Iter() jsoniter.Any {
	return j.getAny()
}

func (j *Json) Has(key string) bool {
	return j.getAny().Get(key).LastError() == nil
}

func (j *Json) Empty(key string) bool {
	return j.getAny().Get(key).ToString() == ""
}

func (j *Json) Get(path ...interface{}) Value {
	var res = j.getAny().Get(path...)
	if res.LastError() != nil {
		return Value{}
	}
	var p = res.ToString()
	return Value{v: &p}
}

func (j *Json) Bytes() []byte {
	return j.Bytes()
}

func (j *Json) String() string {
	return j.getAny().ToString()
}

func (j *Json) Path(path ...interface{}) jsoniter.Any {
	return j.getAny().Get(path...)
}

func (j *Json) Array(path ...interface{}) Array {
	var result []jsoniter.Any
	var val = j.getAny().Get(path...)
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.Get(i))
	}
	return result
}
