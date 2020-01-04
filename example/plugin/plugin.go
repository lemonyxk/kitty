/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-04 19:47
**/

package main

import (
	"github.com/Lemo-yxk/lemo/console"
	"github.com/Lemo-yxk/lemo/exception"
)

func Add(v ...interface{}) (interface{}, exception.ErrorFunc) {
	console.Log("hello plugin", v)
	return nil, nil
}
