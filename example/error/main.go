/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-08-24 04:36
**/

package main

import (
	errors2 "errors"
	"fmt"
	"log"

	"github.com/lemonyxk/kitty/errors"
	"github.com/rs/xid"
)

func main() {

	var e = errors2.New("new error")
	var w = errors.Wrap(e, "wrap error")

	fmt.Printf("%+v\n", w)

	var err = test1()

	fmt.Printf("%+v\n", errors.Unwrap(errors.Unwrap(err)))
	fmt.Println(err)

	var a = []byte("/a" + xid.New().String())

	log.Println(string(a[:len(a)-20]))
}

func test1() error {
	var err = test2()
	if err != nil {
		return errors.Wrap(err, "test1 error")
	}
	return nil
}

func test2() error {
	var err = test3()
	if err != nil {
		return errors.Wrap(err, "test2 error")
	}
	return nil
}

func test3() error {
	return errors.New("test3 error")
}
