/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-21 17:36
**/

package client

import (
	"bytes"
	"net/http"
)

type Req struct {
	err  error
	code int
	buf  *bytes.Buffer
	req  *http.Response
}

func (r *Req) String() string {
	return r.buf.String()
}

func (r *Req) Bytes() []byte {
	return r.buf.Bytes()
}

func (r *Req) Code() int {
	return r.code
}

func (r *Req) LastError() error {
	return r.err
}

func (r *Req) Response() *http.Response {
	return r.req
}
