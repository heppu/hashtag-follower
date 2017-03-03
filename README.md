# Hashtag Follower - telegram bot which follows instagram hastags

## Installation

Get the project

```sh
go get github.com/heppu/hashtag-follower
cd $GOPATH/src/github.com/heppu/hashtag-follower
```

Set your credentials in main.go

```go
const (
    botKey     = "" // telgram bot key
    igUser     = "" // instagram user name
    igPassword = "" // instgram user password
)
```

Build bot

```sh
go build
```

Drop binary to server and run just run it
