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
	"fmt"
	"net/http"
)

// conversations per page, the same as the Google Groups web page
const conversationPageSize = 30

func (g GoogleGroup) GetAllConversations(client http.Client) []GoogleGroupConversation {
	targetUrl := fmt.Sprintf("https://groups.google.com/g/%s", g.GroupName)
	var ret []GoogleGroupConversation

	body, ok := getPage(client, targetUrl, g.Cookie)
	if !ok {
		return ret
	}

	// the page only contains the first page of conversations
	dataArray, ok := getInitData(body, rpcListConversations)
	if !ok {
		return ret
	}
	ret = append(ret, g.parseConversations(dataArray)...)

	// load the remaining pages the same way as scrolling down in the web page
	groupEmail := getGroupEmail(dataArray)
	xsrfToken := getXsrfToken(body)
	pageToken := getPageToken(dataArray)
	for groupEmail != "" && pageToken != "" {
		dataArray, ok = batchExecute(client, g.Cookie, xsrfToken, rpcListConversations,
			[]interface{}{groupEmail, conversationPageSize, pageToken, []interface{}{}, 2})
		if !ok {
			break
		}
		ret = append(ret, g.parseConversations(dataArray)...)
		pageToken = getPageToken(dataArray)
	}
	return ret
}

func getPageToken(dataArray []interface{}) string {
	if len(dataArray) < 4 {
		return ""
	}
	pageToken, _ := dataArray[3].(string)
	return pageToken
}

func (g GoogleGroup) parseConversations(dataArray []interface{}) []GoogleGroupConversation {
	var ret []GoogleGroupConversation
	if len(dataArray) < 3 {
		return ret
	}
	conversationArray, ok := dataArray[2].([]interface{})
	if !ok {
		return ret
	}
	for _, c := range conversationArray {
		cArray, ok := c.([]interface{})
		if !ok {
			continue
		}
		if len(cArray) < 1 {
			continue
		}
		cArray, ok = cArray[0].([]interface{})
		if !ok {
			continue
		}
		if len(cArray) < 6 {
			continue
		}
		id, ok := cArray[1].(string)
		if !ok {
			continue
		}
		title, ok := cArray[2].(string)
		if !ok {
			continue
		}
		cArray, ok = cArray[5].([]interface{})
		if !ok {
			continue
		}
		if len(cArray) < 1 {
			continue
		}
		time, ok := cArray[0].(float64)
		if !ok {
			continue
		}
		ret = append(ret, GoogleGroupConversation{
			GroupName: g.GroupName,
			Id: id,
			Title: title,
			Time: time,
			Cookie: g.Cookie,
		})
	}
	return ret
}
