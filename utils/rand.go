/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-04 16:15
**/

package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"math/big"
	"strconv"
	"time"
)

// [begin,end]
func RandomInt(start int, end int) int {
	if start > end {
		panic("start can not bigger than end")
	}
	var number = new(big.Int).SetInt64(int64(end - start))
	var randomNumber, _ = rand.Int(rand.Reader, number)
	return int(randomNumber.Int64()) + start
}

func UUID() string {
	var bytes = make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func OrderID() string {
	var t = time.Now()
	return strconv.FormatInt(t.UnixNano(), 10) + strconv.Itoa(RandomInt(10000, 99999))
}
