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
	"net"
	"net/http"
	url2 "net/url"
	"time"
)

func Post(url string) *Request {
	var c = &Client{}
	return c.Post(url)
}

func Get(url string) *Request {
	var c = &Client{}
	return c.Get(url)
}

func Head(url string) *Request {
	var c = &Client{}
	return c.Head(url)
}

func Trace(url string) *Request {
	var c = &Client{}
	return c.Trace(url)
}

func Patch(url string) *Request {
	var c = &Client{}
	return c.Patch(url)
}

func Put(url string) *Request {
	var c = &Client{}
	return c.Put(url)
}

func Delete(url string) *Request {
	var c = &Client{}
	return c.Delete(url)
}

func Options(url string) *Request {
	var c = &Client{}
	return c.Options(url)
}

func TSLConfig(tlsConfig *tls.Config) {
	defaultTransport.TLSClientConfig = tlsConfig
}

func Proxy(url string) {
	var fixUrl, err = url2.Parse(url)
	if err != nil {
		panic(err)
	}
	defaultTransport.Proxy = http.ProxyURL(fixUrl)
}

func Transport() *http.Transport {
	return defaultTransport
}

func KeepAlive(keepalive time.Duration) {
	var dialer = net.Dialer{
		Timeout:   dialerTimeout,
		KeepAlive: keepalive,
	}
	defaultTransport.DialContext = dialer.DialContext
}

func Timeout(timeout time.Duration) {
	defaultClient.Timeout = timeout
}
