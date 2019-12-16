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
	ra "math/rand"
	"strconv"
	"time"
)

type rd int

const Rand rd = iota

// [begin,end)
func (r rd) RandomIntn(start int, end int) int {
	if start > end {
		panic("start can not bigger than end")
	}
	var number = new(big.Int).SetInt64(int64(end - start))
	var randomNumber, _ = rand.Int(rand.Reader, number)
	return int(randomNumber.Int64()) + start
}

// [begin,end)
func (r rd) RandomFloat64n(start float64, end float64) float64 {
	ra.Seed(time.Now().UnixNano())
	return start + end*ra.Float64()
}

func (r rd) UUID() string {
	var bytes = make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func (r rd) OrderID() string {
	var t = time.Now()
	return strconv.FormatInt(t.UnixNano(), 10) + strconv.Itoa(r.RandomIntn(10000, 100000))
}
