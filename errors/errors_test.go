/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-03-25 17:26
**/

package errors

import (
	"fmt"
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
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrapfNil(t *testing.T) {
	var err = Wrapf(nil, "wrap %s", "test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewWithStack(t *testing.T) {
	var err = NewWithStack("test")
	if err.Error() != "test" {
		t.Fatal(err)
	}
}

func TestWithStack(t *testing.T) {
	var err = WithStack(New("test"))
	if err.Error() != "test" {
		t.Fatal(err)
	}
}

func TestWithStackNil(t *testing.T) {
	var err = WithStack(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorf(t *testing.T) {
	var err = Errorf("test %s", "test")
	if err.Error() != "test test" {
		t.Fatal(err)
	}
}

func TestWrapWithStack(t *testing.T) {
	var err = Wrap(NewWithStack("test"), "wrap")
	if err.Error() != "wrap: test" {
		t.Fatal(err)
	}
}

func TestWrapfWithStack(t *testing.T) {
	var err = Wrapf(NewWithStack("test"), "wrap %s", "test")
	if err.Error() != "wrap test: test" {
		t.Fatal(err)
	}
}

func TestStack(t *testing.T) {
	var err = New("test")
	if err.Stack().Error() != "test" {
		t.Fatal(err)
	}

	var stack = fmt.Sprintf("%+v", err.Stack())
	if stack == "" {
		t.Fatal(stack)
	}

	assert.True(t, strings.Contains(stack, "TestStack"))
	assert.True(t, strings.Contains(stack, "errors_test.go:104"))
}
