/**
* @program: kitty
*
* @create: 2024-12-19 15:08
**/

package main

import (
	"github.com/lemonyxk/kitty/socket/http"
	"testing"
)

func BenchmarkIsNil3(b *testing.B) {
	var validate = http.NewValidator[*User]()

	user := &User{}

	var bts = []byte(`{
	"FirstName": "Badger",
	"LastName": "Smith",
	"Age": 135,
	"Gender": "111",
	"Email": "B",
	"Addresses": [],
	"Maps": {"a": "b"}
}`)

	for i := 0; i < b.N; i++ {
		validate.From(bts).Bind(user)
	}
}
