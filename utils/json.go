/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-04 18:53
**/

package utils

import (
	"github.com/json-iterator/go"
)

func JsonEncode(v interface{}) []byte {
	res, err := jsoniter.Marshal(v)
	if err != nil {
		return nil
	}
	return res
}

func JsonDecode(data []byte, output interface{}) error {
	return jsoniter.Unmarshal(data, output)
}
