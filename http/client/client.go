/**
* @program: proxy-server
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-03 13:37
**/

package client

import (
	"net/http"

	"github.com/lemoyxk/kitty"
)

type client struct {
	method string
	url    string
}

func (h *client) Post(url string) *info {
	h.method = http.MethodPost
	h.url = url
	var info = &info{handler: h}
	info.SetHeader(kitty.ContentType, kitty.ApplicationFormUrlencoded)
	return info
}

func (h *client) Get(url string) *info {
	h.method = http.MethodGet
	h.url = url
	var info = &info{handler: h}
	return info
}

func (h *client) Head(url string) *info {
	h.method = http.MethodHead
	h.url = url
	var info = &info{handler: h}
	return info
}
