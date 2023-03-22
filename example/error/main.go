/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-08-24 04:36
**/

package main

import (
	"fmt"

	"github.com/lemonyxk/kitty/errors"
)

func main() {

	// panic: 1
	//
	// goroutine 1 [running]:
	// main.test3(...)
	// /Users/lemo/lemo-hub/kitty/example/error/main.go:44
	// main.test2()
	// /Users/lemo/lemo-hub/kitty/example/error/main.go:36 +0x27
	// main.test1()
	// /Users/lemo/lemo-hub/kitty/example/error/main.go:28 +0x19
	// main.main()
	// /Users/lemo/lemo-hub/kitty/example/error/main.go:21 +0x1d

	var e = errors.New("hello")

	fmt.Printf("%+v\n", e.Stack())

	var err = test1()

	fmt.Printf("%+v\n", errors.Unwrap(errors.Unwrap(err)))
	fmt.Println(err)
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
	return errors.NewWithStack("test3 error")
}
