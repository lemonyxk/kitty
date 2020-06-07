/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-04 18:53
**/

package utils

import (
	"github.com/json-iterator/go"
)

type json int

const Json json = iota

func (j json) Encode(v interface{}) []byte {
	res, err := jsoniter.Marshal(v)
	if err != nil {
		return nil
	}
	return res
}

func (j json) Decode(data []byte, output interface{}) error {
	return jsoniter.Unmarshal(data, output)
}

type Any struct {
	any jsoniter.Any
}

type Result struct {
	any jsoniter.Any
}

func (j json) New(v interface{}) Any {
	return Any{any: jsoniter.Get(j.Encode(v))}
}

func (j json) Bytes(v []byte) Any {
	return Any{any: jsoniter.Get(v)}
}

func (j json) String(v string) Any {
	return Any{any: jsoniter.Get([]byte(v))}
}

func (r Any) Any() jsoniter.Any {
	return r.any
}

func (r Any) Get(path ...interface{}) Result {
	return Result{any: r.any.Get(path...)}
}

func (r Result) Exists() bool {
	return r.any.LastError() == nil
}

func (r Result) String() string {
	return r.any.ToString()
}

func (r Result) Bytes() []byte {
	return []byte(r.any.ToString())
}

func (r Result) Size() int {
	return r.any.Size()
}

func (r Result) Int() int {
	return r.any.ToInt()
}

func (r Result) Float64() float64 {
	return r.any.ToFloat64()
}

func (r Result) Bool() bool {
	return r.any.ToBool()
}

func (r Result) Interface() interface{} {
	return r.any.GetInterface()
}

func (r Result) Array() ResultArray {
	var result []jsoniter.Any
	var val = r.any
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.Get(i))
	}
	return result
}

type ResultArray []jsoniter.Any

func (a ResultArray) String() []string {
	var result []string
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToString())
	}
	return result
}

func (a ResultArray) Int() []int {
	var result []int
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToInt())
	}
	return result
}

func (a ResultArray) Float64() []float64 {
	var result []float64
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToFloat64())
	}
	return result
}

func (a ResultArray) Get(path ...interface{}) ResultArray {
	var result []jsoniter.Any
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].Get(path...))
	}
	return result
}
