/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-06 00:00
**/

package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type ne int

const Net ne = iota

func (n ne) ConvertTCPListener(l net.Listener) *net.TCPListener {
	tl, ok := l.(*net.TCPListener)
	if !ok {
		return nil
	}
	return tl
}

func (n ne) SaveFD(l net.Listener, fileName string) error {

	var tl = n.ConvertTCPListener(l)
	if tl == nil {
		return errors.New("type error")
	}

	f, err := tl.File()
	if err != nil {
		return err
	}

	var fd = fmt.Sprintf("%v", f.Fd())
	var name = f.Name()

	return os.Setenv(fileName, fmt.Sprintf("%s %s", fd, name))
}

func (n ne) GetFD(fileName string) (*os.File, error) {

	var arr = strings.Split(os.Getenv(fileName), " ")
	if len(arr) != 2 {
		return nil, errors.New("bad data")
	}

	fdStr, err := strconv.Atoi(arr[0])
	if err != nil {
		return nil, err
	}

	var fd, name = uintptr(fdStr), arr[1]

	var f = os.NewFile(fd, name)

	return f, nil
}
