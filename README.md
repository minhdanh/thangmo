# thangmo-bot
I used to use an app named Hooks to receive notifications of news items from HackerNews (with a minimum score of 200) and from some RSS channels to keep myself updated with the world. The app had been working great for me for many years, until when it was abandoned. I had to find an alternative for Hooks but there was no app in the App Store that offered the same experience like Hooks. Feedly is one of them. Then I came up with the idea of writing a bot to send notifications to my Telegram account. Hence this app.

The name `thangmo` is the role of a man in the villages of ancient Vietnam. His job was to go around the village to tell the villagers to gather at the communal house of the village. The people then will come to hear about the recent news in the area and the new policies from the feudal system.

If you want to see how this bot works, check out this Telegram channel: https://t.me/thangmo

# Configurations
You can use environment variables or a config file to deploy the bot. For RSS channels you need to use the config file.

### Using environment variables
- `HACKERNEWS_ENABLED`: Enable HackerNews notifications.
- `HACKERNEWS_MIN_SCORE`: The minimum score of a news item.
- `HACKERNEWS_YCOMBINATOR_LINK`: Whether or not to include the link to HackerNews.

- `TELEGRAM_CHANNEL`: The Telegram channel to send notifications to.
- `TELEGRAM_API_TOKEN`: Telegram API token.

- `BITLY_ENABLED`: Enable this to have shortened links.
- `BITLY_API_TOKEN`: Bitly API token.
- `REDISCLOUD_URL`: Redis URL. This is used to make sure we don't receive duplicated notifications.

### Using config file (config.yaml)

```
telegram:
  channel: "@thangmo"
  api_token: "<TELEGRAM API TOKEN>"

bitly:
  enabled: true
  api_token: "<BITLY API TOKEN>"

hackernews:
  enabled: true
  min_score: 200
  ycombinator_link: true

rss:
  - name: BBC Vietnamese
    url: "https://www.bbc.co.uk/vietnamese/index.xml"
  - name: Cafebiz - Chính sách
    url: "http://cafebiz.vn/chinh-sach.rss"

redis:
  host: localhost
  port: 6379
  username: ""
  password: ""
```
