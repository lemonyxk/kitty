/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-21 17:35
**/

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/kitty"
)

func getRequest(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	var contentType = strings.ToLower(getContentType(info))
	switch contentType {
	case kitty.ApplicationFormUrlencoded:
		return doFormUrlencoded(method, url, info)
	case kitty.ApplicationJson:
		return doJson(method, url, info)
	case kitty.MultipartFormData:
		return doFormData(method, url, info)
	case kitty.ApplicationProtobuf:
		return doXProtobuf(method, url, info)
	default:
		return doUrl(method, url, info)
	}
}

func doRaw(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	body, ok := info.body.([][]byte)
	if !ok {
		return nil, nil, errors.Wrap(errors.AssertionFailed, "[]byte")
	}

	var rawBody []byte

	for i := 0; i < len(body); i++ {
		rawBody = append(rawBody, body[i]...)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(rawBody))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doXProtobuf(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	body, ok := info.body.([]proto.Message)
	if !ok {
		return nil, nil, errors.Wrap(errors.AssertionFailed, "proto.Message")
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

func doFormData(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	if info.body == nil {
		info.body = []kitty.M{}
	}

	body, ok := info.body.([]kitty.M)
	if !ok {
		return nil, nil, errors.Wrap(errors.AssertionFailed, "map[string]interface{}")
	}

	out, in := io.Pipe()
	part := multipart.NewWriter(in)
	var ctx, cancel = context.WithCancel(context.Background())
	pCtx, pCancel := context.WithCancel(ctx)
	go func() {
		defer func() {
			// if close in first then the part will close too before read.
			// so you can not read the part.
			_ = part.Close()
			_ = in.Close()
			pCancel()
		}()

		for i := 0; i < len(body); i++ {
			for key, value := range body[i] {
				switch value.(type) {
				case string:
					w, err := part.CreateFormField(key)
					if err != nil {
						return
					}
					if _, err := io.Copy(w, strings.NewReader(value.(string))); err != nil {
						return
					}
				case int:
					w, err := part.CreateFormField(key)
					if err != nil {
						return
					}
					str := strconv.Itoa(value.(int))
					if _, err := io.Copy(w, strings.NewReader(str)); err != nil {
						return
					}
				case float64:
					w, err := part.CreateFormField(key)
					if err != nil {
						return
					}
					str := strconv.FormatFloat(value.(float64), 'f', -1, 64)
					if _, err := io.Copy(w, strings.NewReader(str)); err != nil {
						return
					}
				case *os.File:
					w, err := part.CreateFormFile(key, value.(*os.File).Name())
					if err != nil {
						return
					}
					if _, err = io.Copy(w, value.(*os.File)); err != nil {
						return
					}
				default:
					w, err := part.CreateFormField(key)
					if err != nil {
						return
					}
					str := fmt.Sprintf("%v", value)
					if _, err := io.Copy(w, strings.NewReader(str)); err != nil {
						return
					}
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-pCtx.Done():
				_ = in.Close()
				_ = part.Close()
				return
			}
		}
	}()

	request, err := http.NewRequestWithContext(ctx, method, url, out)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	info.SetHeader(kitty.ContentType, part.FormDataContentType())
	return request, cancel, err
}

func doJson(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	body, ok := info.body.([]any)
	if !ok {
		return nil, nil, errors.Wrap(errors.AssertionFailed, "interface{}")
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

func doFormUrlencoded(method string, url string, info *info) (*http.Request, context.CancelFunc, error) {
	if info.body == nil {
		info.body = []kitty.M{}
	}

	body, ok := info.body.([]kitty.M)
	if !ok {
		return nil, nil, errors.Wrap(errors.AssertionFailed, "map[string]interface{}")
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

func doUrl(method string, u string, info *info) (*http.Request, context.CancelFunc, error) {
	Url, err := url.Parse(u)
	if err != nil {
		return nil, nil, err
	}

	if info.body == nil {
		info.body = []kitty.M{}
	}

	body, ok := info.body.([]kitty.M)
	if !ok {
		return nil, nil, err
	}

	var params = url.Values{}

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
		if Url.RawQuery != "" {
			Url.RawQuery = Url.RawQuery + "&" + pStr
		} else {
			Url.RawQuery = pStr
		}
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

	// NOT SAFE FOR GOROUTINE IF YOU SET TIMEOUT OR KEEPALIVE OR PROXY OR PROGRESS
	// MAKE SURE ONE BY ONE
	defer func() {
		defaultClient.Timeout = clientTimeout
		defaultDialer.KeepAlive = dialerKeepAlive
		defaultTransport.Proxy = http.ProxyFromEnvironment
		defaultTransport.DisableCompression = false
	}()

	if req == nil {
		return &Req{err: errors.Invalid}
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

	response, err := defaultClient.Do(req)
	if err != nil {
		return &Req{err: err}
	}
	defer func() { _ = response.Body.Close() }()

	var buf = new(bytes.Buffer)

	if info.progress != nil {
		var total, _ = strconv.ParseInt(response.Header.Get(kitty.ContentLength), 10, 64)
		var writer = &writer{
			total:      total,
			onProgress: info.progress.progress,
			rate:       info.progress.rate,
		}

		_, err = io.Copy(buf, io.TeeReader(response.Body, writer))
		if err != nil {
			return &Req{err: err}
		}
	} else {
		_, err = io.Copy(buf, response.Body)
		if err != nil {
			return &Req{err: err}
		}
	}

	return &Req{code: response.StatusCode, buf: buf, req: response}
}
