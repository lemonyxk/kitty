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
	"log"
)

func Add(v ...interface{}) (interface{}, error) {
	log.Println("hello plugin", v)
	return nil, nil
}
