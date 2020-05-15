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

const year = 1
const month = 2
const day = 3
const hour = 4
const minute = 5
const second = 6

type Date struct {
	time time.Time
}

type Ticker struct {
	duration time.Duration
	fn       func()
	ticker   *time.Ticker
}

type T struct {
	flag int
	time time.Time
}

func (d Date) Format(format string) string {
	return d.time.Format(format)
}

func (d Date) Time() time.Time {
	return d.time
}

func (d Date) Second() T {
	return T{time: d.time, flag: second}
}

func (d Date) Minute() T {
	return T{time: d.time, flag: minute}
}

func (d Date) Hour() T {
	return T{time: d.time, flag: hour}
}

func (d Date) Day() T {
	return T{time: d.time, flag: day}
}

func (d Date) Month() T {
	return T{time: d.time, flag: month}
}

func (d Date) Year() T {
	return T{time: d.time, flag: year}
}

func (t T) Get() int {
	switch t.flag {
	case second:
		return t.time.Second()
	case minute:
		return t.time.Minute()
	case hour:
		return t.time.Hour()
	case day:
		return t.time.Day()
	case month:
		return int(t.time.Month())
	case year:
		return t.time.Year()
	default:
		return 0
	}
}

func (t T) Begin() int64 {
	switch t.flag {
	case second:
		return t.time.Unix()
	case minute:
		return t.time.Unix() - int64(t.time.Second())
	case hour:
		return t.time.Unix() - int64(t.time.Second()+t.time.Minute()*60)
	case day:
		return t.time.Unix() - int64(t.time.Second()+t.time.Minute()*60+t.time.Hour()*60*60)
	case month:
		return time.Date(t.time.Year(), t.time.Month(), 1, 0, 0, 0, 0, t.time.Location()).Unix()
	case year:
		return time.Date(t.time.Year(), 1, 1, 0, 0, 0, 0, t.time.Location()).Unix()
	default:
		return 0
	}
}

func (t T) End() int64 {
	switch t.flag {
	case second:
		return t.time.Unix()
	case minute:
		return t.Begin() + 60
	case hour:
		return t.Begin() + 3600
	case day:
		return t.Begin() + 86400
	case month:
		return time.Date(t.time.Year(), t.time.Month()+1, 1, 0, 0, 0, 0, t.time.Location()).Unix()
	case year:
		return time.Date(t.time.Year()+1, 1, 1, 0, 0, 0, 0, t.time.Location()).Unix()
	default:
		return 0
	}
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

func (ti ti) String(dateString string) Date {
	var t, _ = time.Parse(FULL, dateString)
	return Date{time: t}
}

func (ti ti) FormatString(format string, dateString string) Date {
	var t, _ = time.Parse(format, dateString)
	return Date{time: t}
}

func (ti ti) Ticker(duration time.Duration, fn func()) *Ticker {
	return &Ticker{fn: fn, duration: duration}
}

func (ticker *Ticker) Start() {
	ticker.ticker = time.NewTicker(ticker.duration)
	go func() {
		for {
			select {
			case <-ticker.ticker.C:
				ticker.fn()
			}
		}
	}()
}

func (ticker *Ticker) Stop() {
	ticker.ticker.Stop()
}
