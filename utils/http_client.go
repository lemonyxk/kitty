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
	"io/ioutil"
	"net"
	"net/http"
	url2 "net/url"
	"strconv"
	"time"

	"github.com/json-iterator/go"
)

type hc int

const HttpClient hc = iota

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

func do(client *http.Client, method string, url string, headerKey []string, headerValue []string, body interface{}, cookies []*http.Cookie) ([]byte, error) {

	var request *http.Request
	var response *http.Response
	var err error

	if method == http.MethodGet {

		var params = url2.Values{}

		Url, err := url2.Parse(url)
		if err != nil {
			return nil, err
		}

		if _, ok := body.(map[string]interface{}); !ok {
			return nil, errors.New("get method body must be map[string]interface")
		}

		for key, value := range body.(map[string]interface{}) {
			switch value.(type) {
			case string:
				params.Set(key, value.(string))
			case int:
				params.Set(key, strconv.Itoa(value.(int)))
			case float64:
				params.Set(key, strconv.FormatFloat(value.(float64), 'f', -1, 64))
			default:
				return nil, errors.New("invalid query")
			}
		}

		Url.RawQuery = params.Encode()

		request, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}

	} else {

		var contentType = ""

		for i := 0; i < len(headerKey); i++ {
			if headerKey[i] == "Content-Type" {
				contentType = headerValue[i]
				break
			}
		}

		if contentType == "" {
			return nil, errors.New("invalid content-type")
		}

		switch contentType {
		case "application/x-www-form-urlencoded":

			if _, ok := body.(map[string]interface{}); !ok {
				return nil, errors.New("application/x-www-form-urlencoded body must be map[string]interface")
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
					return nil, errors.New("invalid params")
				}
			}
			var b = buff.Bytes()
			request, err = http.NewRequest(method, url, bytes.NewReader(b[:len(b)-1]))
			if err != nil {
				return nil, err
			}

		case "application/json":

			jsonBody, err := jsoniter.Marshal(body)
			if err != nil {
				return nil, err
			}

			request, err = http.NewRequest(method, url, bytes.NewReader(jsonBody))
			if err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("invalid header")
		}

	}

	if request == nil {
		return nil, errors.New("invalid request")
	}

	for i := 0; i < len(headerKey); i++ {
		request.Header.Add(headerKey[i], headerValue[i])
	}

	for i := 0; i < len(cookies); i++ {
		request.AddCookie(cookies[i])
	}

	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}

	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		_ = response.Body.Close()
		return nil, err
	}
	_ = response.Body.Close()
	return dataBytes, nil
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

func (h hc) New() *httpClient {
	return &httpClient{
		client:    &http.Client{},
		transport: &http.Transport{},
		dial:      &net.Dialer{},
	}
}

func (h hc) Post(url string) *httpClient {
	return &httpClient{
		method:    http.MethodPost,
		url:       url,
		client:    &client,
		transport: &transport,
		dial:      &dial,
	}
}

func (h hc) Get(url string) *httpClient {
	return &httpClient{
		method:    http.MethodGet,
		url:       url,
		client:    &client,
		transport: &transport,
		dial:      &dial,
	}
}

func (h *httpClient) Post(url string) *httpClient {
	h.method = http.MethodPost
	h.url = url
	return h
}

func (h *httpClient) Get(url string) *httpClient {
	h.method = http.MethodGet
	h.url = url
	return h
}

func (h *httpClient) Timeout(timeout time.Duration) *httpClient {
	h.client.Timeout = timeout
	return h
}

func (h *httpClient) Proxy(url string) *httpClient {
	var fixUrl, _ = url2.Parse(url)
	h.transport.Proxy = http.ProxyURL(fixUrl)
	return h
}

func (h *httpClient) KeepAlive(timeout time.Duration) *httpClient {
	h.dial.KeepAlive = timeout
	return h
}

func (h *httpClient) SetHeaders(headers map[string]string) *httpClient {
	for key, value := range headers {
		h.headerKey = append(h.headerKey, key)
		h.headerValue = append(h.headerValue, value)
	}
	return h
}

func (h *httpClient) AddHeader(key string, value string) *httpClient {
	h.headerKey = append(h.headerKey, key)
	h.headerValue = append(h.headerValue, value)
	return h
}

func (h *httpClient) SetHeader(key string, value string) *httpClient {
	for i := 0; i < len(h.headerKey); i++ {
		if h.headerKey[i] == key {
			h.headerValue[i] = value
			return h
		}
	}

	h.headerKey = append(h.headerKey, key)
	h.headerValue = append(h.headerValue, value)
	return h
}

func (h *httpClient) SetCookies(cookies []*http.Cookie) *httpClient {
	h.cookies = cookies
	return h
}

func (h *httpClient) AddCookie(cookie *http.Cookie) *httpClient {
	h.cookies = append(h.cookies, cookie)
	return h
}

func (h *httpClient) Body(body interface{}) *httpClient {
	h.SetHeader("Content-Type", "application/json")
	h.body = body
	return h
}

func (h *httpClient) Form(body map[string]interface{}) *httpClient {
	h.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	h.body = body
	return h
}

func (h *httpClient) Send() ([]byte, error) {
	h.transport.DialContext = h.dial.DialContext
	h.client.Transport = h.transport
	return do(h.client, h.method, h.url, h.headerKey, h.headerValue, h.body, h.cookies)
}
