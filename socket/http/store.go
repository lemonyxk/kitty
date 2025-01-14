/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:31
**/

package http

import (
	"bytes"
)

type Store struct {
	keys   []string
	values [][]string
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
		if i != len(s.keys)-1 {
			buff.WriteString(" ")
		}
	}

	if buff.Len() == 0 {
		return ""
	}

	return string(buff.Bytes())
}
