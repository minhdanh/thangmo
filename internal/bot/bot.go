package bot

import (
	"log"
	"reflect"
	"strconv"
	"strings"

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

const (
	VALIDATION_STATE_SUCCESS = iota
	VALIDATION_STATE_FAILED  = iota
	VALIDATION_STATE_WAITING = iota
)

type ValidationMessage struct {
	Message   string
	State     int
	NextStep  string
	NextState bool
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
	newStateMessages := map[string]string{
		"hackernews": "Do you want to receive posts from HackerNews?",
		"rss":        "Do you want to add RSS channels?",
	}

	stateChanged := map[int]bool{}

	funcMap := map[string]func(string) *ValidationMessage{
		"hackernews_enable": func(value string) *ValidationMessage {
			s := strings.ToLower(value)
			switch s {
			case "yes":
				// set redis to remember enablement for a period of time, until user finish next step
				message := ValidationMessage{
					Message:   "Minimum points of HackNews items?",
					State:     VALIDATION_STATE_SUCCESS,
					NextStep:  "minimum_points",
					NextState: false,
				}
				return &message
			case "no":
				message := ValidationMessage{
					Message:   "Ok. You don't want to receive news from HackerNews",
					State:     VALIDATION_STATE_SUCCESS,
					NextState: true,
				}
				return &message
			default:
				message := ValidationMessage{
					Message:   "Please choose `Yes` or `No`",
					State:     VALIDATION_STATE_FAILED,
					NextState: false,
				}
				return &message
			}
		},
		"hackernews_minimum_points": func(value string) *ValidationMessage {
			point, err := strconv.Atoi(value)
			if err != nil {
				message := ValidationMessage{
					Message:   "Score should be a number",
					State:     VALIDATION_STATE_FAILED,
					NextState: false,
				}
				return &message
			}
			if point > 500 || point < 1 {
				message := ValidationMessage{
					Message:   "Score should be in [1 - 500]",
					State:     VALIDATION_STATE_FAILED,
					NextState: false,
				}
				return &message
			}
			message := ValidationMessage{
				Message:   "Ok. You want to notified items with " + value + " points up.",
				State:     VALIDATION_STATE_SUCCESS,
				NextState: true,
			}
			return &message
		},
		"rss_add_channel": func(value string) *ValidationMessage {
			s := strings.ToLower(value)
			message := ValidationMessage{}
			switch s {
			case "yes":
				message.Message = "Great. Let's add your favourite RSS channel.\nWhat is the name of the RSS channel?"
				message.State = VALIDATION_STATE_SUCCESS
				message.NextStep = "channel_name"
				message.NextState = false
			case "no":
				message.Message = "Ok. You can change that later with `/rss add` command."
				message.State = VALIDATION_STATE_SUCCESS
				message.NextState = true
			default:
				message.Message = "Please choose `Yes` or `No`"
				message.State = VALIDATION_STATE_FAILED
				message.NextState = false
			}
			return &message
		},
		"rss_channel_name": func(value string) *ValidationMessage {
			message := ValidationMessage{}
			if value == "" || len(value) > 64 {
				message.Message = "Channel name is not valid. It should not be empty and not longer than 64 characters."
				message.State = VALIDATION_STATE_FAILED
				message.NextState = false
			} else {
				message.Message = "Your RSS channel name is " + value + ". What is the URL of the channel?\nIt's a link to the RSS feeds. For example: "
				message.State = VALIDATION_STATE_SUCCESS
				message.NextStep = "channel_url"
				message.NextState = false
			}
			return &message
		},
		"rss_channel_url": func(value string) *ValidationMessage {
			message := ValidationMessage{}
			if value == "" || len(value) > 64 {
				message.Message = "Channel name is not valid. It should not be empty and not longer than 64 characters."
				message.State = VALIDATION_STATE_FAILED
				message.NextState = false
			} else {
				message.Message = "Your RSS channel URL is " + value + "\nDo you want to add more channel?"
				message.State = VALIDATION_STATE_SUCCESS
				message.NextStep = "add_channel"
				message.NextState = false
			}
			return &message
		},
	}

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

			b.RedisClient.HSet(strconv.Itoa(userId), "current_state", command)

			switch command {
			case "start":
				msg.Text = "Hi! My name is Professor Bot. I will hold a conversation with you.\n\nSend /cancel to stop talking to me."
				b.RedisClient.HSet(strconv.Itoa(userId), "current_state", "hackernews")
				b.RedisClient.HSet(strconv.Itoa(userId), "last_command_callback", "enable")
				stateChanged[userId] = true
			case "hackernews":
				msg.Text = "Type /hackernews enable <min_score>"
			case "rss":
				msg.Text = "Type /rss add name link"
			default:
				msg.Text = "I don't know that command"
			}
			b.TelegramClient.TelegramBot.Send(msg)
			if stateChanged[userId] {
				newState := b.RedisClient.HGet(strconv.Itoa(userId), "current_state")
				msg.Text = newStateMessages[newState.Val()]
				b.TelegramClient.TelegramBot.Send(msg)
				stateChanged[userId] = false
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			currentState := b.RedisClient.HGet(strconv.Itoa(userId), "current_state")
			log.Printf("currentState: %s", currentState.Val())
			// switch currentState.Val() {
			// case "hackernews":
			lastCommandCallback := b.RedisClient.HGet(strconv.Itoa(userId), "last_command_callback")
			callbackFunc := currentState.Val() + "_" + lastCommandCallback.Val()
			log.Println(callbackFunc)
			message := funcMap[callbackFunc](incomingMessage)
			log.Println("====================")
			log.Println(message)
			log.Println("====================")
			if message.State == VALIDATION_STATE_SUCCESS {
				log.Println("validation succeeded")
				if message.NextStep == "" {
					log.Println("no next step")
					if message.NextState {
						log.Println("New state is true")
						b.RedisClient.HSet(strconv.Itoa(userId), "current_state", "rss")
						b.RedisClient.HSet(strconv.Itoa(userId), "last_command_callback", "add_channel")
						stateChanged[userId] = true
					} else {
						// do nothing
					}
				} else {
					b.RedisClient.HSet(strconv.Itoa(userId), "last_command_callback", message.NextStep)
				}
			}
			msg.Text = message.Message
			b.TelegramClient.TelegramBot.Send(msg)
			// case "rss":

			// }
			if stateChanged[userId] {
				newState := b.RedisClient.HGet(strconv.Itoa(userId), "current_state")
				log.Println("state changed. new state: ")
				log.Println(newState.Val())

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				msg.Text = newStateMessages[newState.Val()]
				log.Println(msg.Text)
				b.TelegramClient.TelegramBot.Send(msg)

				stateChanged[userId] = false
			}
		}
	}
}
