/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2023-03-25 17:36
**/

package router

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func Test_Router_Group(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Group("/test").Handler(func(handler *Handler[int]) {
		handler.Get("/test").Handler(f)
	})

	a, b := r.GetRoute("/test/test")
	assert.Equal(t, string(b), "/test/test")
	assert.True(t, string(a.Path) == "/test/test")
}

func Test_Router_Group2(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Group("/test").Handler(func(handler *Handler[int]) {
		handler.Get("/test").Handler(f)
	})

	g.Group("/test").Handler(func(handler *Handler[int]) {
		handler.Get("/test2").Handler(f)
	})

	a, b := r.GetRoute("/test/test")
	assert.Equal(t, string(b), "/test/test")
	assert.True(t, string(a.Path) == "/test/test")

	a, b = r.GetRoute("/test/test2")
	assert.Equal(t, string(b), "/test/test2")
	assert.True(t, string(a.Path) == "/test/test2")
}

func Test_Router_Get(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Get("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "GET")
}

func Test_Router_Post(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Post("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "POST")
}

func Test_Router_Put(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Put("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "PUT")
}

func Test_Router_Delete(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Delete("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "DELETE")
}

func Test_Router_Patch(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Patch("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "PATCH")
}

func Test_Router_Head(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Head("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "HEAD")
}

func Test_Router_Options(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Options("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "OPTIONS")
}

func Test_Router_Connect(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Connect("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "CONNECT")
}

func Test_Router_Trace(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Trace("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "TRACE")
}

func Test_Router_Multi_Method(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Method("GET", "POST").Route("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "GET")
	assert.True(t, a.Data.Method[1] == "POST")
}

func Test_Router_nil(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Method("GET", "POST").Route("/test").Handler(f)

	a, b := r.GetRoute("/test2")
	assert.Equal(t, len(b), 0)
	assert.True(t, a == nil)
}

func Test_Router_Ptr(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Post("/test").Handler(f)

	a, b := r.GetRoute("/test")
	assert.Equal(t, string(b), "/test")
	assert.True(t, string(a.Path) == "/test")
	assert.True(t, a.Data.Method[0] == "POST")
	// check f pointer address if equal
	var fPtr = fmt.Sprintf("%p", f)
	var aPtr = fmt.Sprintf("%p", a.Data.Function)
	assert.True(t, fPtr == aPtr)
}

func Test_Router_Params(t *testing.T) {
	var r = &Router[int]{}
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Post("/test/:id/:name").Handler(f)

	a, b := r.GetRoute("/test/1/tiny")
	assert.Equal(t, string(b), "/test/1/tiny")
	assert.True(t, string(a.Path) == "/test/:id/:name")
	assert.True(t, a.Data.Method[0] == "POST")
	assert.True(t, a.ParseParams(b)[0] == "1")
	assert.True(t, a.ParseParams(b)[1] == "tiny")
	assert.True(t, a.Keys[0] == "id")
	assert.True(t, a.Keys[1] == "name")
}

func Test_Router_StrictMode(t *testing.T) {
	var r = &Router[int]{}
	r.StrictMode = true
	var g = r.Create()
	var f = func(stream int) error { return nil }
	g.Post("/Test/:id/:name").Handler(f)

	a, b := r.GetRoute("/test/1/tiny")
	assert.True(t, a == nil)
	assert.True(t, len(b) == 0)

	r = &Router[int]{}
	r.StrictMode = false
	g = r.Create()
	g.Post("/Test/:id/:name").Handler(f)

	a, b = r.GetRoute("/test/1/tiny")
	assert.True(t, a != nil)
	assert.True(t, len(b) > 0)

	a, b = r.GetRoute("/Test/1/tiny")
	assert.True(t, a != nil)
	assert.True(t, len(b) > 0)
}

func Test_Router_Before(t *testing.T) {
	var r = &Router[int]{}
	r.StrictMode = false
	var g = r.Create()
	var f = func(stream int) error { return nil }
	var b1 = func(stream int) error { return nil }
	var b2 = func(stream int) error { return nil }
	var gg = g.Post("/Test/Before")
	gg.Before(b1).Before(b2).Handler(f)

	a, b := r.GetRoute("/Test/Before")
	assert.True(t, a != nil)
	assert.True(t, len(b) > 0)

	assert.True(t, len(a.Data.Before) == 2)

	gg = g.Post("/Test/Before1")
	gg.Before(b1).RemoveBefore(b1).Before(b2).Handler(f)
	a, b = r.GetRoute("/Test/Before1")
	assert.True(t, a != nil)
	assert.True(t, len(b) > 0)

	assert.True(t, len(a.Data.Before) == 1, "RemoveBefore failed", len(a.Data.Before))

	gg.CancelBefore()
	assert.True(t, len(gg.before) == 0, "RemoveBefore failed", len(gg.before))

	gg.Before(b1).Before(b2)
	assert.True(t, len(gg.before) == 2, "RemoveBefore failed", len(gg.before))

	gg.RemoveBefore(b2)
	assert.True(t, equal(gg.before[0], b1) && len(gg.before) == 1, "RemoveBefore failed", len(gg.before))
}

func Test_Router_After(t *testing.T) {
	var r = &Router[int]{}
	r.StrictMode = false
	var g = r.Create()
	var f = func(stream int) error { return nil }
	var b1 = func(stream int) error { return nil }
	var b2 = func(stream int) error { return nil }
	var gg = g.Post("/Test/After")
	gg.After(b1).After(b2).Handler(f)

	a, b := r.GetRoute("/Test/After")
	assert.True(t, a != nil)
	assert.True(t, len(b) > 0)

	assert.True(t, len(a.Data.After) == 2)

	gg = g.Post("/Test/After1")
	gg.After(b1).RemoveAfter(b1).After(b2).Handler(f)
	a, b = r.GetRoute("/Test/After1")
	assert.True(t, a != nil)
	assert.True(t, len(b) > 0)

	assert.True(t, len(a.Data.After) == 1, "RemoveAfter failed", len(a.Data.After))

	gg.CancelAfter()
	assert.True(t, len(gg.after) == 0, "RemoveAfter failed", len(gg.after))

	gg.After(b1).After(b2)
	assert.True(t, len(gg.after) == 2, "RemoveAfter failed", len(gg.after))

	gg.RemoveAfter(b2)
	assert.True(t, equal(gg.after[0], b1) && len(gg.after) == 1, "RemoveAfter failed", len(gg.after))
}

func equal(a, b func(stream int) error) bool {
	return *(*unsafe.Pointer)(unsafe.Pointer(&a)) == *(*unsafe.Pointer)(unsafe.Pointer(&b))
}
