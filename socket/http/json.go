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
	json "github.com/json-iterator/go"
	jsoniter "github.com/json-iterator/go"
)

type Any struct {
	jsoniter.Any
}

func (a *Any) Bytes() []byte {
	return []byte(a.ToString())
}

func (a *Any) String() string {
	return a.ToString()
}

func (a *Any) Float64() float64 {
	return a.ToFloat64()
}

func (a *Any) Int() int {
	return a.ToInt()
}

func (a *Any) Int64() int64 {
	return a.ToInt64()
}

func (a *Any) Decode(v any) error {
	return json.Unmarshal(a.Bytes(), v)
}

type Json struct {
	any jsoniter.Any
	bts []byte
}

func (j *Json) Reset(data any) *Any {
	bts, _ := json.Marshal(data)
	j.any = jsoniter.Get(bts)
	j.bts = bts
	return &Any{Any: j.any}
}

func (j *Json) getAny() *Any {
	if j.any != nil {
		return &Any{Any: j.any}
	}
	j.any = jsoniter.Get(nil)
	return &Any{Any: j.any}
}

func (j *Json) Any() *Any {
	return j.getAny()
}

func (j *Json) Has(key string) bool {
	return j.getAny().Get(key).LastError() == nil
}

func (j *Json) Empty(key string) bool {
	return j.getAny().Get(key).ToString() == ""
}

func (j *Json) Get(path ...any) Value {
	var res = j.getAny().Get(path...)
	if res.LastError() != nil {
		return Value{}
	}
	var p = res.ToString()
	return Value{v: &p}
}

func (j *Json) Bytes() []byte {
	return j.bts
}

func (j *Json) String() string {
	return j.getAny().ToString()
}

func (j *Json) Path(path ...any) *Any {
	return &Any{Any: j.getAny().Get(path...)}
}

func (j *Json) Array(path ...any) Array {
	var result []*Any
	var val = j.getAny().Get(path...)
	for i := 0; i < val.Size(); i++ {
		result = append(result, &Any{Any: val.Get(i)})
	}
	return result
}

func (j *Json) Decode(v any) error {
	return json.Unmarshal(j.bts, v)
}

type Array []*Any

func (a Array) String() []string {
	var result []string
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToString())
	}
	return result
}

func (a Array) Int() []int {
	var result []int
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToInt())
	}
	return result
}

func (a Array) Float64() []float64 {
	var result []float64
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToFloat64())
	}
	return result
}
