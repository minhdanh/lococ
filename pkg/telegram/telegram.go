package telegram

import (
	"errors"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/minhdanh/lococ/pkg/hackernews"
	"github.com/mmcdole/gofeed"
)

type TelegramClient struct {
	TelegramBot         *tgbotapi.BotAPI
	TelegramChannel     string
	TelegramPreviewLink bool
	YcombinatorLink     bool
}

func NewClient(apiToken string, channel string, previewLink bool, ycombinatorLink bool) *TelegramClient {
	var client TelegramClient
	telegramBot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		log.Panic(err)
	}
	telegramBot.Debug = false
	client.TelegramBot = telegramBot
	client.TelegramChannel = channel
	client.TelegramPreviewLink = previewLink
	client.YcombinatorLink = ycombinatorLink

	return &client
}

func (t *TelegramClient) SendMessageForItem(item interface{}, url string, messagePrefix string, telegramChannel string) (tgbotapi.Message, error) {
	switch value := item.(type) {
	case *hackernews.HNItem:
		return t.sendMessageForHNItem(value, url)
	case *gofeed.Item:
		return t.sendMessageForRSSItem(value, url, messagePrefix, telegramChannel)
	}
	return tgbotapi.Message{}, errors.New("Item type is incorrect")
}

func (t *TelegramClient) sendMessageForRSSItem(item *gofeed.Item, url string, messagePrefix string, telegramChannel string) (tgbotapi.Message, error) {
	msgBody := strings.TrimSpace(item.Title)
	if url != "" {
		msgBody = "<a href=\"" + url + "\">" + msgBody + "</a>"
	}
	if messagePrefix != "" {
		msgBody = "<strong>" + messagePrefix + "</strong>" + ": " + msgBody
	}
	msg := tgbotapi.MessageConfig{}
	if telegramChannel != "" {
		msg = tgbotapi.NewMessageToChannel(telegramChannel, msgBody)
	} else {
		msg = tgbotapi.NewMessageToChannel(t.TelegramChannel, msgBody)
	}
	msg.DisableWebPagePreview = t.TelegramPreviewLink
	msg.ParseMode = "HTML"
	msg.BaseChat.DisableNotification = true

	return t.TelegramBot.Send(msg)
}

func (t *TelegramClient) sendMessageForHNItem(item *hackernews.HNItem, url string) (tgbotapi.Message, error) {
	msgBody := "<strong>HackerNews</strong>: "
	if t.YcombinatorLink {
		msgBody = "<strong><a href=\"https://news.ycombinator.com/item?id=" + strconv.Itoa(item.ID) + "\">HackerNews</a></strong>: "
	}

	if url != "" {
		msgBody += "<a href=\"" + url + "\">" + item.Title + "</a>"
	} else {
		msgBody += item.Title
	}
	msg := tgbotapi.NewMessageToChannel(t.TelegramChannel, msgBody)
	msg.DisableWebPagePreview = t.TelegramPreviewLink
	msg.ParseMode = "HTML"
	msg.BaseChat.DisableNotification = true

	return t.TelegramBot.Send(msg)
}
