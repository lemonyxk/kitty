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
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func StructToMap(input interface{}) map[string]interface{} {
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

func MapToStruct(input interface{}, output interface{}) error {
	return mapstructure.WeakDecode(input, output)
}
