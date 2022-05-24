/**
* @program: kitty
*
* @description:
*
* @author: lemon
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

func Trace(url string) *info {
	var c = &Client{}
	return c.Trace(url)
}

func Patch(url string) *info {
	var c = &Client{}
	return c.Patch(url)
}

func Put(url string) *info {
	var c = &Client{}
	return c.Put(url)
}

func Delete(url string) *info {
	var c = &Client{}
	return c.Delete(url)
}

func Options(url string) *info {
	var c = &Client{}
	return c.Options(url)
}
