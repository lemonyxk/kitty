/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-02 19:33
**/

package utils

import (
	"os"
	"strings"
	"sync"
	"syscall"
)

type pro int

const Process pro = iota

type proc struct {
	Cmd *cmd
}

var mux sync.Mutex

var worker []*proc

var workerNumber int

var managerPid int

func (p pro) Fork(fn func(), number int) {

	switch os.Getenv("FORK_CHILD") {
	case "":
		managerPid = os.Getpid()
		_ = os.Setenv("FORK_CHILD", "TRUE")
		workerNumber = number
		run()
	default:
		go fn()
		Signal.ListenKill().Done(func(sig os.Signal) {
			os.Exit(0)
		})
	}
}

func run() {
	mux.Lock()
	for i := 0; i < workerNumber; i++ {
		var c = Cmd.New(strings.Join(os.Args, " "))
		err := c.c.Start()
		if err != nil {
			panic(err)
		}
		go func() { _, _ = c.c.Process.Wait() }()
		worker = append(worker, &proc{Cmd: c})
	}
	mux.Unlock()
}

func (p pro) Kill(pid int) {
	if managerPid == 0 {
		return
	}
	mux.Lock()
	_ = syscall.Kill(pid, syscall.SIGTERM)
	for i := 0; i < len(worker); i++ {
		if worker[i].Cmd.c.Process.Pid == pid {
			worker = append(worker[0:i], worker[i+1:]...)
		}
	}
	mux.Unlock()
}

func (p pro) Reload() {
	if managerPid == 0 {
		return
	}
	for _, proc := range worker {
		p.Kill(proc.Cmd.c.Process.Pid)
	}
	run()
}

func (p pro) Manager() int {
	return managerPid
}

func (p pro) Worker() []*proc {
	return worker
}
