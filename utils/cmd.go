/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-02 19:05
**/

package utils

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
)

type cm int

const Cmd cm = iota

type cmd struct {
	c *exec.Cmd
}

func (cm cm) New(command string) *cmd {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/C", command)
	default:
		c = exec.Command("bash", "-c", command)
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout

	return &cmd{c: c}
}

func (c *cmd) Signal(sig syscall.Signal) error {
	return c.c.Process.Signal(sig)
}

func (c *cmd) Cmd() *exec.Cmd {
	return c.c
}

func (c *cmd) Kill() error {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = killWindows(c.c)
	default:
		err = killUnix(c.c, syscall.SIGKILL)
	}
	if err != nil {
		return err
	}

	_, err = c.c.Process.Wait()
	if err != nil {
		return err
	}

	return nil
}

func killUnix(cmd *exec.Cmd, sig syscall.Signal) error {
	return syscall.Kill(-cmd.Process.Pid, sig)
}

func killWindows(cmd *exec.Cmd) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}
