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

type Response struct {
	err  error
	code int
	buf  *bytes.Buffer
	req  *http.Response
}

func (r *Response) String() string {
	return r.buf.String()
}

func (r *Response) Bytes() []byte {
	return r.buf.Bytes()
}

func (r *Response) Code() int {
	return r.code
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) Response() *http.Response {
	return r.req
}
