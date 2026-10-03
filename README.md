# Google Groups Crawler

A Go library that reads the conversations and messages of a [Google Group](https://groups.google.com/) from its web pages.

We use it to sync posts between the [Casnode](https://github.com/casdoor/casnode) forum and our Casdoor Google Group. Please only use it on groups you are allowed to read.

## Installation

```shell
go get github.com/casbin/google-groups-crawler
```

```go
import crawler "github.com/casbin/google-groups-crawler"
```

## Quick start

```go
package main

import (
	"fmt"
	"net/http"

	crawler "github.com/casbin/google-groups-crawler"
)

func main() {
	group := crawler.NewGoogleGroup("golang-announce")

	conversations := group.GetAllConversations(http.Client{})
	fmt.Printf("%d conversations\n", len(conversations))

	for _, conversation := range conversations[:3] {
		messages := conversation.GetAllMessages(http.Client{}, true)
		fmt.Printf("%s: %d messages\n", conversation.Title, len(messages))
	}
}
```

## Usage

### Create a group

```go
group := crawler.NewGoogleGroup(groupName string, cookie ...string)
```

- `groupName` is the name in the group URL, e.g. `golang-nuts` for `https://groups.google.com/g/golang-nuts`. A group email such as `golang-nuts@googlegroups.com` also works.
- `cookie` is optional, see [Cookie](#cookie).

### Get all conversations

```go
conversations := group.GetAllConversations(client http.Client) []GoogleGroupConversation
```

Returns every conversation of the group, newest first. The web page shows 30 conversations at a time and loads more as you scroll down. This function loads all of these pages, so a large group needs one request per 30 conversations.

### Get all messages of a conversation

```go
messages := conversation.GetAllMessages(client http.Client, removeGmailQuote bool) []GoogleGroupMessage
```

Returns the messages of a conversation in posting order. Set `removeGmailQuote` to `true` to cut off the quoted previous message (`<div class="gmail_quote">`) from each reply.

### HTTP client

Both functions take an `http.Client`, so you can set a timeout or a proxy. `http.Client{}` uses the `HTTP_PROXY` / `HTTPS_PROXY` environment variables. If Google Groups is blocked in your network, set a proxy explicitly:

```go
proxyUrl, _ := url.Parse("http://127.0.0.1:10809")
client := http.Client{
	Timeout:   30 * time.Second,
	Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
}
conversations := group.GetAllConversations(client)
```

### Cookie

Without a cookie the crawler can read public groups, but Google hides the senders' email addresses, so `AuthorEmail` is empty. To get the email addresses, or to read a private group, pass the cookie of a logged-in user who is a member of the group:

1. Log in to Google in Chrome and open the group on [Google Groups](https://groups.google.com/).
2. Press F12 and open the **Network** tab.
3. Open any conversation of the group and select the first request in the list.
4. Under **Headers** → **Request Headers**, copy the value of `cookie`.
5. Pass it to `NewGoogleGroup("group-name", cookie)`.

The cookie is copied to every conversation returned by `GetAllConversations`. Keep it secret: it gives access to that Google account.

## Data structure

```go
type GoogleGroup struct {
	GroupName string
	Cookie    string
}

type GoogleGroupConversation struct {
	Title     string
	Id        string // the id in https://groups.google.com/g/<group>/c/<id>
	GroupName string
	Time      float64 // Unix time in seconds of the latest message
	Cookie    string
}

type GoogleGroupMessage struct {
	Author      string
	AuthorEmail string // empty without a member's cookie
	Content     string // HTML
	Time        float64 // Unix time in seconds
	Files       []GoogleGroupFile
}

type GoogleGroupFile struct {
	FileName string
	Url      string
	Type     string // MIME type
}
```

## Limitations

- Google Groups has no public API. The crawler parses the data embedded in the web pages and calls the same internal endpoint the page uses for loading more conversations. If Google changes the page, results may become empty until the crawler is updated.
- Errors are not returned. A failed request prints a message and returns what has been read so far, which can be an empty slice.
- At most 1000 messages are read from one conversation, the same as the web page loads at once.

## License

[Apache 2.0](LICENSE)
