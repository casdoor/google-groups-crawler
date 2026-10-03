# Google Groups Crawler

## Installation

``` go
import crawler "github.com/casbin/google-groups-crawler"
```

## Usage

We must get an instance of `GoogleGroup` first:

``` go
group := crawler.NewGoogleGroup(groupName string, cookie ...string)
```

The second parameter `cookie` is optional. Google group won't tell you email address of all repliers until you logged
in, so you need to fill the parameter with a logged-in user's cookie. (Of course, this user must be a member of the
group)

It is OK to leave `cookie` blank, code still works. But `AuthorEmail` in `GoogleGroupMessage` will be empty. If you do need `cookie` to access emails of repliers, please follow these steps:

- open Google Chrome (or another browser) and login
- Navigate to [Google Group](https://groups.google.com/), select the group you want to craw
- Press F12, and select `network`
- Select a conversation (any conversation in this group is OK)
- Select the first item in the request list
- Select `Headers`
- In `Request Headers`, right click `cookie`, and copy the value
- Fill the parameter `cookie` with what you copied

### Get all conversations of the group

- For some special reasons, you cannot access Google Groups in some area. You can set up a http proxy, and fill the parameter `http.Client` with it. If you can access Google Groups directly, then you can just fill the parameter like the example code.
- this function returns an array of `GoogleGroupConversation`, it loads all pages of the group (30 conversations per page) like scrolling down in the web page

``` go
conversations := group.GetAllConversations(http.Client{})
```

### Get all messages of the conversation

- `conversation` is an instance of `GoogleGroupConversation`
- parameter `http.Client` is the same effect as above
- this function returns an array of `GoogleGroupMessage`

```go
messages := conversation.GetAllMessages(http.Client{}, removeGmailQuote)
```

## Data Structure

```go
type GoogleGroup struct {
    GroupName string
    Cookie    string
}

type GoogleGroupConversation struct {
    Title     string
    Id        string
    GroupName string
    Time      float64
    Cookie    string
}

type GoogleGroupMessage struct {
    Author      string
    AuthorEmail string
    Content     string
    Time        float64
    Files       []GoogleGroupFile
}

type GoogleGroupFile struct {
    FileName string
    Url      string
    Type     string
}
```