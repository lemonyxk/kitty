/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2023-03-01 11:33
**/

package http

import (
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var globalFields = make(map[string]*reflect.StructField)
var globalTags = make(map[string]*Type)
var mux sync.Mutex
var mux1 sync.Mutex

type InvalidError[T any] struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Value    T      `json:"value"`
	Contract string `json:"contract"`
	Op       string `json:"op"`
}

func (i *InvalidError[T]) Message() string {
	var builder strings.Builder
	builder.WriteString("key: ")
	builder.WriteString(i.Key)
	builder.WriteString(", type: ")
	builder.WriteString(i.Type)

	var k = reflect.ValueOf(i.Value)

	builder.WriteString(", value: ")
	switch k.Kind() {
	case reflect.Bool:
		builder.WriteString(strconv.FormatBool(k.Bool()))
	case reflect.String:
		builder.WriteString(k.String())
	case reflect.Int64:
		builder.WriteString(strconv.FormatInt(k.Int(), 10))
	case reflect.Uint64:
		builder.WriteString(strconv.FormatUint(k.Uint(), 10))
	case reflect.Float64:
		builder.WriteString(strconv.FormatFloat(k.Float(), 'f', -1, 64))
	default:
		builder.WriteString("unknown")
	}

	if i.Contract != "" {
		builder.WriteString(", contract: ")
		builder.WriteString(i.Contract)
	}

	builder.WriteString(", op: ")
	builder.WriteString(i.Op)

	return builder.String()
}

func (i *InvalidError[T]) Error() string {
	var builder strings.Builder

	builder.WriteString(i.Key)

	switch i.Op {
	case "required":
		builder.WriteString(" is required")
		builder.WriteString(" but got ")
	case "nonempty":
		builder.WriteString(" must nonempty")
		builder.WriteString(" but got ")
	case "gte":
		builder.WriteString(" >= ")
		builder.WriteString(i.Contract)
		builder.WriteString(" but got ")
	case "lte":
		builder.WriteString(" <= ")
		builder.WriteString(i.Contract)
		builder.WriteString(" but got ")
	case "gt":
		builder.WriteString(" > ")
		builder.WriteString(i.Contract)
		builder.WriteString(" but got ")
	case "lt":
		builder.WriteString(" < ")
		builder.WriteString(i.Contract)
		builder.WriteString(" but got ")
	case "eq":
		builder.WriteString(" = ")
		builder.WriteString(i.Contract)
		builder.WriteString(" but got ")
	}

	var k = reflect.ValueOf(i.Value)

	switch k.Kind() {
	case reflect.Bool:
		builder.WriteString(strconv.FormatBool(k.Bool()))
	case reflect.String:
		v := k.String()
		if v == "" {
			builder.WriteString("''")
		} else {
			builder.WriteString(k.String())
		}
	case reflect.Int64:
		builder.WriteString(strconv.FormatInt(k.Int(), 10))
	case reflect.Uint64:
		builder.WriteString(strconv.FormatUint(k.Uint(), 10))
	case reflect.Float64:
		builder.WriteString(strconv.FormatFloat(k.Float(), 'f', -1, 64))
	default:
		builder.WriteString("unknown")
	}

	return builder.String()
}

type Validator[T any] struct {
	visited map[uintptr]bool
	deep    int
	err     error

	bts []byte
}

func NewValidator[T any]() *Validator[T] {
	return &Validator[T]{visited: make(map[uintptr]bool), deep: 0}
}

func (v *Validator[T]) From(bts []byte) *Validator[T] {
	v.bts = bts
	return v
}

func (v *Validator[T]) Stream(read io.Reader) *Validator[T] {
	var bts, err = io.ReadAll(read)
	if err != nil {
		v.err = err
		return v
	}
	v.bts = bts
	return v
}

func (v *Validator[T]) Bind(t T) error {

	if v.err != nil {
		return v.err
	}

	var err = json.Unmarshal(v.bts, t)
	if err != nil {
		return err
	}

	err = v.do(t)
	if err != nil {
		return err
	}

	v.visited = make(map[uintptr]bool)
	v.deep = 0

	return nil
}

func (v *Validator[T]) do(r any) error {

	var rv = reflect.ValueOf(r)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return errors.New("must be struct")
	}

	var err = v.printStruct(rv)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator[T]) format(rv reflect.Value) error {

	switch rv.Kind() {

	// SIMPLE TYPE
	case reflect.Bool:

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Complex64, reflect.Complex128:

	case reflect.Float32, reflect.Float64:

	case reflect.String:

	case reflect.Func:

	case reflect.UnsafePointer:

	case reflect.Chan:

	case reflect.Invalid:

	// COMPLEX TYPE
	case reflect.Map:
		if err := v.printMap(rv); err != nil {
			return err
		}
	case reflect.Struct:
		if err := v.printStruct(rv); err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		if err := v.printSlice(rv); err != nil {
			return err
		}
	case reflect.Ptr:
		if rv.CanInterface() {
			if err := v.printPtr(rv); err != nil {
				return err
			}
		}
	case reflect.Interface:
		if err := v.format(rv.Elem()); err != nil {
			return err
		}
	default:
		panic("unhandled default case")
	}

	return nil
}

func (v *Validator[T]) printMap(rv reflect.Value) error {

	var d = v.deep
	v.deep++

	if rv.Len() == 0 {
		v.deep = d
		return nil
	}

	if v.visited[rv.Pointer()] {
		v.deep = d
		return nil
	}

	v.visited[rv.Pointer()] = true

	keys := rv.MapKeys()
	for i := 0; i < rv.Len(); i++ {
		value := rv.MapIndex(keys[i])
		var err = v.format(value)
		if err != nil {
			return err
		}
	}

	v.deep = d

	return nil
}

func (v *Validator[T]) printSlice(rv reflect.Value) error {

	var d = v.deep
	v.deep++

	if rv.Len() == 0 {
		v.deep = d
		return nil
	}

	//  if is array, will be handled in printPtr
	if rv.Kind() == reflect.Slice {
		if v.visited[rv.Pointer()] {
			v.deep = d
			return nil
		}
		v.visited[rv.Pointer()] = true
	}

	for i := 0; i < rv.Len(); i++ {
		value := rv.Index(i)
		var err = v.format(value)
		if err != nil {
			return err
		}
	}

	v.deep = d

	return nil
}

func (v *Validator[T]) printStruct(rv reflect.Value) error {

	var d = v.deep
	v.deep++

	if rv.NumField() == 0 {
		v.deep = d
		return nil
	}

	for i := 0; i < rv.NumField(); i++ {
		value := rv.Field(i)

		var key = rv.Type().String() + ":" + strconv.Itoa(i)

		mux1.Lock()
		var typ = globalFields[key]
		if typ == nil {
			var a = rv.Type().Field(i)
			globalFields[key] = &a
			typ = &a
		}
		mux1.Unlock()

		if err := validate(*typ, value); err != nil {
			return err
		}

		// if is private
		// config private & public
		if value.CanInterface() {
			var err = v.format(value)
			if err != nil {
				return err
			}
		} else {

		}
	}

	v.deep = d

	return nil
}

func (v *Validator[T]) printPtr(rv reflect.Value) error {

	if v.visited[rv.Pointer()] {
		return nil
	}

	if rv.Pointer() != 0 {
		v.visited[rv.Pointer()] = true
	}

	if !rv.CanInterface() {
		return nil
	}

	if rv.Elem().IsValid() {
		var err = v.format(rv.Elem())
		if err != nil {
			return err
		}
	}

	return nil
}

func validate(t reflect.StructField, v reflect.Value) error {

	var tag = t.Tag.Get("validate")
	if tag == "" {
		return nil
	}
	mux.Lock()
	var exists = globalTags[tag]
	if exists == nil {
		var tags = strings.Split(tag, ",")
		if len(tags) == 0 {
			mux.Unlock()
			return nil
		}

		var parse = parseTag(tags)
		globalTags[tag] = &parse
		exists = &parse
	}
	mux.Unlock()

	var parse = *exists

	//var tags = strings.Split(tag, ",")
	//if len(tags) == 0 {
	//	return nil
	//}

	var key = t.Tag.Get("json")
	if key == "" {
		key = t.Name
	}

	//var parse = parseTag(tags)
	if parse.Required {
		if v.IsZero() {
			switch v.Kind() {
			case reflect.Bool:
				return &InvalidError[bool]{
					Key:   key,
					Type:  v.Type().String(),
					Value: v.Bool(),
					//Contract: "required",
					Op: "required",
				}
			case reflect.String:
				return &InvalidError[string]{
					Key:   key,
					Type:  v.Type().String(),
					Value: v.String(),
					//Contract: "required",
					Op: "required",
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return &InvalidError[int64]{
					Key:   key,
					Type:  v.Type().String(),
					Value: v.Int(),
					//Contract: "required",
					Op: "required",
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return &InvalidError[uint64]{
					Key:   key,
					Type:  v.Type().String(),
					Value: v.Uint(),
					//Contract: "required",
					Op: "required",
				}
			case reflect.Float32, reflect.Float64:
				return &InvalidError[float64]{
					Key:   key,
					Type:  v.Type().String(),
					Value: v.Float(),
					//Contract: "required",
					Op: "required",
				}
			default:
				return &InvalidError[string]{
					Key:   key,
					Type:  v.Type().String(),
					Value: "nil",
					//Contract: "required",
					Op: "required",
				}
			}
		}
	}

	if parse.NonEmpty {
		if v.Kind() == reflect.Array || v.Kind() == reflect.Slice || v.Kind() == reflect.Map {
			if v.Len() == 0 {
				var vv string

				switch v.Kind() {
				case reflect.Array, reflect.Slice:
					vv = "[]"
				case reflect.Map:
					vv = "{}"
				default:
					vv = "unknown"
				}

				return &InvalidError[string]{
					Key:   key,
					Type:  v.Type().String(),
					Value: vv,
					//Contract: "required",
					Op: "nonempty",
				}
			}
		}
	}

	// default
	if parse.Default != "" {
		if v.IsZero() {
			if v.CanSet() {
				switch v.Kind() {
				case reflect.Bool:
					var val = parse.Default == "true"
					v.SetBool(val)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					var val, _ = strconv.ParseInt(parse.Default, 10, 64)
					v.SetInt(val)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					var val, _ = strconv.ParseUint(parse.Default, 10, 64)
					v.SetUint(val)
				case reflect.Float32, reflect.Float64:
					var val, _ = strconv.ParseFloat(parse.Default, 64)
					v.SetFloat(val)
				case reflect.String:
					v.SetString(parse.Default)
				default:

				}
			}
		}
	}

	// gte
	if parse.Gte != "" {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var val, _ = strconv.ParseInt(parse.Gte, 10, 64)
			var vv = v.Int()
			if vv < val {
				return &InvalidError[int64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Gte,
					Op:       "gte",
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var val, _ = strconv.ParseUint(parse.Gte, 10, 64)
			var vv = v.Uint()
			if vv < val {
				return &InvalidError[uint64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Gte,
					Op:       "gte",
				}
			}
		case reflect.Float32, reflect.Float64:
			var val, _ = strconv.ParseFloat(parse.Gte, 64)
			var vv = v.Float()
			if vv < val {
				return &InvalidError[float64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Gte,
					Op:       "gte",
				}
			}
		default:

		}
	}

	// lte
	if parse.Lte != "" {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var val, _ = strconv.ParseInt(parse.Lte, 10, 64)
			var vv = v.Int()
			if vv > val {
				return &InvalidError[int64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Lte,
					Op:       "lte",
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var val, _ = strconv.ParseUint(parse.Lte, 10, 64)
			var vv = v.Uint()
			if vv > val {
				return &InvalidError[uint64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Lte,
					Op:       "lte",
				}
			}
		case reflect.Float32, reflect.Float64:
			var val, _ = strconv.ParseFloat(parse.Lte, 64)
			var vv = v.Float()
			if vv > val {
				return &InvalidError[float64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Lte,
					Op:       "lte",
				}
			}
		default:

		}
	}

	// gt
	if parse.Gt != "" {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var val, _ = strconv.ParseInt(parse.Gt, 10, 64)
			var vv = v.Int()
			if vv <= val {
				return &InvalidError[int64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Gt,
					Op:       "gt",
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var val, _ = strconv.ParseUint(parse.Gt, 10, 64)
			var vv = v.Uint()
			if vv <= val {
				return &InvalidError[uint64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Gt,
					Op:       "gt",
				}
			}
		case reflect.Float32, reflect.Float64:
			var val, _ = strconv.ParseFloat(parse.Gt, 64)
			var vv = v.Float()
			if vv <= val {
				return &InvalidError[float64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Gt,
					Op:       "gt",
				}
			}
		default:

		}
	}

	// lt
	if parse.Lt != "" {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var val, _ = strconv.ParseInt(parse.Lt, 10, 64)
			var vv = v.Int()
			if vv >= val {
				return &InvalidError[int64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Lt,
					Op:       "lt",
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var val, _ = strconv.ParseUint(parse.Lt, 10, 64)
			var vv = v.Uint()
			if vv >= val {
				return &InvalidError[uint64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Lt,
					Op:       "lt",
				}
			}
		case reflect.Float32, reflect.Float64:
			var val, _ = strconv.ParseFloat(parse.Lt, 64)
			var vv = v.Float()
			if vv >= val {
				return &InvalidError[float64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Lt,
					Op:       "lt",
				}
			}
		default:

		}
	}

	// eq
	if parse.Eq != "" {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var val, _ = strconv.ParseInt(parse.Eq, 10, 64)
			var vv = v.Int()
			if vv != val {
				return &InvalidError[int64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Eq,
					Op:       "eq",
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var val, _ = strconv.ParseUint(parse.Eq, 10, 64)
			var vv = v.Uint()
			if vv != val {
				return &InvalidError[uint64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Eq,
					Op:       "eq",
				}
			}
		case reflect.Float32, reflect.Float64:
			var val, _ = strconv.ParseFloat(parse.Eq, 64)
			var vv = v.Float()
			if vv != val {
				return &InvalidError[float64]{
					Key:      key,
					Type:     v.Type().String(),
					Value:    vv,
					Contract: parse.Eq,
					Op:       "eq",
				}
			}
		default:

		}
	}

	return nil
}

type Type struct {
	Gte string
	Lte string
	Gt  string
	Lt  string
	Eq  string

	Required bool
	NonEmpty bool
	Default  string
}

func parseTag(arr []string) Type {
	var t = Type{}
	for i := 0; i < len(arr); i++ {
		if arr[i] == "required" {
			t.Required = true
			continue
		}

		if arr[i] == "nonempty" {
			t.NonEmpty = true
			continue
		}

		var arr1 = strings.Split(arr[i], ":")
		if len(arr1) < 2 {
			continue
		}

		switch arr1[0] {
		case "gte":
			t.Gte = arr1[1]
		case "lte":
			t.Lte = arr1[1]
		case "gt":
			t.Gt = arr1[1]
		case "lt":
			t.Lt = arr1[1]
		case "eq":
			t.Eq = arr1[1]
		case "default":
			t.Default = arr1[1]
		}
	}

	return t
}
