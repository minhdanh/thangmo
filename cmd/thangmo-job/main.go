package main

import (
	"crypto/md5"
	"github.com/go-redis/redis"
	"github.com/minhdanh/thangmo-bot/internal/config"
	"github.com/minhdanh/thangmo-bot/pkg/bitly"
	"github.com/minhdanh/thangmo-bot/pkg/hackernews"
	"github.com/minhdanh/thangmo-bot/pkg/telegram"
	"github.com/ungerik/go-rss"
	"log"
	"strconv"
)

type ItemWrapper struct {
	Item            interface{}
	Prefix          string
	RssLinkCheckSum string
}

func main() {
	config := config.NewConfig()

	var items []ItemWrapper

	rc := config.RedisClient

	if config.HackerNewsConfig.Enabled {
		log.Println("Getting top stories from HackerNews")
		hnClient := hackernews.NewHNClient()
		hnItemIDs := hnClient.GetItemIDs()

		for _, itemId := range hnItemIDs {
			if checked, err := alreadyChecked(rc, strconv.Itoa(itemId)); err != nil {
				log.Println(err)
				continue
			} else if checked {
				log.Printf("HackerNews item %v already checked", itemId)
				continue
			}
			hnItem := hnClient.GetItem(itemId)
			if hnItem.Score >= config.HackerNewsConfig.MinScore {
				items = append(items, ItemWrapper{Item: hnItem})
			} else {
				log.Printf("HackerNews item %v doesn't have enough points (%v)", itemId, hnItem.Score)
			}
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
			md5Sum := md5.Sum([]byte(item.Link))
			linkHash := string(md5Sum[:])
			if checked, err := alreadyChecked(rc, linkHash); err != nil {
				log.Println(err)
				continue
			} else if checked {
				log.Printf("RSS item \"%v\" already checked", item.Title)
				continue
			}
			items = append(items, ItemWrapper{Item: item, Prefix: rssChannel.Name, RssLinkCheckSum: linkHash})
		}
	}

	log.Printf("Processing %v items", len(items))

	t := telegram.NewClient(config.TelegramApiToken, config.TelegramChannel, config.TelegramPreviewLink, config.HackerNewsConfig.YcombinatorLink)
	for _, item := range items {
		var url, redisKey string
		switch value := item.Item.(type) {
		case hackernews.HNItem:
			log.Printf("Sending Telegram message, HackerNews item: %v", value.ID)
			url = value.URL
			redisKey = strconv.Itoa(value.ID)
		case rss.Item:
			log.Printf("Sending Telegram message, RSS item: \"%v\"", value.Title)
			url = value.Link
			redisKey = item.RssLinkCheckSum
		}
		if config.BitLyEnabled {
			bitly := bitly.NewClient(config.BitLyApiToken)
			url = bitly.ShortenUrl(url)
		}
		_, err := t.SendMessageForItem(item.Item, url, item.Prefix)
		rc.Set(redisKey, "", 0)
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