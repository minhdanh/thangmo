package config

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Config struct {
	TelegramApiToken string
	TelegramChannel  string
	BitLyEnabled     bool
	BitLyApiToken    string
	HackerNewsConfig *HackerNewsConfig
	RSSChannels      []RSSChannel
	RedisClient      *redis.Client
}

type HackerNewsConfig struct {
	Enabled         bool
	MinScore        int
	YcombinatorLink bool
}

type RSSChannel struct {
	Name string
	URL  string
}

func NewConfig() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Error: config.yaml not found.")
		} else {
			log.Println(err)
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var config Config

	config.TelegramChannel = viper.GetString("telegram.channel")
	config.TelegramApiToken = viper.GetString("telegram.api_token")
	// bitly
	config.BitLyEnabled = viper.GetBool("bitly.enabled")
	config.BitLyApiToken = viper.GetString("bitly.api_token")
	// hackernews
	hnConfig := HackerNewsConfig{}
	hnConfig.Enabled = viper.GetBool("hacker_news.enabled")
	hnConfig.MinScore = viper.GetInt("hacker_news.min_score")
	hnConfig.YcombinatorLink = viper.GetBool("hacker_news.ycombinator_link")
	config.HackerNewsConfig = &hnConfig
	// rss
	var rssChannels []RSSChannel
	viper.UnmarshalKey("rss", &rssChannels)
	config.RSSChannels = rssChannels
	// redis
	rc := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.host") + ":" + viper.GetString("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       0,
	})
	config.RedisClient = rc

	return &config
}
