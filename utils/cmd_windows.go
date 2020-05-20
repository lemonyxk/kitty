// +build windows

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
	"os/exec"
)

type cm int

const Cmd cm = iota

type cmd struct {
	c *exec.Cmd
}

func (cm cm) New(command string) *cmd {
	var c = exec.Command("cmd", "/c", command)
	return &cmd{c: c}
}

func (c *cmd) Cmd() *exec.Cmd {
	return c.c
}
