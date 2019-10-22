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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	url2 "net/url"
	"time"

	"github.com/Lemo-yxk/lemo"
)

var httpClient = &http.Client{
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

func Post(url string, values interface{}) ([]byte, error) {

	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Post(url, "application/json", bytes.NewReader(jsonValue))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	return body, nil

}

func Get(url string, values lemo.M) ([]byte, error) {

	var params = url2.Values{}

	Url, err := url2.Parse(url)
	if err != nil {
		return nil, err
	}

	for key, value := range values {
		params.Set(key, fmt.Sprintf("%s", value))
	}

	Url.RawQuery = params.Encode()

	res, err := httpClient.Get(Url.String())
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	return body, nil

}
