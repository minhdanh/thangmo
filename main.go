package main

import (
	"github.com/go-redis/redis"
	"github.com/minhdanh/thangmo-bot/internal"
	"github.com/minhdanh/thangmo-bot/pkg/bitly"
	"github.com/minhdanh/thangmo-bot/pkg/hackernews"
	"github.com/minhdanh/thangmo-bot/pkg/telegram"
	"github.com/ungerik/go-rss"
	"log"
	"strconv"
)

func main() {
	config := config.NewConfig()

	var items []interface{}
	rc := config.RedisClient

	if config.HackerNewsConfig.Enabled {
		hnClient := hackernews.NewHNClient()
		hnItemIDs := hnClient.GetItemIDs()

		for _, itemId := range hnItemIDs {
			if checked, _ := alreadyChecked(rc, strconv.Itoa(itemId)); !checked {
				continue
			}
			hnItem := hnClient.GetItem(itemId)
			if hnItem.Score >= config.HackerNewsConfig.MinScore {
				items = append(items, hnItem)
			} else {
				log.Printf("HackerNews item %v doesn't have enough points (%v).", itemId, hnItem.Score)
			}
			rc.Set(strconv.Itoa(itemId), "", 0)
		}
	}

	for _, rssChannel := range config.RSSChannels {
		log.Printf("Getting RSS content for %v", rssChannel.Name)
		channel, err := rss.Read(rssChannel.URL)
		if err != nil {
			log.Println(err)
			// TODO: notify about error such as time out when connecting to bbc
			continue
		}

		log.Printf("RSS channel %v has %v items", rssChannel.Name, len(channel.Item))
		for _, item := range channel.Item {
			if checked, _ := alreadyChecked(rc, item.Link); !checked {
				continue
			}
			items = append(items, item)
			rc.Set(item.Link, "", 0)
		}
	}

	for _, item := range items {
		var url string
		switch value := item.(type) {
		case hackernews.HNItem:
			url = value.URL
		case rss.Item:
			url = value.Link
		}
		if config.BitLyEnabled {
			bitly := bitly.NewClient(config.BitLyApiToken)
			url = bitly.ShortenUrl(url)
		}
		t := telegram.NewClient(config.TelegramApiToken, config.TelegramChannel)
		_, err := t.SendMessageForItem(item, url)
		if err != nil {
			log.Println(err)
		}
	}
}

func alreadyChecked(rc *redis.Client, key string) (bool, error) {
	if _, err := rc.Get(key).Result(); err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
