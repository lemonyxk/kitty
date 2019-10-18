/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-05 14:18
**/

package lemo

import (
	"os"
	"os/signal"
	"syscall"
)

func ListenSignal(fn func(sig os.Signal)) {
	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞
	fn(<-signalChan)
}

func ParseMessage(bts []byte) (string, []byte) {

	var s, e int

	var l = len(bts)

	// 正序
	if bts[8] == 58 {

		s = 8

		for i, b := range bts {
			if b == 44 {
				e = i
				break
			}
		}

		if e == 0 {
			return string(bts[s+2 : l-2]), nil
		}

		return string(bts[s+2 : e-1]), bts[e+9 : l-2]

	} else {

		for i := l - 1; i >= 0; i-- {

			if bts[i] == 58 {
				s = i
			}

			if bts[i] == 44 {
				e = i
				break
			}
		}

		if s == 0 {
			return "", nil
		}

		if e == 0 {
			return string(bts[s+2 : l-2]), nil
		}

		return string(bts[s+2 : l-2]), bts[9 : e-1]
	}
}
