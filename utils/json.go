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
	"encoding/json"
)

func JsonEncode(v interface{}) []byte {
	res, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return res
}

func JsonDecode(data []byte, output interface{}) {
	_ = json.Unmarshal(data, output)
}
