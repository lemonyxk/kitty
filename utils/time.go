/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-01 20:53
**/

package utils

import (
	"time"
)

type ti int

const Time ti = iota

const format = "2006-01-02 15:04:05"

func (_ ti) FormatNow() string {
	return time.Now().Format(format)
}

func (_ ti) FormatTime(t time.Time) string {
	return t.Format(format)
}

func (t ti) FormatTimestamp(timestamp int) string {
	return t.ParseTimestamp(timestamp).Format(format)
}

func (_ ti) ParseTimeString(timestamp string) (time.Time, error) {
	return time.Parse(format, timestamp)
}

func (_ ti) ParseTimestamp(timestamp int) time.Time {
	return time.Unix(int64(timestamp), 0)
}
