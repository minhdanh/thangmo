package bot

import (
	"log"
	"reflect"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"github.com/minhdanh/thangmo-bot/internal/config"
	"github.com/minhdanh/thangmo-bot/pkg/telegram"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Bot struct {
	TelegramClient *telegram.TelegramClient
	RedisClient    *redis.Client
}

func NewBot(config *config.Config) *Bot {
	var bot Bot
	t := telegram.NewClient(config.TelegramApiToken, config.TelegramChannel, config.TelegramPreviewLink, config.HackerNewsConfig.YcombinatorLink)
	bot.TelegramClient = t
	bot.RedisClient = config.RedisClient

	db, err := gorm.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	tables := []interface{}{&RSSLink{}, &HNRegistration{}, &RSSRegistration{}}
	for _, table := range tables {
		tableType := reflect.TypeOf(table)
		if !db.HasTable(tableType) {
			db.CreateTable(tableType)
		}
		db.AutoMigrate(tableType)
	}
	defer db.Close()

	return &bot
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.TelegramClient.TelegramBot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		userId := update.Message.From.ID
		incomingMessage := update.Message.Text
		log.Printf("[%v] %s", userId, incomingMessage)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			command := update.Message.Command()

			b.RedisClient.HSet(strconv.Itoa(userId), "last_command", command)

			switch command {
			case "start":
				msg.Text = "Hello. Do you want to receive posts from HackerNews?"
				// 'Hi! My name is Professor Bot. I will hold a conversation with you. '
				// 'Send /cancel to stop talking to me.\n\n'
				// 'Are you a boy or a girl?',
				b.RedisClient.HSet(strconv.Itoa(userId), "last_command_callback", func(value string) string {
					return value
				})
			case "help":
				msg.Text = "type /sayhi or /status."
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			default:
				msg.Text = "I don't know that command"
			}
			b.TelegramClient.TelegramBot.Send(msg)
		} else {
			lastCommand := b.RedisClient.HGet(strconv.Itoa(userId), "last_command")
			log.Printf("Lastcommand: %s", lastCommand.Val())
			switch lastCommand.Val() {
			case "start":
				log.Println("start command entered")
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				lastCommandCallback := b.RedisClient.HGet(strconv.Itoa(userId), "last_command_callback")
				var fn func(a string) string
				lastCommandCallback.Scan(fn)
				aaa := fn("string")
				log.Println(aaa)
				msg.Text = ""
				// value := lastCommandCallback(incomingMessage)
				// msg.Text = "You entered " + value
				// b.TelegramClient.TelegramBot.Send(msg)
			}

		}

	}
}
