# thangmo
I used to use an app named Hooks to receive notifications of news items from HackerNews (with a minimum score of 200) and from some RSS channels to keep myself updated with the world. The app had been working great for me for many years, until when it was abandoned. I had to find an alternative for Hooks but there was no app in the App Store that offered the same experience like Hooks. Feedly was one of them. Then I came up with the idea of writing a bot to send notifications to my Telegram account. Hence this app.

The name `thangmo` (full Vietnamese: `thằng mõ`) is the role name of a man in the villages of ancient Vietnam. His job was to go around the village to tell the villagers to gather at the communal house of the village. The people then will come to hear about the recent news in the area and the new policies from the feudal system.

Here's how a message sent to your Telegram channel will look like:

![Sample Message](/sample_message.png)

# Features
- Filter HackerNews posts by point/score
- Shorten links with [bitly](https://bitly.com/)
- Send to multiple Telegram channels
- Retry messages if rate limited by Telegram
- That's all :-p

# Installation
`thangmo` includes two smaller go apps: `thangmo-web` and `thangmo-job`.
`thangmo-job` is the main command that actually does the job.
`thangmo-web` was done just for fun :D. It's not needed for the job to function.

`thangmo` requires a Redis server running so it can avoid sending duplicated messages, so please make sure you have a Redis server ready.

The following commands will download and install `thangmo-job` to a Linux server. Please refer to the [releases page](https://github.com/minhdanh/thangmo/releases) for the latest version.
```
wget https://github.com/minhdanh/thangmo/releases/download/v0.1.1/thangmo-job-v0.1.1-linux-amd64.tar.gz -O thangmo-job.tar.gz
tar xvf thangmo-job.tar.gz
chmod +x thangmo-job
sudo mv thangmo-job /usr/local/bin/
```
Then create a directory for the configuration file:
```
sudo mkdir /etc/thangmo
```
You will need to put a file named `config.yaml` to this directory. Please refer to section [Configurations](#configurations) for the content of this file. Make sure the values of the fields are set correctly.

After that we need to create a cronjob to run `thangmo-job` periodically. For example the following cronjob will run `thangmo-job` hourly:
```
0 * * * * /usr/local/bin/thangmo-job --config-dir=/etc/thangmo
```

That's it. Now wait for messages to be sent to your Telegram channel at the begining of every hour.

# Configurations
You can use environment variables or a config file to deploy the bot.

### Using environment variables
- `HACKERNEWS_ENABLED`: Enable HackerNews notifications.
- `HACKERNEWS_MIN_SCORE`: The minimum score of a news item.
- `HACKERNEWS_YCOMBINATOR_LINK`: Whether or not to include the link to HackerNews.

- `TELEGRAM_CHANNEL`: The Telegram channel to send notifications to.
- `TELEGRAM_API_TOKEN`: Telegram API token.

- `BITLY_ENABLED`: Enable this to have shortened links.
- `BITLY_API_TOKEN`: Bitly API token.
- `REDISCLOUD_URL`: Redis URL. This is used to make sure we don't receive duplicated notifications.
- `RSS_CONFIG_BASE64`: A list of RSS channels encoded in base64 format. Useful if you want to deploy this on Heroku. Just encode a list of the channels (be careful with the indent whitespaces). For example:
```
- name: BBC Vietnamese
  url: "https://www.bbc.co.uk/vietnamese/index.xml"
- name: StatusCode Weekly
  url: "https://weekly.statuscode.com/rss/"
  telegram_channel: "-1001340592770"
```

### Using config file (config.yaml)

```
retry:
  enabled: true
  count: 3

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
  - name: StatusCode Weekly
    url: "https://weekly.statuscode.com/rss/"
    telegram_channel: "-1001340592770"

redis:
  host: localhost
  port: 6379
  username: ""
  password: ""
```
# Heroku deployment
[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

You can click the `Deploy to Heroku` button above to deploy this app to Heroku.
Please note that you will need to configure Heroku Scheduler to run this command periodically:

```
thangmo-job --config-dir=/app
```

# Development
There're Dockerfile and docker-compose.yml.sample files to help get this app up and running in a local environment. Remember to set the correct values for the environment variables.

```
# Create your docker-compose.yml file
cp docker-compose.yml.sample docker-compose.yml

# Then update the environment variables in docker-compose.yml

# Then start the containers
docker compose up

# Then run the job
docker compose run --rm thangmo /bin/thangmo-job
```
