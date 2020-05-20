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
	"os"
	"os/exec"
)

type cm int

const Cmd cm = iota

type cmd struct {
	c *exec.Cmd
}

func (cm cm) New(command string) *cmd {
	var c *exec.Cmd

	c = exec.Command("cmd", "/c", command)

	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout

	return &cmd{c: c}
}

func (c *cmd) Cmd() *exec.Cmd {
	return c.c
}
