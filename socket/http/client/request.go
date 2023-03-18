/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-21 17:37
**/

package client

import (
	"io"
	"net/http"
	"net/textproto"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/v2/kitty"
)

type Request struct {
	handler       *Client
	headerKey     []string
	headerValue   []string
	cookies       []*http.Cookie
	body          any
	progress      *Progress
	userName      string
	passWord      string
	clientTimeout time.Duration
}

func (h *Request) Progress(progress *Progress) *Request {
	h.progress = progress
	return h
}

func (h *Request) Timeout(timeout time.Duration) *Request {
	h.clientTimeout = timeout
	return h
}

func (h *Request) SetBasicAuth(userName, passWord string) *Request {
	h.userName = userName
	h.passWord = passWord
	return h
}

func (h *Request) SetHeaders(headers map[string]string) *Request {
	h.headerKey = nil
	h.headerValue = nil
	for key, value := range headers {
		h.headerKey = append(h.headerKey, textproto.CanonicalMIMEHeaderKey(key))
		h.headerValue = append(h.headerValue, value)
	}
	return h
}

func (h *Request) AddHeader(key string, value string) *Request {
	h.headerKey = append(h.headerKey, textproto.CanonicalMIMEHeaderKey(key))
	h.headerValue = append(h.headerValue, value)
	return h
}

func (h *Request) SetHeader(key string, value string) *Request {
	for i := 0; i < len(h.headerKey); i++ {
		if textproto.CanonicalMIMEHeaderKey(h.headerKey[i]) == textproto.CanonicalMIMEHeaderKey(key) {
			h.headerValue[i] = value
			return h
		}
	}

	h.headerKey = append(h.headerKey, key)
	h.headerValue = append(h.headerValue, value)
	return h
}

func (h *Request) SetCookies(cookies []*http.Cookie) *Request {
	h.cookies = cookies
	return h
}

func (h *Request) AddCookie(cookie *http.Cookie) *Request {
	for i := 0; i < len(h.cookies); i++ {
		if h.cookies[i].String() == cookie.String() {
			h.cookies[i] = cookie
			return h
		}
	}
	h.cookies = append(h.cookies, cookie)
	return h
}

func (h *Request) Protobuf(body ...proto.Message) *Sender {
	h.SetHeader(kitty.ContentType, kitty.ApplicationProtobuf)
	h.body = body
	request, cancel, err := getRequest(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}

func (h *Request) Json(body ...any) *Sender {
	h.SetHeader(kitty.ContentType, kitty.ApplicationJson)
	h.body = body
	request, cancel, err := getRequest(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}

func (h *Request) Query(body ...kitty.M) *Sender {
	h.body = body
	request, cancel, err := getRequest(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}

func (h *Request) Form(body ...kitty.M) *Sender {
	h.SetHeader(kitty.ContentType, kitty.ApplicationFormUrlencoded)
	h.body = body
	request, cancel, err := getRequest(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}

func (h *Request) Multipart(body ...kitty.M) *Sender {
	h.SetHeader(kitty.ContentType, kitty.MultipartFormData)
	h.body = body
	request, cancel, err := getRequest(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}

func (h *Request) OctetStream(r io.Reader) *Sender {
	h.SetHeader(kitty.ContentType, kitty.ApplicationOctetStream)
	h.body = r
	request, cancel, err := doRaw(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}

func (h *Request) Raw(r io.Reader) *Sender {
	h.body = r
	request, cancel, err := doRaw(h.handler.method, h.handler.url, h)
	if err != nil {
		return &Sender{err: err}
	}
	return &Sender{info: h, req: request, cancel: cancel}
}