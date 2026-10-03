// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package google_groups_crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	// rpc id that lists the conversations of a group
	rpcListConversations = "Dq0xse"
	// rpc id that lists the messages of a conversation
	rpcListMessages = "H08Fi"

	batchExecuteUrl = "https://groups.google.com/_/GroupsFrontendUi/data/batchexecute"
)

var xsrfTokenRegexp = regexp.MustCompile(`"SNlM0e":"([^"]+)"`)

func doRequest(client http.Client, req *http.Request) (string, bool) {
	res, err := client.Do(req)
	if err != nil || res == nil {
		if err != nil {
			fmt.Printf("Google Groups Crawler: http request error: %s\n", err.Error())
		}
		return "", false
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.Body == nil {
		return "", false
	}

	if res.StatusCode != 200 {
		fmt.Printf("Google Groups Crawler: http %s request status code: %d\n", req.Method, res.StatusCode)
		return "", false
	}
	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return "", false
	}
	return string(resp), true
}

func getPage(client http.Client, targetUrl string, cookie string) (string, bool) {
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return "", false
	}
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}
	return doRequest(client, req)
}

// getInitData returns the data embedded in the page for the given rpc id, e.g.
// AF_initDataCallback({key: 'ds:7', hash: '9', data:[...], sideChannel: {}});
// The page declares which ds key belongs to which rpc id in AF_dataServiceRequests.
func getInitData(body string, rpcId string) ([]interface{}, bool) {
	start := -1
	keyRegexp := regexp.MustCompile(`'(ds:\d+)' : \{id:'` + regexp.QuoteMeta(rpcId) + `'`)
	if match := keyRegexp.FindStringSubmatch(body); match != nil {
		start = strings.Index(body, "AF_initDataCallback({key: '"+match[1]+"'")
	}
	if start < 0 {
		// fall back to the last data block of the page
		start = strings.LastIndex(body, "AF_initDataCallback({key: 'ds:")
	}
	if start < 0 {
		return nil, false
	}
	body = body[start:]

	end := strings.Index(body, ", sideChannel: {}});")
	if end < 0 {
		return nil, false
	}
	body = body[:end]
	start = strings.Index(body, "data:")
	if start < 0 {
		return nil, false
	}
	body = body[start+5:]

	var dataArray []interface{}
	err := json.Unmarshal([]byte(body), &dataArray)
	if err != nil {
		return nil, false
	}
	return dataArray, true
}

func getXsrfToken(body string) string {
	match := xsrfTokenRegexp.FindStringSubmatch(body)
	if match == nil {
		return ""
	}
	return match[1]
}

// batchExecute calls the same rpc the Google Groups web page uses, e.g. when
// loading the next page of conversations.
func batchExecute(client http.Client, cookie string, xsrfToken string, rpcId string, args []interface{}) ([]interface{}, bool) {
	argsJson, err := json.Marshal(args)
	if err != nil {
		return nil, false
	}
	fReq, err := json.Marshal([]interface{}{[]interface{}{[]interface{}{rpcId, string(argsJson), nil, "generic"}}})
	if err != nil {
		return nil, false
	}
	form := url.Values{}
	form.Set("f.req", string(fReq))
	if xsrfToken != "" {
		form.Set("at", xsrfToken)
	}

	req, err := http.NewRequest("POST", batchExecuteUrl+"?rpcids="+rpcId, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}
	body, ok := doRequest(client, req)
	if !ok {
		return nil, false
	}

	// the response starts with the anti-XSSI prefix )]}'
	start := strings.Index(body, "[")
	if start < 0 {
		return nil, false
	}
	var entries []interface{}
	err = json.Unmarshal([]byte(body[start:]), &entries)
	if err != nil {
		return nil, false
	}
	for _, e := range entries {
		entry, ok := e.([]interface{})
		if !ok || len(entry) < 3 || entry[0] != "wrb.fr" || entry[1] != rpcId {
			continue
		}
		data, ok := entry[2].(string)
		if !ok {
			fmt.Printf("Google Groups Crawler: rpc %s returned an error\n", rpcId)
			return nil, false
		}
		var dataArray []interface{}
		err = json.Unmarshal([]byte(data), &dataArray)
		if err != nil {
			return nil, false
		}
		return dataArray, true
	}
	return nil, false
}

func getGroupEmail(dataArray []interface{}) string {
	if len(dataArray) < 1 {
		return ""
	}
	groupArray, ok := dataArray[0].([]interface{})
	if !ok || len(groupArray) < 2 {
		return ""
	}
	email, _ := groupArray[1].(string)
	return email
}
