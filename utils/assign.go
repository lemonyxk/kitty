/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-04-27 03:33
**/

package utils

import (
	"errors"
	"fmt"
	"reflect"
)

type destination struct {
	dst       interface{}
	src       interface{}
	tag       string
	allowZero bool
	allowTag  bool
}

type field struct {
	data *destination
	name string
}

type source struct {
	data *destination
}

func Struct(dst interface{}) *destination {
	return &destination{
		dst:       dst,
		src:       nil,
		tag:       "json",
		allowZero: false,
		allowTag:  false,
	}
}

func (d *destination) Src(src interface{}) *source {
	d.src = src
	return &source{data: d}
}

func (s *source) Do() error {
	var d = s.data
	return doAssign(d.dst, d.src, d.tag, d.allowZero, d.allowTag)
}

func (s *source) AllowZero() *source {
	s.data.allowZero = true
	return s
}

func (s *source) AllowTag() *source {
	s.data.allowTag = true
	return s
}

func (s *source) SetTag(tag string) *source {
	s.data.tag = tag
	return s
}

func (d *destination) Field(name string) *field {
	return &field{data: d, name: name}
}

func (f *field) Set(v interface{}) error {
	var d = f.data
	d.src = map[string]interface{}{f.name: v}
	d.allowZero = true
	return doAssign(d.dst, d.src, d.tag, d.allowZero, d.allowTag)
}

func (f *field) AllowTag() *field {
	f.data.allowTag = true
	return f
}

func (f *field) SetTag(tag string) *field {
	f.data.tag = tag
	return f
}

func doAssign(dst, src interface{}, tag string, allowZero, allowTag bool) error {

	var dstValue = reflect.ValueOf(dst)
	var srcValue = reflect.ValueOf(src)
	var dstType = reflect.TypeOf(dst)
	var srcType = reflect.TypeOf(src)

	if dstValue.IsNil() {
		return errors.New("dst is nil")
	}

	var dstValueElem = dstValue.Elem()
	var srcValueElem = srcValue
	var dstTypeElem = dstType.Elem()
	var srcTypeElem = srcType

	if dstType.Kind() != reflect.Ptr {
		return errors.New("kind of dst is not ptr")
	}

	if dstTypeElem.Kind() != reflect.Struct {
		return errors.New("kind of dst is not struct")
	}

	if srcType.Kind() == reflect.Ptr {
		// nothing to copy
		if srcValue.IsNil() {
			return nil
		}
		srcValueElem = srcValue.Elem()
		srcTypeElem = srcType.Elem()
	}

	// if !allowDiffType {
	// 	if dstValueElem.Type() != srcValueElem.Type() {
	// 		return errors.New("dst and src has different kind")
	// 	}
	// }

	switch srcTypeElem.Kind() {
	case reflect.Struct:
		for i := 0; i < srcTypeElem.NumField(); i++ {

			if !allowZero {
				if srcValueElem.Field(i).IsZero() {
					continue
				}
			}

			var name = srcTypeElem.Field(i).Name
			var t = srcValueElem.Field(i).Kind()

			var s, ok = dstTypeElem.FieldByName(name)
			if !ok || s.Type.Kind() != t {
				continue
			}

			dstValueElem.FieldByName(name).Set(srcValueElem.Field(i))
		}
	case reflect.Map:
		var keys = srcValueElem.MapKeys()
		for i := 0; i < len(keys); i++ {

			if keys[i].Kind() != reflect.String {
				continue
			}

			var name = keys[i]
			var t = srcValueElem.MapIndex(name)

			if !allowZero {
				if t.IsZero() {
					continue
				}
			}

			srcKey := name.String()

			var s, ok = dstTypeElem.FieldByName(srcKey)
			if !ok {
				if !allowTag {
					if len(keys) == 1 {
						return fmt.Errorf("not found field %s", srcKey)
					}
					continue
				}
				k, ok := hasTag(dstTypeElem, tag, srcKey)
				if !ok {
					if len(keys) == 1 {
						return fmt.Errorf("not found %s in tag %s", srcKey, tag)
					}
					continue
				}

				srcKey = k.Name
				s = k

			}

			var it = reflect.TypeOf(t.Interface())

			if s.Type.Kind() != it.Kind() {
				if len(keys) == 1 {
					return fmt.Errorf("field %s type is %s get %s", s.Name, s.Type.String(), it.Kind().String())
				}
				continue
			}

			v := dstValueElem.FieldByName(srcKey)

			switch t.Interface().(type) {
			case int:
				v.SetInt(int64(t.Interface().(int)))
			case uint64:
				v.SetUint(t.Interface().(uint64))
			case float64:
				v.SetFloat(t.Interface().(float64))
			case bool:
				v.SetBool(t.Interface().(bool))
			case []byte:
				v.SetBytes(t.Interface().([]byte))
			case string:
				v.SetString(t.Interface().(string))
			}

		}
	default:
		return errors.New("kind of src is not struct or map")
	}

	return nil
}

func hasTag(s reflect.Type, t, k string) (reflect.StructField, bool) {
	var n = s.NumField()
	for i := 0; i < n; i++ {
		if s.Field(i).Tag.Get(t) == k {
			return s.Field(i), true
		}
	}
	return reflect.StructField{}, false
}
