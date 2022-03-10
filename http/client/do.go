/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-21 17:35
**/

package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	url2 "net/url"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemoyxk/kitty/v2/kitty"
)

func getRequest(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	var contentType = strings.ToLower(getContentType(info))
	switch contentType {
	case kitty.ApplicationFormUrlencoded:
		return doPostFormUrlencoded(method, url, info)
	case kitty.ApplicationJson:
		return doPostJson(method, url, info)
	case kitty.MultipartFormData:
		return doPostFormData(method, url, info)
	case kitty.ApplicationProtobuf:
		return doPostXProtobuf(method, url, info)
	default:
		return doUrl(method, url, info)
	}
}

func doPostXProtobuf(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	body, ok := info.body.([]proto.Message)
	if !ok {
		return nil, nil, errors.New(kitty.ApplicationProtobuf + " body must be proto.Message")
	}

	var protobufBody []byte

	for i := 0; i < len(body); i++ {
		b, err := proto.Marshal(body[i])
		if err != nil {
			return nil, nil, err
		}
		protobufBody = append(protobufBody, b...)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(protobufBody))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doPostFormData(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	if info.body == nil {
		info.body = []map[string]interface{}{}
	}

	body, ok := info.body.([]map[string]interface{})
	if !ok {
		return nil, nil, errors.New(kitty.MultipartFormData + " body must be map[string]interface")
	}

	var buf = new(bytes.Buffer)
	part := multipart.NewWriter(buf)

	for i := 0; i < len(body); i++ {
		for key, value := range body[i] {
			switch value.(type) {
			case string:
				if err := part.WriteField(key, value.(string)); err != nil {
					return nil, nil, err
				}
			case int:
				if err := part.WriteField(key, strconv.Itoa(value.(int))); err != nil {
					return nil, nil, err
				}
			case float64:
				if err := part.WriteField(key, strconv.FormatFloat(value.(float64), 'f', -1, 64)); err != nil {
					return nil, nil, err
				}
			case *os.File:
				ff, err := part.CreateFormFile(key, value.(*os.File).Name())
				if err != nil {
					return nil, nil, err
				}
				_, err = io.Copy(ff, value.(*os.File))
				if err != nil {
					return nil, nil, err
				}
			default:
				if err := part.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
					return nil, nil, err
				}
			}
		}
	}

	if err := part.Close(); err != nil {
		return nil, nil, err
	}

	var ctx, cancel = context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	info.SetHeader(kitty.ContentType, part.FormDataContentType())
	return request, cancel, err
}

func doPostJson(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	body, ok := info.body.([]interface{})
	if !ok {
		return nil, nil, errors.New(kitty.ApplicationJson + " body must be interface")
	}

	var jsonBody []byte

	for i := 0; i < len(body); i++ {
		b, err := jsoniter.Marshal(body[i])
		if err != nil {
			return nil, nil, err
		}
		jsonBody = append(jsonBody, b...)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonBody))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doPostFormUrlencoded(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	if info.body == nil {
		info.body = []map[string]interface{}{}
	}

	body, ok := info.body.([]map[string]interface{})
	if !ok {
		return nil, nil, errors.New(kitty.ApplicationFormUrlencoded + " body must be map[string]interface")
	}

	var buff bytes.Buffer
	for i := 0; i < len(body); i++ {
		for key, value := range body[i] {
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
	}

	var b = buff.Bytes()
	if len(b) != 0 {
		b = b[:len(b)-1]
	}

	var ctx, cancel = context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doUrl(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	Url, err := url2.Parse(url)
	if err != nil {
		return nil, nil, err
	}

	if info.body == nil {
		info.body = []map[string]interface{}{}
	}

	body, ok := info.body.([]map[string]interface{})
	if !ok {
		return nil, nil, err
	}

	var params = url2.Values{}

	for i := 0; i < len(body); i++ {
		for key, value := range body[i] {
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
	}

	var pStr = params.Encode()

	if pStr != "" {
		Url.RawQuery = Url.RawQuery + "&" + pStr
	}

	var ctx, cancel = context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, method, Url.String(), nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func getContentType(info *info) string {
	for i := 0; i < len(info.headerKey); i++ {
		if info.headerKey[i] == kitty.ContentType {
			return info.headerValue[i]
		}
	}
	return ""
}

func send(info *info, req *http.Request, cancel context.CancelFunc) *Req {

	defer cancel()

	if req == nil {
		return &Req{err: errors.New("invalid request")}
	}

	for i := 0; i < len(info.headerKey); i++ {
		req.Header.Add(info.headerKey[i], info.headerValue[i])
	}

	for i := 0; i < len(info.cookies); i++ {
		req.AddCookie(info.cookies[i])
	}

	if info.userName != "" || info.passWord != "" {
		req.SetBasicAuth(info.userName, info.passWord)
	}

	if info.clientTimeout != 0 {
		defaultClient.Timeout = info.clientTimeout
	}

	if info.dialerKeepAlive != 0 {
		defaultDialer.KeepAlive = info.dialerKeepAlive
	}

	if info.proxy != nil {
		defaultTransport.Proxy = info.proxy
	}

	if info.progress != nil {
		defaultTransport.DisableCompression = true
	}

	// NOT SAFE FOR GOROUTINE IF YOU SET TIMEOUT OR KEEPALIVE OR PROXY OR PROGRESS
	// MAKE SURE ONE BY ONE
	defer func() {
		defaultClient.Timeout = clientTimeout
		defaultDialer.KeepAlive = dialerKeepAlive
		defaultTransport.Proxy = http.ProxyFromEnvironment
		defaultTransport.DisableCompression = false
	}()

	response, err := defaultClient.Do(req)
	if err != nil {
		return &Req{err: err}
	}
	defer func() { _ = response.Body.Close() }()

	var dataBytes []byte

	if info.progress != nil {
		var total, _ = strconv.ParseInt(response.Header.Get(kitty.ContentLength), 10, 64)
		var writer = &writer{
			total:      total,
			onProgress: info.progress.progress,
			rate:       info.progress.rate,
		}

		dataBytes, err = ioutil.ReadAll(io.TeeReader(response.Body, writer))
		if err != nil {
			return &Req{err: err}
		}
	} else {
		dataBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return &Req{err: err}
		}
	}

	return &Req{code: response.StatusCode, data: dataBytes, req: response}
}
