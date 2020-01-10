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
			return true
		}
	}
	return false
}

func (a arr) InString(s string, v []string) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == s {
			return true
		}
	}
	return false
}

func (a arr) InFloat64(s float64, v []float64) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == s {
			return true
		}
	}
	return false
}

func (a arr) UniqueInt(input []int) []int {
	var temp = make(map[int]struct{})
	for i := 0; i < len(input); i++ {
		temp[input[i]] = struct{}{}
	}
	var res []int
	for key := range temp {
		res = append(res, key)
	}
	return res
}

func (a arr) UniqueString(input []string) []string {
	var temp = make(map[string]struct{})
	for i := 0; i < len(input); i++ {
		temp[input[i]] = struct{}{}
	}
	var res []string
	for key := range temp {
		res = append(res, key)
	}
	return res
}

func (a arr) UniqueFloat64(input []float64) []float64 {
	var temp = make(map[float64]struct{})
	for i := 0; i < len(input); i++ {
		temp[input[i]] = struct{}{}
	}
	var res []float64
	for key := range temp {
		res = append(res, key)
	}
	return res
}
