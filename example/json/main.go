/**
* @program: kitty
*
* @create: 2024-12-11 22:37
**/

package main

import (
	"github.com/lemonyxk/kitty/json"
	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/kitty/object"
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

type Profile struct {
	// 自我介绍
	Bio string `json:"bio,omitempty" bson:"bio,omitempty"`
	// 详细地址
	Address string `json:"address,omitempty" bson:"address,omitempty"`
	// 性别
	Gender int `json:"gender,omitempty" bson:"gender,omitempty"`
	// 生日
	Birthday int64 `json:"birthday,omitempty" bson:"birthday,omitempty"`
}

type Request struct {
	ID      string   `json:"id" validate:"required"`
	Profile *Profile `json:"profile" validate:"required"`
}

type LogEntry struct {
	Level  string          `json:"level"`
	Time   string          `json:"time"`
	Params json.RawMessage `json:"params"` // 关键：使用 json.RawMessage
	// 其他字段...
}

type User111 struct {
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	Age            uint8  `json:"age" validate:"gte:0,lte:190"`
	Email          string `json:"email"`
	Gender         string
	FavouriteColor string
	IsAdmin        bool                `validate:"default:true"`
	Addresses      []*Address          `json:"addresses" validate:"required,nonempty"` // a person can have a home and cottage...
	Maps           map[string]LogEntry `validate:"required" json:"maps"`               // a person can have a home and cottage...
	LogEntry
}

func main() {
	//rawJSON := `{
	//	"level": "INF",
	//	"time": "2025-04-22 15:26:07",
	//	"params": ` + "{\"id\":\"2400000002\",\"profile\":{\"bio\":\"2021年主打产品\",\"gender\":1}}" + `
	//}`
	//
	//// 解析 JSON 到结构体
	//var entry LogEntry
	//err := json.Unmarshal([]byte(rawJSON), &entry)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// 重新 Marshal（params 保持原样）
	//outputJSON, err := json.Marshal(entry)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println(string(outputJSON))

	//	user := &User{}
	//
	//	var bts = []byte(`{
	//	"FirstName": "Badger",
	//	"LastName": "Smith",
	//	"Age": 135,
	//	"Gender": "111",
	//	"Email": "B",
	//	"Addresses": [],
	//	"Maps": {"a": "b"}
	//}`)
	//
	//	if err := validate.From(bts).Bind(user); err != nil {
	//		log.Println(err)
	//	} else {
	//		log.Printf("%+v", user)
	//	}
	//
	//	user1 := &User1{}
	//
	//	var bts1 = []byte(`{
	//	"FirstName": "Badger",
	//	"LastName": "Smith",
	//	"Age": 135,
	//	"Gender": "1111",
	//	"Email": "B",
	//	"Addresses": "1"
	//}`)
	//
	//	if err := validate.From(bts1).Bind(user1); err != nil {
	//		log.Println(err)
	//	} else {
	//		log.Printf("%+v", user1)
	//	}
	//
	//	var a = kitty.M{"a": 2}
	//
	//	var buf = new(bytes.Buffer)
	//
	//	if err := json.NewEncoder(buf).Encode(a); err != nil {
	//		log.Println(err.Error())
	//	}
	//
	//	log.Println(buf.Bytes())
	//
	//	var err *errors.Error
	//	log.Println(kitty.IsNil(err))

	//(map[string]interface{}{
	//	"1": map[string]interface{}{
	//		"2": []int{1, 3},
	//		"belongs": map[string]interface{}{
	//			"jjj": map[string]interface{}{
	//				"kk": 11,
	//				"pp": nil,
	//				"pro.p111": &Profile{
	//					Bio:      "1",
	//					Address:  "1",
	//					Gender:   2,
	//					Birthday: 2,
	//				},
	//				"hh": []any{1, []int{1}},
	//			},
	//		},
	//	},
	//}, "")

	//var obj = kitty.M{
	//	"first_name": "John",
	//	"last_name":  "Doe",
	//	"age":        30,
	//	"email":      "1",
	//	"addresses": []kitty.M{
	//		{
	//			"street": "123 Main St",
	//			"city":   "Anytown",
	//			"planet": "Earth",
	//			"phone":  "555-1234",
	//		},
	//	},
	//	"maps": kitty.M{
	//		"log1": kitty.M{
	//			"level": "INFO",
	//		},
	//	},
	//}

	var str = `{"level":{"level1":"a"},"last_name":"Doe","age":30,"email":"1","addresses":[{"street":"123 Main St","city":"Anytown","planet":"Earth","phone":"555-1234"}],"maps":{"log1":{"level":"INFO"}}}`

	var obj kitty.M
	if err := json.Unmarshal([]byte(str), &obj); err != nil {
		panic(err)
	}

	if err := object.Omitempty[User111](obj); err != nil {
		panic(err)
	}

	var bts, err = json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	println(string(bts))

	var o = object.DotNotation(obj)

	bts, err = json.Marshal(o)
	if err != nil {
		panic(err)
	}

	println(string(bts))
}
