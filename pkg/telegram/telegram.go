package telegram

import (
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/minhdanh/thangmo-bot/pkg/hackernews"
	"github.com/ungerik/go-rss"
	"log"
	"strconv"
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

func (t *TelegramClient) SendMessageForItem(item interface{}, url string, messagePrefix string) (tgbotapi.Message, error) {
	switch value := item.(type) {
	case hackernews.HNItem:
		return t.sendMessageForHNItem(value, url)
	case rss.Item:
		return t.sendMessageForRSSItem(value, url, messagePrefix)
	}
	return tgbotapi.Message{}, errors.New("Item type is incorrect")
}

func (t *TelegramClient) sendMessageForRSSItem(item rss.Item, url string, messagePrefix string) (tgbotapi.Message, error) {
	msgBody := item.Title + "\n" + url
	if messagePrefix != "" {
		msgBody = messagePrefix + ": " + item.Title + "\n" + url
	}
	msg := tgbotapi.NewMessageToChannel(t.TelegramChannel, msgBody)
	msg.DisableWebPagePreview = t.TelegramPreviewLink
	msg.ParseMode = "HTML"
	msg.BaseChat.DisableNotification = true

	return t.TelegramBot.Send(msg)
}

func (t *TelegramClient) sendMessageForHNItem(item hackernews.HNItem, url string) (tgbotapi.Message, error) {
	msgBody := "HackerNews: " + item.Title + " (" + strconv.Itoa(item.Score) + " points)"
	if url != "" {
		msgBody += "\n" + url
	}
	if t.YcombinatorLink {
		msgBody += "\n" + "https://news.ycombinator.com/item?id=" + strconv.Itoa(item.ID)
	}
	msg := tgbotapi.NewMessageToChannel(t.TelegramChannel, msgBody)
	msg.DisableWebPagePreview = t.TelegramPreviewLink
	msg.ParseMode = "HTML"
	msg.BaseChat.DisableNotification = true

	return t.TelegramBot.Send(msg)
}
