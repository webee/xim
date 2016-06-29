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

import "fmt"

var (
	apiDomain  = "openapi.xg.qq.com"
	apiVersion = "v2"
)

var (
	singleDeviceURL  = fmt.Sprintf("http://%s/%s/push/single_device", apiDomain, apiVersion)
	singleAccountURL = fmt.Sprintf("http://%s/%s/push/single_account", apiDomain, apiVersion)
	multiAccountURL  = fmt.Sprintf("http://%s/%s/push/account_list", apiDomain, apiVersion)
	allDeviceURL     = fmt.Sprintf("http://%s/%s/push/all_device", apiDomain, apiVersion)
	tagsDeviceURL    = fmt.Sprintf("http://%s/%s/push/tags_device", apiDomain, apiVersion)
	deviceNumURL     = fmt.Sprintf("http://%s/%s/application/get_app_device_num", apiDomain, apiVersion)
	appTagsURL       = fmt.Sprintf("http://%s/%s/tags/query_app_tags", apiDomain, apiVersion)
	batchSetTagsURL  = fmt.Sprintf("http://%s/%s/tags/batch_set", apiDomain, apiVersion)
	batchDelTagsURL  = fmt.Sprintf("http://%s/%s/tags/batch_del", apiDomain, apiVersion)
	tokenTagsURL     = fmt.Sprintf("http://%s/%s/tags/query_token_tags", apiDomain, apiVersion)
	tagTokensURL     = fmt.Sprintf("http://%s/%s/tags/query_tag_token_num", apiDomain, apiVersion)
)

// PlatformType the byte type.
type PlatformType byte

// PlatformType ios 0 android 1
const (
	PlatformIos PlatformType = iota
	PlatformAndroid
)

// MessageType the byte type.
type MessageType byte

// MessageType ios 0 android notify 1 android pushthrough 2
const (
	MessageTypeIos MessageType = iota
	MessageTypeNotify
	MessageTypePassthrough
)

// PushEnv the byte type.
type PushEnv byte

// push evn android 0 prod 1 dev 2
const (
	PushEnvAndroid PushEnv = iota
	PushEnvProd
	PushEnvDev
)

// MultiPkgType the byte type.
type MultiPkgType byte

// MultiPkgType android 1 ios 2.
const (
	MultiPkgPkg MultiPkgType = iota
	MultiPkgAid
	MultiPkgIos
)

// PushType the byte type.
type PushType byte

// PushType single_device 0 single_account 1 multi_account 2 all_device 3 tags_device 4
const (
	PushTypeSingleDevice PushType = iota
	PushTypeSingleAccount
	PushTypeMultiAccount
	PushTypeAllDevice
	PushTypeTagsDevice
)

// and or
const (
	TagsOpAND = "AND"
	TagsOpOR  = "OR"
)
