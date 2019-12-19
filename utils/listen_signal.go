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

type sig int

const Signal sig = iota

type done struct {
	Done func(func(sig os.Signal))
}

// listen all signal
func (s sig) ListenAll() *done {
	var signalList = []os.Signal{
		syscall.SIGABRT, syscall.SIGALRM, syscall.SIGBUS, syscall.SIGCHLD, syscall.SIGCONT,
		syscall.SIGEMT, syscall.SIGFPE, syscall.SIGHUP, syscall.SIGILL, syscall.SIGINFO,
		syscall.SIGINT, syscall.SIGIO, syscall.SIGIOT, syscall.SIGKILL, syscall.SIGPIPE,
		syscall.SIGPROF, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGSYS,
		syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGTSTP, syscall.SIGTTIN, syscall.SIGTTOU,
		syscall.SIGURG, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGVTALRM, syscall.SIGWINCH, syscall.SIGXCPU, syscall.SIGXFSZ,
	}
	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, signalList...)
	return &done{Done: func(f func(signal os.Signal)) {
		// 阻塞
		f(<-signalChan)
		// 停止
		signal.Stop(signalChan)
	}}
}

func (s sig) Send(signal syscall.Signal) {
	_ = syscall.Kill(syscall.Getpid(), signal)
}

func (s sig) ListenKill() *done {
	var signalList = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2}
	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, signalList...)
	return &done{Done: func(f func(signal os.Signal)) {
		// 阻塞
		f(<-signalChan)
		// 停止
		signal.Stop(signalChan)
	}}
}

func (s sig) Listen(sig ...os.Signal) *done {
	var signalList = sig
	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, signalList...)
	return &done{Done: func(f func(signal os.Signal)) {
		// 阻塞
		f(<-signalChan)
		// 停止
		signal.Stop(signalChan)
	}}
}
