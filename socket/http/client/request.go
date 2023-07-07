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
	"github.com/lemonyxk/kitty/kitty"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"net/textproto"
	"time"

	"github.com/lemonyxk/kitty/kitty/header"
)

type Request struct {
	handler       *Client
	headerKey     []string
	headerValue   []string
	cookies       []*http.Cookie
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

func (h *Request) Protobuf(body ...proto.Message) *Protobuf {
	h.SetHeader(header.ContentType, header.ApplicationProtobuf)
	return &Protobuf{info: h, body: body}
}

func (h *Request) Json(body ...any) *Json {
	h.SetHeader(header.ContentType, header.ApplicationJson)
	return &Json{info: h, body: body}
}

func (h *Request) Query(body ...kitty.M) *URL {
	return &URL{info: h, body: body}
}

func (h *Request) Form(body ...kitty.M) *Form {
	h.SetHeader(header.ContentType, header.ApplicationFormUrlencoded)
	return &Form{info: h, body: body}
}

func (h *Request) Multipart(body ...kitty.M) *FormData {
	h.SetHeader(header.ContentType, header.MultipartFormData)
	return &FormData{info: h, body: body}
}

func (h *Request) OctetStream(body io.Reader) *OctetStream {
	h.SetHeader(header.ContentType, header.ApplicationOctetStream)
	return &OctetStream{info: h, body: body}
}

func (h *Request) Raw(body io.Reader) *Raw {
	return &Raw{info: h, body: body}
}
