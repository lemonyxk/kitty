/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-16 21:24
**/

package utils

import (
	"strconv"
	"unsafe"
)

type conv int

const Conv conv = iota

func (c conv) Itoa(i int) string {
	return strconv.Itoa(i)
}

func (c conv) Atoi(i string) int {
	var n, _ = strconv.Atoi(i)
	return n
}

func (c conv) StringToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func (c conv) BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (c conv) Float64ToString(i float64) string {
	return strconv.FormatFloat(i, 'f', -1, 64)
}

func (c conv) Float32ToString(i float64) string {
	return strconv.FormatFloat(i, 'f', -1, 32)
}

func (c conv) BoolToInt(i bool) int {
	if i {
		return 1
	}
	return 0
}

func (c conv) IntToBool(i int) bool {
	if i > 0 {
		return true
	}
	return false
}

func (c conv) StringToFloat64(i string) float64 {
	var n, _ = strconv.ParseFloat(i, 64)
	return n
}

func (c conv) StringToFloat32(i string) float64 {
	var n, _ = strconv.ParseFloat(i, 32)
	return n
}
