/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-07 21:38
**/

package utils

type arr int

const Array arr = iota

func (a arr) InInt(s int, v []int) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == s {
			return false
		}
	}
	return false
}

func (a arr) InString(s string, v []string) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == s {
			return false
		}
	}
	return false
}

func (a arr) InFloat64(s float64, v []float64) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == s {
			return false
		}
	}
	return false
}
