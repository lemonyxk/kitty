/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-21 17:36
**/

package client

import "net/http"

type Req struct {
	err  error
	code int
	data []byte
	req  *http.Response
}

func (r *Req) String() string {
	return string(r.data)
}

func (r *Req) Bytes() []byte {
	return r.data
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
