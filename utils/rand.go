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
	ra "math/rand"
	"strconv"
	"time"
)

type rd int

const Rand rd = iota

// [begin,end)
func (r rd) RandomIntn(start int, end int) int {
	if start == end {
		return start
	}
	if start > end {
		panic("start can not bigger than end")
	}
	ra.Seed(time.Now().UnixNano())
	return start + ra.Intn(end-start)
}

// [begin,end)
func (r rd) RandomFloat64n(start float64, end float64) float64 {
	if start == end {
		return start
	}
	if start > end {
		panic("start can not bigger than end")
	}
	ra.Seed(time.Now().UnixNano())
	return start + (end-start)*ra.Float64()
}

func (r rd) UUID() string {
	var bytes = make([]byte, 16)
	var _, err = io.ReadFull(rand.Reader, bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func (r rd) OrderID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(r.RandomIntn(10000, 100000))
}
