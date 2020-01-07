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

var dial = net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var transport = http.Transport{
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 15 * time.Second,
	ExpectContinueTimeout: 2 * time.Second,
}

var client = http.Client{
	Timeout: 15 * time.Second,
}

func do(client *http.Client, method string, url string, headerKey []string, headerValue []string, body interface{}, cookies []*http.Cookie) *Request {

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
					buff.WriteString(key + "=" + fmt.Sprintf("%v", value))
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

	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		_ = response.Body.Close()
		return &Request{err: err}
	}
	_ = response.Body.Close()

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
	dial        *net.Dialer
}

type httpRequest struct {
	h *httpClient
}

type requestType struct {
	h *httpClient
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
		transport: &http.Transport{},
		dial:      &net.Dialer{},
	}
}

func (h hc) Post(url string) *httpRequest {
	return (&httpClient{
		method:    http.MethodPost,
		url:       url,
		client:    &client,
		transport: &transport,
		dial:      &dial,
	}).Post(url)
}

func (h hc) Get(url string) *httpRequest {
	return (&httpClient{
		method:    http.MethodGet,
		url:       url,
		client:    &client,
		transport: &transport,
		dial:      &dial,
	}).Get(url)
}

func (h *httpClient) Post(url string) *httpRequest {
	h.method = http.MethodPost
	h.url = url
	return &httpRequest{h: h}
}

func (h *httpClient) Get(url string) *httpRequest {
	h.method = http.MethodGet
	h.url = url
	return &httpRequest{h: h}
}

func (h *httpRequest) Timeout(timeout time.Duration) *httpRequest {
	h.h.client.Timeout = timeout
	return h
}

func (h *httpRequest) Proxy(url string) *httpRequest {
	var fixUrl, _ = url2.Parse(url)
	h.h.transport.Proxy = http.ProxyURL(fixUrl)
	return h
}

func (h *httpRequest) KeepAlive(timeout time.Duration) *httpRequest {
	h.h.dial.KeepAlive = timeout
	return h
}

func (h *httpRequest) SetHeaders(headers map[string]string) *httpRequest {
	for key, value := range headers {
		h.h.headerKey = append(h.h.headerKey, key)
		h.h.headerValue = append(h.h.headerValue, value)
	}
	return h
}

func (h *httpRequest) AddHeader(key string, value string) *httpRequest {
	h.h.headerKey = append(h.h.headerKey, key)
	h.h.headerValue = append(h.h.headerValue, value)
	return h
}

func (h *httpRequest) SetHeader(key string, value string) *httpRequest {
	for i := 0; i < len(h.h.headerKey); i++ {
		if h.h.headerKey[i] == key {
			h.h.headerValue[i] = value
			return h
		}
	}

	h.h.headerKey = append(h.h.headerKey, key)
	h.h.headerValue = append(h.h.headerValue, value)
	return h
}

func (h *httpRequest) SetCookies(cookies []*http.Cookie) *httpRequest {
	h.h.cookies = cookies
	return h
}

func (h *httpRequest) AddCookie(cookie *http.Cookie) *httpRequest {
	h.h.cookies = append(h.h.cookies, cookie)
	return h
}

func (h *httpRequest) Json(body interface{}) *requestType {
	h.SetHeader("Content-Type", applicationJson)
	h.h.body = body
	return &requestType{h: h.h}
}

func (h *httpRequest) Query(body map[string]interface{}) *requestType {
	h.h.body = body
	return &requestType{h: h.h}
}

func (h *httpRequest) Form(body map[string]interface{}) *requestType {
	h.SetHeader("Content-Type", applicationFormUrlencoded)
	h.h.body = body
	return &requestType{h: h.h}
}

func (h *httpRequest) Multipart(body map[string]interface{}) *requestType {
	h.SetHeader("Content-Type", multipartFormData)
	h.h.body = body
	return &requestType{h: h.h}
}

func (h *requestType) Send() *Request {
	h.h.transport.DialContext = h.h.dial.DialContext
	h.h.client.Transport = h.h.transport
	return do(h.h.client, h.h.method, h.h.url, h.h.headerKey, h.h.headerValue, h.h.body, h.h.cookies)
}
