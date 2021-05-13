/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-13 19:31
**/

package http

import (
	"bytes"
	"reflect"
	"strconv"
)

type Store struct {
	keys   []string
	values [][]string
}

func (s *Store) Struct(input interface{}) {

	if input == nil {
		panic("input can not be nil")
	}

	var kf = reflect.TypeOf(input)

	var kv = reflect.ValueOf(input)

	if kf.Kind() != reflect.Ptr {
		panic("input must be a pointer")
	}

	if kf.Elem().Kind() != reflect.Struct {
		panic("input must be a struct")
	}

	if !kv.IsValid() || kv.IsNil() {
		panic("input is invalid or nil")
	}

	var findIndex = func(k1 string) int {
		for i := 0; i < kf.Elem().NumField(); i++ {
			var k2 = kf.Elem().Field(i).Tag.Get("json")
			if k1 == k2 {
				return i
			}
		}
		return -1
	}

	for i := 0; i < len(s.keys); i++ {
		var k = s.keys[i]
		var v = s.values[i]

		var index = findIndex(k)
		if index == -1 {
			continue
		}

		var vv = kv.Elem().Field(index)

		// v at least more than 1 item
		switch vv.Interface().(type) {
		case bool:
			vv.SetBool(v[0] == "TRUE")
		case int, int8, int16, int32, int64:
			r, _ := strconv.ParseInt(v[0], 10, 64)
			vv.SetInt(r)
		case uint, uint8, uint16, uint32, uint64:
			r, _ := strconv.ParseUint(v[0], 10, 64)
			vv.SetUint(r)
		case float32, float64:
			r, _ := strconv.ParseFloat(v[0], 64)
			vv.SetFloat(r)
		case string:
			vv.SetString(v[0])
		case []bool:
			var res []bool
			for j := 0; j < len(v); j++ {
				res = append(res, v[j] == "TRUE")
			}
			vv.Set(reflect.ValueOf(res))
		case []int, []int8, []int16, []int32, []int64:
			var res []int64
			for j := 0; j < len(v); j++ {
				r, _ := strconv.ParseInt(v[j], 10, 64)
				res = append(res, r)
			}
			vv.Set(reflect.ValueOf(res))
		case []uint, []uint16, []uint32, []uint64:
			var res []uint64
			for j := 0; j < len(v); j++ {
				r, _ := strconv.ParseUint(v[j], 10, 64)
				res = append(res, r)
			}
			vv.Set(reflect.ValueOf(res))
		case []float32, []float64:
			var res []float64
			for j := 0; j < len(v); j++ {
				r, _ := strconv.ParseFloat(v[j], 64)
				res = append(res, r)
			}
			vv.Set(reflect.ValueOf(res))
		case []string:
			var res []string
			for j := 0; j < len(v); j++ {
				res = append(res, v[j])
			}
			vv.Set(reflect.ValueOf(res))
		case []byte:
			var res = []byte(v[0])
			vv.Set(reflect.ValueOf(res))
		case [][]byte:
			var res [][]byte
			for j := 0; j < len(v); j++ {
				res = append(res, []byte(v[j]))
			}
			vv.Set(reflect.ValueOf(res))
		}
	}
}

func (s *Store) Has(key string) bool {
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			return true
		}
	}
	return false
}

func (s *Store) Empty(key string) bool {
	var v = s.First(key).v
	return v == nil || *v == ""
}

func (s *Store) First(key string) Value {
	var res Value
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			res.v = &s.values[i][0]
			return res
		}
	}
	return res
}

func (s *Store) Index(key string, index int) Value {
	var res Value
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			res.v = &s.values[i][index]
			return res
		}
	}
	return res
}

func (s *Store) All(key string) Values {
	var res []string
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			for j := 0; j < len(s.values[i]); j++ {
				res = append(res, s.values[i][j])
			}
		}
	}
	return res
}

func (s *Store) Add(key string, value []string) {
	s.keys = append(s.keys, key)
	s.values = append(s.values, value)
}

func (s *Store) Remove(key string) {
	var index = -1
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	s.keys = append(s.keys[0:index], s.keys[index+1:]...)
	s.values = append(s.values[0:index], s.values[index+1:]...)
}

func (s *Store) Keys() []string {
	return s.keys
}

func (s *Store) Values() [][]string {
	return s.values
}

func (s *Store) String() string {

	var buff bytes.Buffer

	for i := 0; i < len(s.keys); i++ {
		buff.WriteString(s.keys[i] + ":")
		for j := 0; j < len(s.values[i]); j++ {
			buff.WriteString(s.values[i][j])
			if j != len(s.values[i])-1 {
				buff.WriteString(",")
			}
		}
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
