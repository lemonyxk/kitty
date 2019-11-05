/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-05 20:45
**/

package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(input []byte) string {
	var byte16 = md5.Sum(input)
	var bytes = make([]byte, 16)
	for i := 0; i < 16; i++ {
		bytes[i] = byte16[i]
	}
	return hex.EncodeToString(bytes)
}
