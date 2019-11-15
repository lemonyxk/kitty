package main

import (
	jsoniter "github.com/json-iterator/go"
	"log"
)

type People struct {
	value int
}

func (p *People) Value() int {
	return p.value
}

func main() {

	val := []byte(`{"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}`)
	log.Println(jsoniter.Get(val, "Colors", "*").ToString())

}
