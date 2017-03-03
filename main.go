package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/heppu/hashtag-follower/db"
	"github.com/heppu/instagram-open-sdk"

	"gopkg.in/telegram-bot-api.v4"
)

const (
	botKey     = "" // telgram bot key
	igUser     = "" // instagram user name
	igPassword = "" // instgram user password
)

type LastInfo struct {
	Count     int
	LastStamp int
}

var (
	bot      *tgbotapi.BotAPI
	igClient *ig.Client
	tagDb    *db.Client
	lastTags = make(map[int64]map[string]*LastInfo)
	tagLock  = &sync.Mutex{}
)

func init() {
	var err error

	if tagDb, err = db.NewClient("instatele.db", "tags"); err != nil {
		log.Fatal(err)
	}

	if igClient, err = ig.NewClient(igUser, igPassword); err != nil {
		log.Fatal(err)
	}

	if bot, err = tgbotapi.NewBotAPI(botKey); err != nil {
		log.Fatal(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

}

func main() {
	go tagLoop()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		checkChat(update.Message.Chat.ID)

		switch update.Message.Command() {
		case "add":
			log.Printf("[%s] Add hashtag #%s", update.Message.From.UserName, update.Message.CommandArguments())
			addTag(&update)

		case "del":
			log.Printf("[%s] Del hashtag #%s", update.Message.From.UserName, update.Message.CommandArguments())
			delTag(&update)

		case "list":
			log.Printf("[%s] List hashtags", update.Message.From.UserName)
			listTags(&update)
		}
	}
}

func checkChat(chatID int64) {
	if _, ok := lastTags[chatID]; ok {
		return
	}

	tags, err := tagDb.GetTags(chatID)
	if err != nil {
		log.Println(err)
	}

	lastTags[chatID] = make(map[string]*LastInfo)
	for tag, _ := range tags {
		lastTags[chatID][tag] = &LastInfo{}
	}
}

func addTag(update *tgbotapi.Update) {
	tagLock.Lock()
	defer tagLock.Unlock()

	tag := strings.TrimSpace(strings.Split(update.Message.CommandArguments(), " ")[0])
	if len(tag) == 0 {
		reply(update, "Empty tag")
		return
	}

	if _, ok := lastTags[update.Message.Chat.ID][tag]; ok {
		reply(update, "Tag already on list")
		return
	}

	lastTags[update.Message.Chat.ID][tag] = &LastInfo{}
	var err error
	if err = tagDb.AddTag(update.Message.Chat.ID, tag); err != nil {
		reply(update, err.Error())
		return
	}

	lastTags[update.Message.Chat.ID][tag] = &LastInfo{}
	reply(update, "Added tag: "+tag)
}

func delTag(update *tgbotapi.Update) {
	tagLock.Lock()
	defer tagLock.Unlock()

	tag := strings.Split(update.Message.CommandArguments(), " ")[0]

	if len(tag) == 0 {
		reply(update, "Empty tag")
		return
	}

	if _, ok := lastTags[update.Message.Chat.ID][tag]; !ok {
		reply(update, "Tag not on list")
		return
	}

	var err error
	if err = tagDb.DeleteTag(update.Message.Chat.ID, tag); err != nil {
		reply(update, err.Error())
		return
	}

	delete(lastTags[update.Message.Chat.ID], tag)
	reply(update, "Deleted tag: "+tag)
}

func listTags(update *tgbotapi.Update) {
	if len(lastTags[update.Message.Chat.ID]) == 0 {
		reply(update, "No hastags to follow")
		return
	}
	tags := ""
	i := 1
	for k, _ := range lastTags[update.Message.Chat.ID] {
		tags += fmt.Sprintf("%d : %s\n", i, k)
		i++
	}

	reply(update, tags[:len(tags)-1])
}

func reply(update *tgbotapi.Update, message string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
}

func tagLoop() {
	for {
		<-time.After(time.Second * 5)
		tagLock.Lock()

		for chatID, data := range lastTags {
			for tag, info := range data {
				go updateImage(chatID, tag, info)
			}
		}

		tagLock.Unlock()
	}
}

func updateImage(chatID int64, tag string, info *LastInfo) {
	res, err := igClient.TagService.Recent(tag)
	if err != nil {
		log.Println(err)
	}

	if len(res.Data.Nodes) == 0 {
		log.Println("No nodes for tag", tag)
		return
	}

	if res.Data.Nodes[0].Date <= info.LastStamp {
		return
	}

	info.Count = res.Data.Count
	info.LastStamp = res.Data.Nodes[0].Date

	u, err := url.Parse(res.Data.Nodes[0].DisplaySrc)
	if err != nil {
		log.Println(err)
		return
	}

	str := res.Data.Nodes[0].Caption
	str += "\n\n" + u.String()
	msg := tgbotapi.NewMessage(chatID, str)
	bot.Send(msg)
}
