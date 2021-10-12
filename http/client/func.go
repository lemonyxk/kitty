/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-10-13 02:26
**/

package client

func Post(url string) *info {
	var c = &Client{}
	return c.Post(url)
}

func Get(url string) *info {
	var c = &Client{}
	return c.Get(url)
}

func Head(url string) *info {
	var c = &Client{}
	return c.Head(url)
}
