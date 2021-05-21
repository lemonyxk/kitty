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

type request struct {
	err      error
	code     int
	data     []byte
	response *http.Response
}

func (r *request) String() string {
	return string(r.data)
}

func (r *request) Bytes() []byte {
	return r.data
}

func (r *request) Code() int {
	return r.code
}

func (r *request) LastError() error {
	return r.err
}

func (r *request) Response() *http.Response {
	return r.response
}
