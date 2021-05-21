/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-21 17:48
**/

package client

import "net/http"

func NewProgress() *progress {
	return &progress{}
}

func New() *client {
	return &client{}
}

func Post(url string) *info {
	var info = &client{method: http.MethodPost, url: url}
	return info.Post(url)
}

func Get(url string) *info {
	var info = &client{method: http.MethodGet, url: url}
	return info.Get(url)
}

func Head(url string) *info {
	var info = &client{method: http.MethodHead, url: url}
	return info.Head(url)
}
