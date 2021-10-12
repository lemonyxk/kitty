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

func NewProgress() *progress {
	return &progress{}
}

func NewHttpClient() *client {
	return &client{}
}

func Post(url string) *info {
	var c = &client{}
	return c.Post(url)
}

func Get(url string) *info {
	var c = &client{}
	return c.Get(url)
}

func Head(url string) *info {
	var c = &client{}
	return c.Head(url)
}
