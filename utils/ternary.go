/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-17 14:42
**/

package utils

type ternary int

const Ternary ternary = iota

func (ternary ternary) String(b bool, s1, s2 string) string {
	if b {
		return s1
	}
	return s2
}

func (ternary ternary) Int(b bool, s1, s2 int) int {
	if b {
		return s1
	}
	return s2
}

func (ternary ternary) Int64(b bool, s1, s2 int64) int64 {
	if b {
		return s1
	}
	return s2
}

func (ternary ternary) Float32(b bool, s1, s2 float32) float32 {
	if b {
		return s1
	}
	return s2
}

func (ternary ternary) Float64(b bool, s1, s2 float64) float64 {
	if b {
		return s1
	}
	return s2
}

func (ternary ternary) Interface(b interface{}, s1, s2 interface{}) interface{} {
	if b == nil {
		return s1
	} else {
		return s2
	}
}
