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
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Lemo-yxk/lemo"
)

var handler = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			Deadline:  time.Now().Add(3 * time.Second),
			KeepAlive: 15 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 2 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: 5 * time.Second,
}

func do(method string, url string, headerKey []string, headerValue []string, body interface{}, cookies []*http.Cookie) ([]byte, error) {

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

		for key, value := range headerKey {
			if value == "Content-Type" {
				contentType = headerValue[key]
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

			var str = ""
			for key, value := range body.(map[string]interface{}) {
				switch value.(type) {
				case string:
					str += key + "=" + value.(string) + "&"
				case int:
					str += key + "=" + strconv.Itoa(value.(int)) + "&"
				case float64:
					str += key + "=" + strconv.FormatFloat(value.(float64), 'f', -1, 64) + "&"
				default:
					return nil, errors.New("invalid params")
				}
			}

			request, err = http.NewRequest(method, url, strings.NewReader(str[:len(str)-1]))
			if err != nil {
				return nil, err
			}

		case "application/json":

			jsonBody, err := json.Marshal(body)
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

	for key, value := range headerKey {
		request.Header.Add(value, headerValue[key])
	}

	for _, value := range cookies {
		request.AddCookie(value)
	}

	response, err = handler.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() { _ = response.Body.Close() }()

	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return dataBytes, nil
}

type httpClient struct {
	method      string
	url         string
	headerKey   []string
	headerValue []string
	cookies     []*http.Cookie
	body        interface{}
}

func Post(url string) *httpClient {
	return &httpClient{
		method: http.MethodPost,
		url:    url,
	}
}

func Get(url string) *httpClient {
	return &httpClient{
		method: http.MethodGet,
		url:    url,
	}
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
	for k, v := range h.headerKey {
		if v == key {
			h.headerValue[k] = value
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

func (h *httpClient) Form(body lemo.M) *httpClient {
	h.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	h.body = body
	return h
}

func (h *httpClient) Send() ([]byte, error) {
	return do(h.method, h.url, h.headerKey, h.headerValue, h.body, h.cookies)
}
