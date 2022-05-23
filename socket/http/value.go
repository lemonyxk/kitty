/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:30
**/

package http

import (
	"strconv"
	"strings"
)

type Values []string

func (v Values) Int() []int {
	var res []int

	if len(v) == 0 {
		return res
	}

	for i := 0; i < len(v); i++ {
		r, _ := strconv.Atoi(v[i])
		res = append(res, r)
	}

	return res
}

func (v Values) Float64() []float64 {
	var res []float64

	if len(v) == 0 {
		return res
	}

	for i := 0; i < len(v); i++ {
		r, _ := strconv.ParseFloat(v[i], 64)
		res = append(res, r)
	}

	return res
}

func (v Values) String() []string {
	var res []string

	if len(v) == 0 {
		return res
	}

	return v
}

func (v Values) Bool() []bool {

	var res []bool

	if len(v) == 0 {
		return res
	}

	for i := 0; i < len(v); i++ {
		res = append(res, strings.ToUpper(v[i]) == "TRUE")
	}

	return res
}

func (v Values) Bytes() [][]byte {
	var res [][]byte

	if len(v) == 0 {
		return res
	}

	for i := 0; i < len(v); i++ {
		res = append(res, []byte(v[i]))
	}

	return res
}

type Value struct {
	v *string
}

func (v Value) Int() int {
	if v.v == nil {
		return 0
	}
	r, _ := strconv.Atoi(*v.v)
	return r
}

func (v Value) Float64() float64 {
	if v.v == nil {
		return 0
	}
	r, _ := strconv.ParseFloat(*v.v, 64)
	return r
}

func (v Value) String() string {
	if v.v == nil {
		return ""
	}
	return *v.v
}

func (v Value) Bool() bool {
	return strings.ToUpper(v.String()) == "TRUE"
}

func (v Value) Bytes() []byte {
	return []byte(v.String())
}
