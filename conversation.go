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
	"strings"
)

func (c GoogleGroupConversation) GetAllMessages(client http.Client, removeGmailQuote bool) []GoogleGroupMessage {
	targetUrl := fmt.Sprintf("https://groups.google.com/g/%s/c/%s", c.GroupName, c.Id)
	var ret []GoogleGroupMessage

	body, ok := getPage(client, targetUrl, c.Cookie)
	if !ok {
		return ret
	}

	dataArray, ok := getInitData(body, rpcListMessages)
	if !ok {
		return ret
	}
	if len(dataArray) < 3 {
		return ret
	}
	msgArray, ok := dataArray[2].([]interface{})
	if !ok {
		return ret
	}

	for _, msg := range msgArray {
		var files []GoogleGroupFile
		singleMsgArray, ok := msg.([]interface{})
		if !ok || len(singleMsgArray) < 1 {
			continue
		}
		singleMsgArray, ok = singleMsgArray[0].([]interface{})
		if !ok || len(singleMsgArray) < 2 {
			continue
		}
		singleMsgArray0, ok := singleMsgArray[0].([]interface{})
		if !ok || len(singleMsgArray0) < 9 {
			continue
		}
		// attachments are only present when the message has any
		if len(singleMsgArray) > 2 {
			singleMsgArray2, ok := singleMsgArray[2].([]interface{})
			if ok && len(singleMsgArray2) > 0 {
				for _, singleFileArray := range singleMsgArray2 {
					singleFile, ok := singleFileArray.([]interface{})
					if !ok || len(singleFile) < 5 {
						continue
					}
					fileName, _ := singleFile[4].(string)
					fileUrl, _ := singleFile[0].(string)
					fileType, _ := singleFile[3].(string)
					files = append(files, GoogleGroupFile{
						FileName: fileName,
						Url: fileUrl,
						Type: fileType,
					})
				}
			}
		}
		authorEmailArray, ok := singleMsgArray0[2].([]interface{})
		if !ok || len(authorEmailArray) < 1 {
			continue
		}
		authorEmailArray, ok = authorEmailArray[0].([]interface{})
		if !ok || len(authorEmailArray) < 1 {
			continue
		}
		author, ok := authorEmailArray[0].(string)
		if !ok {
			continue
		}
		// the email is null unless the cookie of a group member is provided,
		// and senders without a Google account only have a name
		email := ""
		if len(authorEmailArray) > 2 {
			email, _ = authorEmailArray[2].(string)
		}
		singleMsgArray0, ok = singleMsgArray0[8].([]interface{})
		if !ok || len(singleMsgArray0) < 1 {
			continue
		}
		time, ok := singleMsgArray0[0].(float64)
		if !ok {
			continue
		}
		singleMsgArray1, ok := singleMsgArray[1].([]interface{})
		if !ok || len(singleMsgArray1) < 2 {
			continue
		}
		singleMsgArray1, ok = singleMsgArray1[1].([]interface{})
		if !ok || len(singleMsgArray1) < 1 {
			continue
		}
		singleMsgArray1, ok = singleMsgArray1[0].([]interface{})
		if !ok || len(singleMsgArray1) < 2 {
			continue
		}
		singleMsgArray1, ok = singleMsgArray1[1].([]interface{})
		if !ok || len(singleMsgArray1) < 2 {
			continue
		}
		content, ok := singleMsgArray1[1].(string)
		if !ok {
			continue
		}

		if removeGmailQuote {
			content = strings.Split(content, "<div class=\"gmail_quote\">")[0]
		}

		ret = append(ret, GoogleGroupMessage{
			Author: author,
			AuthorEmail: email,
			Content: content,
			Time: time,
			Files: files,
		})
	}
	return ret
}

