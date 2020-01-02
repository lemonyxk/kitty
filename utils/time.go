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

const YMD = "2006-01-02"
const HMS = "15:04:05"
const FULL = "2006-01-02 15:04:05"

type Date struct {
	time time.Time
}

func (d Date) Format(format string) string {
	return d.time.Format(format)
}

func (d Date) Get(format string) time.Time {
	return d.time
}

func (ti ti) New() Date {
	return Date{time: time.Now()}
}

func (ti ti) Time(t time.Time) Date {
	return Date{time: t}
}

func (ti ti) Timestamp(timestamp int64) Date {
	return Date{time: time.Unix(timestamp, 0)}
}

func (ti ti) String(timestamp string) Date {
	var t, _ = time.Parse(FULL, timestamp)
	return Date{time: t}
}
