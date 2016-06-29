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

// AppTags app tags.
func (cli *Client) AppTags(start, limit int) (int64, []string, error) {
	var tags []string

	request := cli.NewRequest("GET", appTagsURL)

	request.SetParam("start", start)
	request.SetParam("limit", limit)

	response, err := request.Execute()
	if err != nil {
		return 0, tags, errors.New("<xinge> request err:" + err.Error())
	}

	if !response.OK() {
		return 0, tags, errors.New("<xinge> response err:" + response.Error())
	}

	result := response.Result.(map[string]interface{})
	total := int64(result["total"].(float64))
	if total <= 0 {
		return 0, tags, nil
	}

	tagList := result["tags"].([]interface{})
	for _, tag := range tagList {
		tags = append(tags, tag.(string))
	}

	return total, tags, nil
}

// SetTags set tags.
func (cli *Client) SetTags(tagTokenList ...[2]string) error {
	request := cli.NewRequest("GET", batchSetTagsURL)

	tagTokens := "["
	for _, tagToken := range tagTokenList {
		pair := `["` + tagToken[0] + `","` + tagToken[1] + `"]` + ","
		tagTokens += pair
	}
	tagTokens = tagTokens[:len(tagTokens)-1]
	tagTokens += "]"

	request.SetParam("tag_token_list", tagTokens)
	response, err := request.Execute()
	if err != nil {
		return errors.New("<xinge> request err:" + err.Error())
	}

	if !response.OK() {
		return errors.New("<xinge> response err:" + response.Error())
	}

	return nil
}

// DelTags delete tags.
func (cli *Client) DelTags(tagTokenList ...[2]string) error {
	request := cli.NewRequest("GET", batchDelTagsURL)

	tagTokens := "["
	for _, tagToken := range tagTokenList {
		pair := `["` + tagToken[0] + `","` + tagToken[1] + `"]` + ","
		tagTokens += pair
	}
	tagTokens = tagTokens[:len(tagTokens)-1]
	tagTokens += "]"

	request.SetParam("tag_token_list", tagTokens)
	response, err := request.Execute()
	if err != nil {
		return errors.New("<xinge> request err:" + err.Error())
	}

	if !response.OK() {
		return errors.New("<xinge> response err:" + response.Error())
	}

	return nil
}

// TokenTags get token tags.
func (cli *Client) TokenTags(token string) ([]string, error) {
	var tags []string
	request := cli.NewRequest("GET", tokenTagsURL)

	request.SetParam("device_token", token)
	response, err := request.Execute()
	if err != nil {
		return tags, errors.New("<xinge> request err:" + err.Error())
	}

	if !response.OK() {
		return tags, errors.New("<xinge> response err:" + response.Error())
	}

	result := response.Result.(map[string]interface{})
	tagList := result["tags"].([]interface{})
	for _, tag := range tagList {
		tags = append(tags, tag.(string))
	}

	return tags, nil
}

// TagTokensNum get tag token num.
func (cli *Client) TagTokensNum(tag string) (int64, error) {
	request := cli.NewRequest("GET", tagTokensURL)

	request.SetParam("tag", tag)
	response, err := request.Execute()
	if err != nil {
		return 0, errors.New("<xinge> request err:" + err.Error())
	}

	if !response.OK() {
		return 0, errors.New("<xinge> response err:" + response.Error())
	}

	result := response.Result.(map[string]interface{})
	return int64(result["device_num"].(float64)), nil
}
