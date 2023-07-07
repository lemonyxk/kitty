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
	"strconv"
	"strings"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/kitty/header"
	"google.golang.org/protobuf/proto"
)

//func getRequest(method string, url string, info *Request) (*http.Request, context.CancelFunc, error) {
//	var contentType = strings.ToLower(getContentType(info))
//	switch contentType {
//	case header.ApplicationFormUrlencoded:
//		return doFormUrlencoded(method, url, info)
//	case header.ApplicationJson:
//		return doJson(method, url, info)
//	case header.MultipartFormData:
//		return doFormData(method, url, info)
//	case header.ApplicationProtobuf:
//		return doXProtobuf(method, url, info)
//	default:
//		return doUrl(method, url, info)
//	}
//}

func doRaw(method string, url string, info *Request, body io.Reader) (*http.Request, context.CancelFunc, error) {

	if body == nil {
		body = bytes.NewReader([]byte{})
	}

	//// check body type generic is io.Reader or not
	//body, ok := any(info.body).(io.Reader)
	//if !ok {
	//	return nil, nil, errors.Wrap(errors.AssertionFailed, "io.Reader")
	//}

	out, in := io.Pipe()
	var ctx, cancel = context.WithTimeout(context.Background(), info.clientTimeout)
	pCtx, pCancel := context.WithCancel(ctx)
	go func() {
		defer func() {
			// if close in first then the part will close too before read,
			// cuz when in close the part will be clean up,
			// then you can not read the part.
			// NOTICE: we need to close the part to write the last boundary
			// or the message will be broken.
			_ = in.Close()
			pCancel()
		}()

		if _, err := io.Copy(in, body); err != nil {
			return
		}
	}()

	go func() {
		for {
			select {
			case <-pCtx.Done():
				_ = in.Close()
				return
			}
		}
	}()

	request, err := http.NewRequestWithContext(ctx, method, url, out)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doProtobuf(method string, url string, info *Request, body ...proto.Message) (*http.Request, context.CancelFunc, error) {

	var protobufBody []byte

	for i := 0; i < len(body); i++ {
		b, err := proto.Marshal(body[i])
		if err != nil {
			return nil, nil, err
		}
		protobufBody = append(protobufBody, b...)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), info.clientTimeout)
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(protobufBody))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

type file interface {
	io.Reader
	Name() string
}

func doFormData(method string, url string, info *Request, body ...kitty.M) (*http.Request, context.CancelFunc, error) {
	out, in := io.Pipe()
	part := multipart.NewWriter(in)
	var ctx, cancel = context.WithTimeout(context.Background(), info.clientTimeout)
	pCtx, pCancel := context.WithCancel(ctx)
	go func() {
		defer func() {
			// if close in first then the part will close too before read,
			// cuz when in close the part will be clean up,
			// then you can not read the part.
			// NOTICE: we need to close the part to write the last boundary
			// or the message will be broken.
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
				case file:
					w, err := part.CreateFormFile(key, value.(file).Name())
					if err != nil {
						return
					}
					if _, err = io.Copy(w, value.(file)); err != nil {
						return
					}
				default:
					w, err := part.CreateFormField(key)
					if err != nil {
						return
					}
					str := fmt.Sprintf("%+v", value)
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
	info.SetHeader(header.ContentType, part.FormDataContentType())
	return request, cancel, err
}

func doJson(method string, url string, info *Request, body ...any) (*http.Request, context.CancelFunc, error) {

	var jsonBody []byte

	for i := 0; i < len(body); i++ {
		b, err := jsoniter.Marshal(body[i])
		if err != nil {
			return nil, nil, err
		}
		jsonBody = append(jsonBody, b...)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), info.clientTimeout)
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonBody))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doFormUrlencoded(method string, url string, info *Request, body ...kitty.M) (*http.Request, context.CancelFunc, error) {
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
				buff.WriteString(key + "=" + fmt.Sprintf("%+v", value) + "&")
			}
		}
	}

	var b = buff.Bytes()
	if len(b) != 0 {
		b = b[:len(b)-1]
	}

	var ctx, cancel = context.WithTimeout(context.Background(), info.clientTimeout)
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func doUrl(method string, u string, info *Request, body ...kitty.M) (*http.Request, context.CancelFunc, error) {
	Url, err := url.Parse(u)
	if err != nil {
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
				params.Set(key, fmt.Sprintf("%+v", value))
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

	var ctx, cancel = context.WithTimeout(context.Background(), info.clientTimeout)
	request, err := http.NewRequestWithContext(ctx, method, Url.String(), nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return request, cancel, err
}

func getContentType(info *Request) string {
	for i := 0; i < len(info.headerKey); i++ {
		if info.headerKey[i] == header.ContentType {
			return info.headerValue[i]
		}
	}
	return ""
}

func send(info *Request, req *http.Request, cancel context.CancelFunc) *Response {

	defer cancel()

	// // NOT SAFE FOR GOROUTINE IF YOU SET TIMEOUT OR KEEPALIVE OR PROXY OR PROGRESS
	// // MAKE SURE ONE BY ONE
	// defer func() {
	// 	defaultClient.Timeout = clientTimeout
	// 	defaultDialer.KeepAlive = dialerKeepAlive
	// 	defaultTransport.Proxy = http.ProxyFromEnvironment
	// }()

	if req == nil {
		return &Response{err: errors.Invalid}
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

	response, err := defaultClient.Do(req)
	if err != nil {
		return &Response{err: err}
	}
	defer func() { _ = response.Body.Close() }()

	var buf = new(bytes.Buffer)

	if info.progress != nil {
		var total, _ = strconv.ParseInt(response.Header.Get(header.ContentLength), 10, 64)
		var writer = &writer{
			total:      total,
			onProgress: info.progress.progress,
			rate:       info.progress.rate,
		}

		_, err = io.Copy(buf, io.TeeReader(response.Body, writer))
		if err != nil {
			return &Response{err: err}
		}
	} else {
		_, err = io.Copy(buf, response.Body)
		if err != nil {
			return &Response{err: err}
		}
	}

	return &Response{code: response.StatusCode, buf: buf, req: response}
}

func do(info *Request, req *http.Request, cancel context.CancelFunc) (*http.Response, error) {

	// // NOT SAFE FOR GOROUTINE IF YOU SET TIMEOUT OR KEEPALIVE OR PROXY OR PROGRESS
	// // MAKE SURE ONE BY ONE
	// defer func() {
	// 	defaultClient.Timeout = clientTimeout
	// 	defaultDialer.KeepAlive = dialerKeepAlive
	// 	defaultTransport.Proxy = http.ProxyFromEnvironment
	// }()

	if req == nil {
		cancel()
		return nil, errors.Invalid
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

	res, err := defaultClient.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	return res, nil
}
