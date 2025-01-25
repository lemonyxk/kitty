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
	"net/url"
)

type Store struct {
	url.Values
}

func (s *Store) Has(key string) bool {
	for k := range s.Values {
		if k == key {
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
	for k, v := range s.Values {
		if k == key {
			res.v = &v[0]
			return res
		}
	}
	return res
}

func (s *Store) Index(key string, index int) Value {
	var res Value
	for k, v := range s.Values {
		if k == key {
			if index < len(v) {
				res.v = &v[index]
				return res
			}
		}
	}
	return res
}

func (s *Store) All(key string) Values {
	for k, v := range s.Values {
		if k == key {
			return v
		}
	}
	return nil
}

func (s *Store) Add(key string, value []string) {
	s.Values[key] = value
}

func (s *Store) Remove(key string) {
	delete(s.Values, key)
}

func (s *Store) String() string {

	var buff bytes.Buffer

	for k, v := range s.Values {
		buff.WriteString(k + ":")
		for i := 0; i < len(v); i++ {
			buff.WriteString(v[i])
			if i != len(v)-1 {
				buff.WriteString(",")
			}
		}
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var bts = buff.Bytes()

	return string(bts[:len(bts)-1])
}
