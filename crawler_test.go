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
	"net/http"
	"testing"
	"time"
)

// a public group with more than one page of conversations
const testGroupName = "golang-announce"

func TestGetAllConversations(t *testing.T) {
	group := NewGoogleGroup(testGroupName)
	conversations := group.GetAllConversations(http.Client{Timeout: 30 * time.Second})

	if len(conversations) <= conversationPageSize {
		t.Fatalf("got %d conversations, want more than one page (%d)", len(conversations), conversationPageSize)
	}

	ids := map[string]bool{}
	for _, c := range conversations {
		if c.Id == "" || c.Time == 0 || c.GroupName != testGroupName {
			t.Fatalf("invalid conversation: %+v", c)
		}
		if ids[c.Id] {
			t.Fatalf("duplicate conversation: %s", c.Id)
		}
		ids[c.Id] = true
	}
}

func TestGetAllMessages(t *testing.T) {
	client := http.Client{Timeout: 30 * time.Second}
	conversations := NewGoogleGroup(testGroupName).GetAllConversations(client)
	if len(conversations) == 0 {
		t.Fatal("got no conversations")
	}

	messages := conversations[0].GetAllMessages(client, true)
	if len(messages) == 0 {
		t.Fatalf("got no messages in conversation %s", conversations[0].Id)
	}
	for _, m := range messages {
		if m.Author == "" || m.Content == "" || m.Time == 0 {
			t.Fatalf("invalid message: %+v", m)
		}
	}
}
