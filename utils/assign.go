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
	"reflect"
)

type assign struct {
	dst       interface{}
	src       interface{}
	allowZero bool
	allowWeak bool
	allowTag  bool
}

func Assign(dst, src interface{}) *assign {
	return &assign{
		dst:       dst,
		src:       src,
		allowZero: false,
		allowWeak: false,
		allowTag:  false,
	}
}

func (a *assign) AllowZero() *assign {
	a.allowZero = true
	return a
}

func (a *assign) AllowWeak() *assign {
	a.allowWeak = true
	return a
}

func (a *assign) AllowTag() *assign {
	a.allowTag = true
	return a
}

func (a *assign) Do() error {
	return doAssign(a.dst, a.src, a.allowZero, a.allowWeak, a.allowTag)
}

func doAssign(dst, src interface{}, allowZero, allowWeak, allowTag bool) error {

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
					continue
				}
				k, ok := hasTag(dstTypeElem, srcKey)
				if !ok {
					continue
				}

				srcKey = k

			}

			if !allowWeak {
				if s.Type.Kind() != t.Kind() {
					continue
				}
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

func hasTag(s reflect.Type, k string) (string, bool) {
	var n = s.NumField()
	for i := 0; i < n; i++ {
		if s.Field(i).Tag.Get("json") == k {
			return s.Field(i).Name, true
		}
	}
	return "", false
}
