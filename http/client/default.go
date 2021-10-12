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

import (
	"net"
	"net/http"
	"runtime"
	"time"
)

const dialerTimeout = 30 * time.Second
const dialerKeepAlive = 30 * time.Second

const clientTimeout = 15 * time.Second

var defaultDialer = net.Dialer{
	Timeout:   dialerTimeout,
	KeepAlive: dialerKeepAlive,
}

var defaultTransport = http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DisableCompression:    false,
	DisableKeepAlives:     false,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 15 * time.Second,
	ExpectContinueTimeout: 2 * time.Second,
	MaxIdleConns:          runtime.NumCPU() * 2,
	MaxIdleConnsPerHost:   runtime.NumCPU() * 2,
	MaxConnsPerHost:       runtime.NumCPU() * 2,
	DialContext:           defaultDialer.DialContext,
}

var defaultClient = http.Client{
	Timeout:   clientTimeout,
	Transport: &defaultTransport,
}
