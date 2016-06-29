// Package xinge
// Copyright 2015 mint.zhao.chiu@gmail.com
//
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

package xinge

import "errors"

// Client xinge client struct
type Client struct {
	AccessID  string
	AccessKey string
	ValidTime uint
	SecretKey string
}

// NewClient new xinge client.
func NewClient(accessID string, validTime uint, accessKey, secretKey string) *Client {
	return &Client{
		AccessID:  accessID,
		AccessKey: accessKey,
		ValidTime: validTime,
		SecretKey: secretKey,
	}
}

// NewRequest new xinge request.
func (cli *Client) NewRequest(method, url string) *Request {
	return &Request{
		HTTPMethod: method,
		HTTPURL:    url,
		Params:     make(map[string]interface{}),
		Client:     cli,
	}
}

// AppDeviceNum get app device num.
func (cli *Client) AppDeviceNum() (int64, error) {
	request := cli.NewRequest("GET", deviceNumURL)

	response, err := request.Execute()
	if err != nil {
		return 0, errors.New("<xinge> request app device num err:" + err.Error())
	}

	if !response.OK() {
		return 0, errors.New("<xinge> response err:" + response.Error())
	}

	result := response.Result.(map[string]interface{})

	return int64(result["device_num"].(float64)), nil
}
