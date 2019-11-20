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

	var kf = reflect.TypeOf(input)

	if kf.Kind() == reflect.Ptr {
		return nil
	}

	if kf.Kind() != reflect.Struct {
		return nil
	}

	var vf = reflect.ValueOf(input)

	for i := 0; i < kf.NumField(); i++ {
		output[kf.Field(i).Tag.Get("json")] = vf.Field(i).Interface()
	}

	return output
}

func MapToStruct(input interface{}, output interface{}) error {
	return mapstructure.WeakDecode(input, output)
}
