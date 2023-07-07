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
	"context"
	"github.com/lemonyxk/kitty/kitty"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

// Raw

type Raw struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   io.Reader
}

func (p *Raw) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doRaw(p.info.handler.method, p.info.handler.url, p.info, p.body)
	return send(p.info, p.req, p.cancel)
}

func (p *Raw) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doRaw(p.info.handler.method, p.info.handler.url, p.info, p.body)
	return do(p.info, p.req, p.cancel)
}

func (p *Raw) Abort() {
	p.cancel()
}

// OctetStream

type OctetStream struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   io.Reader
}

func (p *OctetStream) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doRaw(p.info.handler.method, p.info.handler.url, p.info, p.body)
	return send(p.info, p.req, p.cancel)
}

func (p *OctetStream) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doRaw(p.info.handler.method, p.info.handler.url, p.info, p.body)
	return do(p.info, p.req, p.cancel)
}

func (p *OctetStream) Abort() {
	p.cancel()
}

// Json

type Json struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   []any
}

func (p *Json) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doJson(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return send(p.info, p.req, p.cancel)
}

func (p *Json) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doJson(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return do(p.info, p.req, p.cancel)
}

func (p *Json) Abort() {
	p.cancel()
}

// Form

type Form struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   []kitty.M
}

func (p *Form) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doFormUrlencoded(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return send(p.info, p.req, p.cancel)
}

func (p *Form) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doFormUrlencoded(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return do(p.info, p.req, p.cancel)
}

func (p *Form) Abort() {
	p.cancel()
}

// FormData

type FormData struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   []kitty.M
}

func (p *FormData) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doFormData(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return send(p.info, p.req, p.cancel)
}

func (p *FormData) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doFormData(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return do(p.info, p.req, p.cancel)
}

func (p *FormData) Abort() {
	p.cancel()
}

// Protobuf

type Protobuf struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   []proto.Message
}

func (p *Protobuf) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doProtobuf(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return send(p.info, p.req, p.cancel)
}

func (p *Protobuf) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doProtobuf(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return do(p.info, p.req, p.cancel)
}

func (p *Protobuf) Abort() {
	p.cancel()
}

// URL

type URL struct {
	info   *Request
	req    *http.Request
	cancel context.CancelFunc
	err    error
	body   []kitty.M
}

func (p *URL) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	p.req, p.cancel, p.err = doUrl(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return send(p.info, p.req, p.cancel)
}

func (p *URL) Do() (*http.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.req, p.cancel, p.err = doUrl(p.info.handler.method, p.info.handler.url, p.info, p.body...)
	return do(p.info, p.req, p.cancel)
}

func (p *URL) Abort() {
	p.cancel()
}
