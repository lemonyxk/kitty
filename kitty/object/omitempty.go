/**
* @program: engine
*
* @create: 2025-04-05 19:41
**/

package object

import (
	"reflect"
	"strings"

	"github.com/lemonyxk/kitty/errors"
)

func doStruct(dstRv reflect.Value, srcRv reflect.Value, mapKeys map[string]struct{}, remove bool) error {

	if !dstRv.IsValid() || !srcRv.IsValid() {
		return nil
	}

	if dstRv.Kind() == reflect.Ptr {
		if dstRv.IsNil() {
			dstRv = reflect.New(dstRv.Type().Elem()).Elem()
		} else {
			dstRv = dstRv.Elem()
		}
	}

	if dstRv.Kind() != reflect.Struct {
		return errors.New("destination must be a struct")
	}

	if srcRv.Kind() == reflect.Ptr {
		if srcRv.IsNil() {
			return nil
		}
		srcRv = srcRv.Elem()
	}

	var srcType = reflect.TypeOf(srcRv.Interface())

	if srcType.Kind() != reflect.Map {
		return errors.New("source must be a map")
	}

	if srcType.Key().Kind() != reflect.String || srcType.Elem().Kind() != reflect.Interface {
		return errors.New("source map must be string to interface")
	}

	if srcRv.CanConvert(reflect.TypeOf(map[string]any{})) {
		srcRv = srcRv.Convert(reflect.TypeOf(map[string]any{}))
	}

	var src, ok = srcRv.Interface().(map[string]any)
	if !ok {
		return errors.New("source must be a map[string]any")
	}

	for i := 0; i < dstRv.NumField(); i++ {
		var field = dstRv.Type().Field(i)
		var fieldName = field.Name
		var dstField = dstRv.FieldByName(fieldName)

		if field.Anonymous {
			if err := doStruct(dstField, srcRv, mapKeys, false); err != nil {
				return err
			}
			continue
		}

		var tag = field.Tag.Get("json")
		if tag == "" {
			continue
		}

		var jsonName = strings.Split(tag, ",")[0]
		if jsonName == "-" {
			continue
		}

		mapKeys[jsonName] = struct{}{}

		if _, ok := src[jsonName]; !ok {
			continue
		}

		switch dstField.Kind() {
		case reflect.Struct:
			var mapKeys = map[string]struct{}{}
			if err := doStruct(dstField, reflect.ValueOf(src[jsonName]), mapKeys, true); err != nil {
				return err
			}
		case reflect.Ptr:
			if err := doPtr(dstField, reflect.ValueOf(src[jsonName])); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			if err := doSlice(dstField, reflect.ValueOf(src[jsonName])); err != nil {
				return err
			}
		case reflect.Map:
			if err := doMap(dstField, reflect.ValueOf(src[jsonName])); err != nil {
				return err
			}
		default:

		}
	}

	if !remove {
		return nil
	}

	for k := range src {
		if _, ok := mapKeys[k]; ok {
			continue
		}
		delete(src, k)
	}

	return nil
}

func doPtr(dstRv reflect.Value, srcRv reflect.Value) error {
	if !srcRv.IsValid() {
		return nil
	}

	if dstRv.IsNil() {
		dstRv = reflect.New(dstRv.Type().Elem()).Elem()
	} else {
		dstRv = dstRv.Elem()
	}
	switch dstRv.Kind() {
	case reflect.Ptr:
		return doPtr(dstRv, srcRv)
	case reflect.Struct:
		return doStruct(dstRv, srcRv, map[string]struct{}{}, true)
	case reflect.Slice, reflect.Array:
		return doSlice(dstRv, srcRv)
	case reflect.Map:
		return doMap(dstRv, srcRv)
	default:
		return nil
	}
}

func doSlice(dstRv reflect.Value, srcRv reflect.Value) error {
	//if srcRv.Kind() != reflect.Slice && srcRv.Kind() != reflect.Array {
	//	return errors.New("source must be a slice or array")
	//}

	if !srcRv.IsValid() {
		return nil
	}

	if srcRv.Len() == 0 {
		return nil
	}

	var dstType = dstRv.Type().Elem()

	switch dstType.Kind() {
	case reflect.Struct:
		return doStruct(reflect.New(dstType), srcRv.Index(0), map[string]struct{}{}, true)
	case reflect.Ptr:
		return doPtr(reflect.New(dstType.Elem()), srcRv.Index(0))
	case reflect.Slice, reflect.Array:
		return doSlice(reflect.New(dstType.Elem()), srcRv.Index(0))
	case reflect.Map:
		return doMap(reflect.New(dstType.Elem()), srcRv.Index(0))
	default:
		return nil
	}
}

func doMap(dstRv reflect.Value, srcRv reflect.Value) error {
	if !srcRv.IsValid() {
		return nil
	}

	if srcRv.Len() == 0 {
		return nil
	}

	var dstType = dstRv.Type().Elem()

	switch dstType.Kind() {
	case reflect.Struct:
		return doStruct(reflect.New(dstType), srcRv.MapIndex(srcRv.MapKeys()[0]), map[string]struct{}{}, true)
	case reflect.Ptr:
		return doPtr(reflect.New(dstType.Elem()), srcRv.MapIndex(srcRv.MapKeys()[0]))
	case reflect.Slice, reflect.Array:
		return doSlice(reflect.New(dstType.Elem()), srcRv.MapIndex(srcRv.MapKeys()[0]))
	case reflect.Map:
		return doMap(reflect.New(dstType.Elem()), srcRv.MapIndex(srcRv.MapKeys()[0]))
	default:
		return nil
	}
}

func Omitempty[T any](src any) error {
	var srcRv = reflect.ValueOf(src)
	var dstRv = reflect.ValueOf(new(T))

	switch dstRv.Kind() {
	case reflect.Struct:
		return doStruct(dstRv, srcRv, map[string]struct{}{}, true)
	case reflect.Ptr:
		return doPtr(dstRv, srcRv)
	case reflect.Slice, reflect.Array:
		return doSlice(srcRv, srcRv)
	case reflect.Map:
		return doMap(dstRv, srcRv)
	default:
		return nil
	}
}
