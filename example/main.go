package main

import "log"

func main() {

	test()
}

func test(a ...int) {
	var b = make([]int, 0)
	log.Println(b == nil)
	b = append(b, a...)
	log.Println(b == nil)
}
