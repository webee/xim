// Package httpclient
// Copyright 2015 mint.zhao.chiu@gmail.com
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package httpclient

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	transport = &Transport{
		ConnectTimeout:        2 * time.Second,
		RequestTimeout:        10 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	client = &http.Client{Transport: transport}
)

// ForwardHTTP 转发http请求
func ForwardHTTP(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// GetForwardHTTPBody 获取http response body
func GetForwardHTTPBody(body io.ReadCloser) []byte {
	bodyBytes, _ := ioutil.ReadAll(body)
	defer body.Close()

	return bodyBytes
}

// BindURLParams 绑定url & params
func BindURLParams(url string, params url.Values) string {
	if params == nil || len(params) == 0 {
		return url
	}

	return url + "?" + params.Encode()
}
