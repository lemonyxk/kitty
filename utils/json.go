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

func (j json) JsonEncode(v interface{}) []byte {
	res, err := jsoniter.Marshal(v)
	if err != nil {
		return nil
	}
	return res
}

func (j json) JsonDecode(data []byte, output interface{}) error {
	return jsoniter.Unmarshal(data, output)
}

type Result struct {
	data jsoniter.Any
}

type Array struct {
	data []jsoniter.Any
}

func (j json) JsonPath(data []byte, path ...interface{}) *Result {
	return &Result{data: jsoniter.Get(data, path...)}
}

func (r *Result) Data() jsoniter.Any {
	return r.data
}

func (r *Result) Exists() bool {
	return r.data.LastError() == nil
}

func (r *Result) String() string {
	return r.data.ToString()
}

func (r *Result) Int() int {
	return r.data.ToInt()
}

func (r *Result) Float64() float64 {
	return r.data.ToFloat64()
}

func (r *Result) Bool() bool {
	return r.data.ToBool()
}

func (r *Result) Interface() interface{} {
	return r.data.GetInterface()
}

func (r *Result) Array() *Array {
	var result []jsoniter.Any
	var val = r.data
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.Get(i))
	}
	return &Array{data: result}
}

func (a *Array) String() []string {
	var result []string
	for i := 0; i < len(a.data); i++ {
		result = append(result, a.data[i].ToString())
	}
	return result
}

func (a *Array) Int() []int {
	var result []int
	for i := 0; i < len(a.data); i++ {
		result = append(result, a.data[i].ToInt())
	}
	return result
}

func (a *Array) Float64() []float64 {
	var result []float64
	for i := 0; i < len(a.data); i++ {
		result = append(result, a.data[i].ToFloat64())
	}
	return result
}
