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

func (j json) Decode(data []byte, output interface{}) {
	_ = jsoniter.Unmarshal(data, output)
}

type data struct {
	any jsoniter.Any
}

type result struct {
	any jsoniter.Any
}

func (j json) New(v interface{}) data {
	return data{any: jsoniter.Get(j.Encode(v))}
}

func (j json) Bytes(v []byte) data {
	return data{any: jsoniter.Get(v)}
}

func (j json) String(v string) data {
	return data{any: jsoniter.Get([]byte(v))}
}

func (r data) Any() jsoniter.Any {
	return r.any
}

func (r data) Get(path ...interface{}) result {
	return result{any: r.any.Get(path...)}
}

func (r result) Exists() bool {
	return r.any.LastError() == nil
}

func (r result) String() string {
	return r.any.ToString()
}

func (r result) Bytes() []byte {
	return []byte(r.any.ToString())
}

func (r result) Size() int {
	return r.any.Size()
}

func (r result) Int() int {
	return r.any.ToInt()
}

func (r result) Float64() float64 {
	return r.any.ToFloat64()
}

func (r result) Bool() bool {
	return r.any.ToBool()
}

func (r result) Interface() interface{} {
	return r.any.GetInterface()
}

func (r result) Array() array {
	var result []jsoniter.Any
	var val = r.any
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.Get(i))
	}
	return result
}

type array []jsoniter.Any

func (a array) String() []string {
	var result []string
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToString())
	}
	return result
}

func (a array) Int() []int {
	var result []int
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToInt())
	}
	return result
}

func (a array) Float64() []float64 {
	var result []float64
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToFloat64())
	}
	return result
}

func (a array) Get(path ...interface{}) array {
	var result []jsoniter.Any
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].Get(path...))
	}
	return result
}
