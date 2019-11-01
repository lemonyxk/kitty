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

const format = "2006-01-02 15:04:05"

func FormatNow() string {
	return time.Now().Format(format)
}

func FormatTime(t time.Time) string {
	return t.Format(format)
}

func FormatTimestamp(timestamp int) string {
	return ParseTimestamp(timestamp).Format(format)
}

func ParseTimeString(timestamp string) (time.Time, error) {
	return time.Parse(format, timestamp)
}

func ParseTimestamp(timestamp int) time.Time {
	return time.Unix(int64(timestamp), 0)
}
