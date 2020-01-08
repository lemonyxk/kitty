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
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
)

type crypto int

const Crypto crypto = iota

func (c crypto) Md5(input []byte) string {
	var byte16 = md5.Sum(input)
	var bytes = make([]byte, 16)
	for i := 0; i < 16; i++ {
		bytes[i] = byte16[i]
	}
	return hex.EncodeToString(bytes)
}

func (c crypto) Sha1(input []byte) string {
	var byte20 = sha1.Sum(input)
	var bytes = make([]byte, 20)
	for i := 0; i < 20; i++ {
		bytes[i] = byte20[i]
	}
	return hex.EncodeToString(bytes)
}

func (c crypto) Base64Encode(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

func (c crypto) Base64Decode(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}
