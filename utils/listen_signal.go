/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-22 17:44
**/

package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func ListenSignal(fn func(sig os.Signal)) {
	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	// 阻塞
	fn(<-signalChan)
	// 停止
	signal.Stop(signalChan)
}
