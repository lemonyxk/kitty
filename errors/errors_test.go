/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2023-03-25 17:26
**/

package errors

import (
	"fmt"
	kitty2 "github.com/lemonyxk/kitty/kitty"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var err = New("test")
	if err.Error() != "test" {
		t.Fatal(err)
	}
}

func TestWrap(t *testing.T) {
	var err = Wrap(New("test"), "wrap")
	if err.Error() != "wrap: test" {
		t.Fatal(err)
	}
}

func TestWrapf(t *testing.T) {
	var err = Wrapf(New("test"), "wrap %s", "test")
	if err.Error() != "wrap test: test" {
		t.Fatal(err)
	}
}

func TestWrapNil(t *testing.T) {
	var err = Wrap(nil, "wrap")
	if err.Error() != "wrap" {
		t.Fatal(err)
	}
}

func TestWrapfNil(t *testing.T) {
	var err = Wrapf(nil, "wrap %s", "test")
	if err.Error() != "wrap test" {
		t.Fatal(err)
	}
}

func TestErrorf(t *testing.T) {
	var err = Errorf("test %s", "test")
	if err.Error() != "test test" {
		t.Fatal(err)
	}
}

func TestStack(t *testing.T) {
	var err = New("test")
	if err.Error() != "test" {
		t.Fatal(err)
	}

	var stack = fmt.Sprintf("%+v", err)
	if stack == "" {
		t.Fatal(stack)
	}

	assert.True(t, strings.Contains(stack, "TestStack"))
	assert.True(t, strings.Contains(stack, "errors_test.go:64"))
}

func TestIs(t *testing.T) {

	var fn = func() error {
		return StopPropagation
	}

	var err = fn()
	if !Is(err, StopPropagation) {
		t.Fatal(err)
	}
}

func BenchmarkIsNil(b *testing.B) {
	for i := 0; i < b.N; i++ {
		kitty2.IsNil(&Error{})
	}
}

func BenchmarkIsNil2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reflect.ValueOf(&Error{}).IsNil()
	}
}
