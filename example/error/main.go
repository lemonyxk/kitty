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

	"github.com/lemonyxk/kitty/v2/errors"
)

func main() {

	var err = test1()

	fmt.Printf("%+v\n", err)
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
	return fmt.Errorf("test3 error")
}
