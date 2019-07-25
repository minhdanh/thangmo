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
	TelegramBot     *tgbotapi.BotAPI
	TelegramChannel string
}

func NewClient(apiToken string, channel string) *TelegramClient {
	var client TelegramClient
	telegramBot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		log.Panic(err)
	}
	telegramBot.Debug = false
	client.TelegramBot = telegramBot
	client.TelegramChannel = channel

	return &client
}

func (t *TelegramClient) SendMessageForItem(item interface{}, url string) (tgbotapi.Message, error) {
	switch value := item.(type) {
	case hackernews.HNItem:
		return t.sendMessageForHNItem(value, url)
	case rss.Item:
		return t.sendMessageForRSSItem(value, url)
	}
	return tgbotapi.Message{}, errors.New("Item type is incorrect")
}

func (t *TelegramClient) sendMessageForRSSItem(item rss.Item, url string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessageToChannel(t.TelegramChannel, item.Title+"\n"+url)
	msg.DisableWebPagePreview = false
	msg.ParseMode = "HTML"
	msg.BaseChat.DisableNotification = true

	return t.TelegramBot.Send(msg)
}

func (t *TelegramClient) sendMessageForHNItem(item hackernews.HNItem, url string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessageToChannel(t.TelegramChannel, "HackerNews: "+item.Title+" ("+strconv.Itoa(item.Score)+" points)"+"\n"+url)
	msg.DisableWebPagePreview = false
	msg.ParseMode = "HTML"
	msg.BaseChat.DisableNotification = true

	return t.TelegramBot.Send(msg)
}
