package main

import (
	"github.com/Lemo-yxk/lemo/logger"
	"github.com/Lemo-yxk/lemo/utils"
)

type User struct {
	Name string `json:"name" mapstructure:"name"`
	Addr string `json:"addr"`
	Age  int    `json:"age"`
}

func main() {

	var m = map[string]interface{}{
		"name": "haha",
	}

	var user = &User{}

	utils.MapToStruct(m, user)

	logger.Println(user)

}
