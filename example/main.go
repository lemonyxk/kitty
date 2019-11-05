package main

import (
	"crypto/md5"
	"fmt"

	"github.com/Lemo-yxk/lemo/logger"
	"github.com/Lemo-yxk/lemo/utils"
)

type User struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
	Age  int    `json:"age"`
}

func main() {

	var user User

	logger.Println(utils.StructToMap(user))

	logger.Println(utils.Md5([]byte("1")))

	fmt.Printf("%x", md5.Sum([]byte("1")))

}
