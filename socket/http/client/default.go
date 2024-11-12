/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-21 17:48
**/

package client

import (
	"crypto/tls"
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

var defaultTransport = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DisableCompression:    false,
	DisableKeepAlives:     false,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 15 * time.Second,
	ExpectContinueTimeout: 2 * time.Second,
	IdleConnTimeout:       60 * time.Second,
	MaxIdleConns:          runtime.NumCPU() * 8,
	MaxIdleConnsPerHost:   runtime.NumCPU() * 4,
	MaxConnsPerHost:       runtime.NumCPU() * 8,
	DialContext:           defaultDialer.DialContext,
	TLSClientConfig:       &tls.Config{},
}

var defaultClient = http.Client{
	Timeout:   clientTimeout,
	Transport: defaultTransport,
}
