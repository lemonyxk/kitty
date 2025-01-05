/**
* @program: kitty
*
* @create: 2024-12-11 22:37
**/

package main

import (
	"bytes"
	"github.com/lemonyxk/kitty/errors"
	json "github.com/lemonyxk/kitty/json"
	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/socket/http"
	"log"
)

type User struct {
	FirstName      string `validate:"required"`
	LastName       string `validate:"required"`
	Age            uint8  `validate:"gte:0,lte:190"`
	Email          string `validate:"required,email"`
	Gender         string `validate:"default:unknown"`
	FavouriteColor string
	IsAdmin        bool              `validate:"default:true"`
	Addresses      []*Address        `json:"addresses" validate:"required,nonempty"` // a person can have a home and cottage...
	Maps           map[string]string `validate:"required"`                           // a person can have a home and cottage...
}

type User1 struct {
	FirstName      string `validate:"required"`
	LastName       string `validate:"required"`
	Age            uint8  `validate:"gte:0,lte:190"`
	Email          string `validate:"required,email"`
	Gender         string `validate:"default:unknown"`
	FavouriteColor string
	IsAdmin        bool              `validate:"default:true"`
	Addresses      string            `json:"addresses" validate:"required,nonempty"` // a person can have a home and cottage...
	Maps           map[string]string `validate:"required"`                           // a person can have a home and cottage...
}

// Address houses a users address information
type Address struct {
	Street string
	City   string
	Planet string
	Phone  string
}

func main() {
	var validate = http.NewValidator[any]()

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

	if err := validate.From(bts).Bind(user); err != nil {
		log.Println(err)
	} else {
		log.Printf("%+v", user)
	}

	user1 := &User1{}

	var bts1 = []byte(`{
	"FirstName": "Badger",
	"LastName": "Smith",
	"Age": 135,
	"Gender": "1111",
	"Email": "B",
	"Addresses": "1"
}`)

	if err := validate.From(bts1).Bind(user1); err != nil {
		log.Println(err)
	} else {
		log.Printf("%+v", user1)
	}

	var a = kitty.M{"a": 2}

	var buf = new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(a); err != nil {
		log.Println(err.Error())
	}

	log.Println(buf.Bytes())

	var err *errors.Error
	log.Println(kitty.IsNil(err))
}
