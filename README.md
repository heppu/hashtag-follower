# Hashtag Follower
 
Telegram bot which follows instagram hastags and post those to user/group

## Installation

Get the project

```sh
go get github.com/heppu/hashtag-follower
cd $GOPATH/src/github.com/heppu/hashtag-follower
```

Build bot

```sh
go build
```

Drop binary to server and run just run it with enviroment variables

```sh
env BOTKEY=<telegram bot key> IGUSER=<instragram user> IGPASSWORD=<instagram pass> ./hashtag-follower
```
