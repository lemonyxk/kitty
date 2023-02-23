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

import (
	"crypto/tls"
	"net/http"
	url2 "net/url"
	"time"
)

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

func TSLConfig(tlsConfig *tls.Config) {
	defaultTlsConfig = tlsConfig
}

func Proxy(url string) {
	var fixUrl, err = url2.Parse(url)
	if err != nil {
		panic(err)
	}
	defaultProxy = http.ProxyURL(fixUrl)
}

func KeepAlive(keepalive time.Duration) {
	defaultDialer.KeepAlive = keepalive
}

func Timeout(timeout time.Duration) {
	defaultClient.Timeout = timeout
}