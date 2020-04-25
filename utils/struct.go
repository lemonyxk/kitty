/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-05 11:42
**/

package utils

import (
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type structure int

const Structure structure = iota

func (d structure) StructToMap(input interface{}) map[string]interface{} {
	var output = make(map[string]interface{})

	if input == nil {
		return output
	}

	var kf = reflect.TypeOf(input)

	var vf = reflect.ValueOf(input)

	if kf.Kind() == reflect.Ptr {
		kf = kf.Elem()
		vf = vf.Elem()
	}

	if kf.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < kf.NumField(); i++ {
		output[kf.Field(i).Tag.Get("json")] = vf.Field(i).Interface()
	}

	return output
}

func (d structure) MapToStruct(input interface{}, output interface{}) error {
	return mapstructure.WeakDecode(input, output)
}

func (d structure) Assign(dst, src interface{}) error {
	return assign(dst, src, true, false)
}

func (d structure) AssignDiff(dst, src interface{}) error {
	return assign(dst, src, true, true)
}

func (d structure) AssignNoZero(dst, src interface{}) error {
	return assign(dst, src, false, false)
}

func (d structure) AssignDiffNoZero(dst, src interface{}) error {
	return assign(dst, src, false, true)
}

func assign(dst, src interface{}, allowZero, allowDiffType bool) error {

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

	if srcTypeElem.Kind() != reflect.Struct {
		return errors.New("kind of src is not struct")
	}

	if !allowDiffType {
		if dstValueElem.Type() != srcValueElem.Type() {
			return errors.New("dst and src has different kind")
		}
	}

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

	return nil
}
