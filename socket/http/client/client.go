/**
* @program: proxy-server
*
* @description:
*
* @author: lemon
*
* @create: 2019-10-03 13:37
**/

package client

import (
	"net/http"
)

type Client struct {
	method string
	url    string
}

func (h *Client) Post(url string) *info {
	h.method = http.MethodPost
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Get(url string) *info {
	h.method = http.MethodGet
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Head(url string) *info {
	h.method = http.MethodHead
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Trace(url string) *info {
	h.method = http.MethodTrace
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Options(url string) *info {
	h.method = http.MethodOptions
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Put(url string) *info {
	h.method = http.MethodPut
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Delete(url string) *info {
	h.method = http.MethodDelete
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}

func (h *Client) Patch(url string) *info {
	h.method = http.MethodPatch
	h.url = url
	var info = &info{handler: h, clientTimeout: clientTimeout}
	return info
}
