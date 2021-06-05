package config

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramApiToken    string
	TelegramChannel     string
	TelegramPreviewLink bool
	BitLyEnabled        bool
	BitLyApiToken       string
	HackerNewsConfig    *HackerNewsConfig
	RSSChannels         []RSSChannel
	RedisClient         *redis.Client
	Port                int
	DryRun              bool
}

type HackerNewsConfig struct {
	Enabled         bool
	MinScore        int
	YcombinatorLink bool
}

type RSSChannel struct {
	Name            string
	URL             string
	TelegramChannel string `mapstructure:"telegram_channel"`
}

func NewConfig() *Config {
	configDir := ""

	flag.String("config-dir", "/etc/thangmo", "Default config directory")
	pflag.Bool("dry-run", false, "Do not send real Telegram messages")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	configDir = viper.GetString("config-dir")

	if configDir != "" {
		viper.AddConfigPath(configDir)
		log.Printf("Using config dir: %v", configDir)
	}

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

	viper.SetDefault("Port", 3000)
	config.Port = viper.GetInt("port")

	config.TelegramChannel = viper.GetString("telegram.channel")
	config.TelegramApiToken = viper.GetString("telegram.api_token")
	config.TelegramPreviewLink = viper.GetBool("telegram.preview_link")
	// bitly
	config.BitLyEnabled = viper.GetBool("bitly.enabled")
	config.BitLyApiToken = viper.GetString("bitly.api_token")
	// hackernews
	hnConfig := HackerNewsConfig{}
	hnConfig.Enabled = viper.GetBool("hackernews.enabled")
	hnConfig.MinScore = viper.GetInt("hackernews.min_score")
	hnConfig.YcombinatorLink = viper.GetBool("hackernews.ycombinator_link")
	config.HackerNewsConfig = &hnConfig
	// rss
	var rssChannels []RSSChannel
	viper.UnmarshalKey("rss", &rssChannels)
	config.RSSChannels = rssChannels
	// redis
	redisCloudUrl := os.Getenv("REDISCLOUD_URL")
	redisOptions, err := redis.ParseURL(redisCloudUrl)
	if err == nil {
		log.Println("Using Redis config from REDISCLOUD_URL")
	}

	if redisOptions == nil {
		redisOptions = &redis.Options{
			Addr:     viper.GetString("redis.host") + ":" + viper.GetString("redis.port"),
			Password: viper.GetString("redis.password"),
			DB:       0,
		}
	}
	rc := redis.NewClient(redisOptions)
	config.RedisClient = rc

	// dry-run
	config.DryRun = viper.GetBool("dry-run")
	return &config
}
