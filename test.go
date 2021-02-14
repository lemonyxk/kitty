/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2020-09-18 16:58
**/

package kitty

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func NewAssertEqual(t *testing.T) func(condition bool, args ...interface{}) {
	return func(condition bool, args ...interface{}) {
		AssertEqual(t, condition, args...)
	}
}

func AssertEqual(t *testing.T, condition bool, args ...interface{}) {
	if !condition {
		t.Fatal(args...)
	}
}

type TestResponse struct {
	Data     string
	Response *http.Response
}

func TestGet(url string, params M) TestResponse {
	var p = ""
	for k, v := range params {
		p += fmt.Sprintf("%v=%v&", k, v)
	}
	if len(p) > 0 {
		p = p[:len(p)-1]
	}
	resp, err := http.Get(url + "?" + p)
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return TestResponse{Data: string(bodyBytes), Response: resp}
}

func TestPost(url string, params M) TestResponse {
	var p = ""
	for k, v := range params {
		p += fmt.Sprintf("%v=%v&", k, v)
	}
	if len(p) > 0 {
		p = p[:len(p)-1]
	}
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(p))
	if err != nil {
		panic(err)
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return TestResponse{Data: string(body), Response: resp}
}
