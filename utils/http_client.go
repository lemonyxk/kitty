/**
* @program: proxy-server
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-03 13:37
**/

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	url2 "net/url"
	"os"
	"strconv"
	"time"

	"github.com/json-iterator/go"
)

type hc int

const HttpClient hc = iota

const applicationFormUrlencoded = "application/x-www-form-urlencoded"
const applicationJson = "application/json"
const multipartFormData = "multipart/form-data"

var dialer = net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var client = http.Client{
	Timeout: 15 * time.Second,
}

var transport = http.Transport{
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 15 * time.Second,
	ExpectContinueTimeout: 2 * time.Second,
}

type writeProgress struct {
	total      int64
	current    int64
	onProgress func(p []byte, current int64, total int64)
	last       int64
	rate       int64
}

func (w *writeProgress) Write(p []byte) (int, error) {
	n := len(p)
	w.current += int64(n)

	if w.total == 0 {
		w.onProgress(p, w.current, -1)
	} else {
		w.last += int64(n) * w.rate

		if w.last >= w.total {
			w.onProgress(p, w.current, w.total)
			w.last = w.last - w.total
		}
	}

	return n, nil
}

func (h hc) NewProgress() *progress {
	return &progress{}
}

type progress struct {
	rate     int64
	progress func(p []byte, current int64, total int64)
}

// 0.01 - 100
func (p *progress) Rate(rate float64) *progress {
	if rate < 0.01 || rate > 100 {
		rate = 1
	}
	p.rate = int64(100 / rate)
	return p
}

func (p *progress) OnProgress(fn func(p []byte, current int64, total int64)) *progress {
	p.progress = fn
	return p
}

func do(httpClient *httpClient) *Request {

	var client = httpClient.client
	var method = httpClient.method
	var url = httpClient.url
	var headerKey = httpClient.headerKey
	var headerValue = httpClient.headerValue
	var body = httpClient.body
	var cookies = httpClient.cookies
	var progress = httpClient.progress

	var request *http.Request
	var response *http.Response
	var err error

	if method == http.MethodGet {

		Url, err := url2.Parse(url)
		if err != nil {
			return &Request{err: err}
		}

		if body == nil {
			body = make(map[string]interface{})
		}

		if _, ok := body.(map[string]interface{}); !ok {
			return &Request{err: errors.New("get method body must be map[string]interface")}
		}

		var params = url2.Values{}

		for key, value := range body.(map[string]interface{}) {
			switch value.(type) {
			case string:
				params.Set(key, value.(string))
			case int:
				params.Set(key, strconv.Itoa(value.(int)))
			case float64:
				params.Set(key, strconv.FormatFloat(value.(float64), 'f', -1, 64))
			default:
				params.Set(key, fmt.Sprintf("%v", value))
			}
		}

		Url.RawQuery = Url.RawQuery + "&" + params.Encode()

		request, err = http.NewRequest(method, Url.String(), nil)
		if err != nil {
			return &Request{err: err}
		}

	} else {

		var contentType = ""
		var headerIndex = -1

		for i := 0; i < len(headerKey); i++ {
			if headerKey[i] == "Content-Type" {
				contentType = headerValue[i]
				headerIndex = i
				break
			}
		}

		if contentType == "" {
			contentType = applicationFormUrlencoded
		}

		switch contentType {
		case applicationFormUrlencoded:

			if body == nil {
				body = make(map[string]interface{})
			}

			if _, ok := body.(map[string]interface{}); !ok {
				return &Request{err: errors.New("application/x-www-form-urlencoded body must be map[string]interface")}
			}

			var buff bytes.Buffer
			for key, value := range body.(map[string]interface{}) {
				switch value.(type) {
				case string:
					buff.WriteString(key + "=" + value.(string) + "&")
				case int:
					buff.WriteString(key + "=" + strconv.Itoa(value.(int)) + "&")
				case float64:
					buff.WriteString(key + "=" + strconv.FormatFloat(value.(float64), 'f', -1, 64) + "&")
				default:
					buff.WriteString(key + "=" + fmt.Sprintf("%v", value) + "&")
				}
			}

			var b = buff.Bytes()
			request, err = http.NewRequest(method, url, bytes.NewReader(b[:len(b)-1]))
			if err != nil {
				return &Request{err: err}
			}

		case applicationJson:

			jsonBody, err := jsoniter.Marshal(body)
			if err != nil {
				return &Request{err: err}
			}

			request, err = http.NewRequest(method, url, bytes.NewReader(jsonBody))
			if err != nil {
				return &Request{err: err}
			}

		case multipartFormData:

			if body == nil {
				body = make(map[string]interface{})
			}

			if _, ok := body.(map[string]interface{}); !ok {
				return &Request{err: errors.New("application/x-www-form-urlencoded body must be map[string]interface")}
			}

			var buf = new(bytes.Buffer)
			part := multipart.NewWriter(buf)

			for key, value := range body.(map[string]interface{}) {
				switch value.(type) {
				case string:
					if err := part.WriteField(key, value.(string)); err != nil {
						return &Request{err: err}
					}
				case int:
					if err := part.WriteField(key, strconv.Itoa(value.(int))); err != nil {
						return &Request{err: err}
					}
				case float64:
					if err := part.WriteField(key, strconv.FormatFloat(value.(float64), 'f', -1, 64)); err != nil {
						return &Request{err: err}
					}
				case *os.File:
					ff, err := part.CreateFormFile(key, value.(*os.File).Name())
					if err != nil {
						return &Request{err: err}
					}
					_, err = io.Copy(ff, value.(*os.File))
					if err != nil {
						return &Request{err: err}
					}
				default:
					if err = part.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
						return &Request{err: err}
					}
				}
			}

			if err := part.Close(); err != nil {
				return &Request{err: err}
			}

			request, err = http.NewRequest(method, url, buf)
			if err != nil {
				return &Request{err: err}
			}

			headerValue[headerIndex] = part.FormDataContentType()

		default:
			return &Request{err: errors.New("invalid context type")}
		}

	}

	if request == nil {
		return &Request{err: errors.New("invalid request")}
	}

	for i := 0; i < len(headerKey); i++ {
		request.Header.Add(headerKey[i], headerValue[i])
	}

	for i := 0; i < len(cookies); i++ {
		request.AddCookie(cookies[i])
	}

	response, err = client.Do(request)
	if err != nil {
		return &Request{err: err}
	}
	defer func() { _ = response.Body.Close() }()

	var dataBytes []byte

	if progress != nil {

		var total, _ = strconv.ParseInt(response.Header.Get("Content-Length"), 10, 64)

		var writer = &writeProgress{
			total:      total,
			onProgress: progress.progress,
			rate:       progress.rate,
		}

		dataBytes, err = ioutil.ReadAll(io.TeeReader(response.Body, writer))
		if err != nil {
			return &Request{err: err}
		}
	} else {
		dataBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return &Request{err: err}
		}
	}

	return &Request{err: nil, code: response.StatusCode, data: dataBytes, requestHeader: response.Request.Header, responseHeader: response.Header}
}

type httpClient struct {
	method      string
	url         string
	headerKey   []string
	headerValue []string
	cookies     []*http.Cookie
	body        interface{}
	client      *http.Client
	transport   *http.Transport
	dialer      *net.Dialer
	progress    *progress
}

type httpInfo struct {
	handler *httpClient
}

type Request struct {
	err            error
	code           int
	data           []byte
	responseHeader http.Header
	requestHeader  http.Header
}

func (r *Request) String() string {
	return string(r.data)
}

func (r *Request) Bytes() []byte {
	return r.data
}

func (r *Request) Code() int {
	return r.code
}

func (r *Request) LastError() error {
	return r.err
}

func (r *Request) ResponseHeader() http.Header {
	return r.responseHeader
}

func (r *Request) RequestHeader() http.Header {
	return r.requestHeader
}

func (h hc) New() *httpClient {
	return &httpClient{
		client:    &http.Client{},
		transport: &transport,
		dialer:    &net.Dialer{},
	}
}

func (h hc) Post(url string) *httpInfo {
	return (&httpClient{
		method:    http.MethodPost,
		url:       url,
		client:    &client,
		transport: &transport,
		dialer:    &dialer,
	}).Post(url)
}

func (h hc) Get(url string) *httpInfo {
	return (&httpClient{
		method:    http.MethodGet,
		url:       url,
		client:    &client,
		transport: &transport,
		dialer:    &dialer,
	}).Get(url)
}

func (h *httpClient) Post(url string) *httpInfo {
	h.method = http.MethodPost
	h.url = url
	return &httpInfo{handler: h}
}

func (h *httpClient) Get(url string) *httpInfo {
	h.method = http.MethodGet
	h.url = url
	return &httpInfo{handler: h}
}

func (h *httpInfo) Progress(progress *progress) *httpInfo {
	h.Timeout(0)
	h.handler.progress = progress
	return h
}

func (h *httpInfo) Timeout(timeout time.Duration) *httpInfo {
	h.handler.client.Timeout = timeout
	return h
}

func (h *httpInfo) Proxy(url string) *httpInfo {
	var fixUrl, _ = url2.Parse(url)
	h.handler.transport.Proxy = http.ProxyURL(fixUrl)
	return h
}

func (h *httpInfo) KeepAlive(keepalive time.Duration) *httpInfo {
	h.handler.dialer.KeepAlive = keepalive
	return h
}

func (h *httpInfo) SetHeaders(headers map[string]string) *httpInfo {
	for key, value := range headers {
		h.handler.headerKey = append(h.handler.headerKey, key)
		h.handler.headerValue = append(h.handler.headerValue, value)
	}
	return h
}

func (h *httpInfo) AddHeader(key string, value string) *httpInfo {
	h.handler.headerKey = append(h.handler.headerKey, key)
	h.handler.headerValue = append(h.handler.headerValue, value)
	return h
}

func (h *httpInfo) SetHeader(key string, value string) *httpInfo {
	for i := 0; i < len(h.handler.headerKey); i++ {
		if h.handler.headerKey[i] == key {
			h.handler.headerValue[i] = value
			return h
		}
	}

	h.handler.headerKey = append(h.handler.headerKey, key)
	h.handler.headerValue = append(h.handler.headerValue, value)
	return h
}

func (h *httpInfo) SetCookies(cookies []*http.Cookie) *httpInfo {
	h.handler.cookies = cookies
	return h
}

func (h *httpInfo) AddCookie(cookie *http.Cookie) *httpInfo {
	h.handler.cookies = append(h.handler.cookies, cookie)
	return h
}

func (h *httpInfo) Json(body interface{}) *httpInfo {
	h.SetHeader("Content-Type", applicationJson)
	h.handler.body = body
	return h
}

func (h *httpInfo) Query(body map[string]interface{}) *httpInfo {
	h.handler.body = body
	return h
}

func (h *httpInfo) Form(body map[string]interface{}) *httpInfo {
	h.SetHeader("Content-Type", applicationFormUrlencoded)
	h.handler.body = body
	return h
}

func (h *httpInfo) Multipart(body map[string]interface{}) *httpInfo {
	h.SetHeader("Content-Type", multipartFormData)
	h.handler.body = body
	return h
}

func (h *httpInfo) Send() *Request {
	h.handler.transport.DialContext = h.handler.dialer.DialContext
	h.handler.client.Transport = h.handler.transport
	return do(h.handler)
}
