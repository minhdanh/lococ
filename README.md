# lococ

A Telegram bot application that monitors HackerNews top stories and RSS feeds, sending notifications directly to configured Telegram channels.

Here's how a message sent to your Telegram channel will look like:

![Sample Message](/sample_message.png)

# Features
- Filter HackerNews posts by point/score
- Shorten links with [bitly](https://bitly.com/)
- Send to multiple Telegram channels
- Retry messages if rate limited by Telegram

# Installation
`lococ` includes two Go applications:
- `lococ-job`: The main service that fetches news and posts to Telegram.
- `lococ-web`: An optional web application.

`lococ` requires a Redis server to track sent items and prevent duplicated messages.

The following commands will download and install `lococ-job` to a Linux server. Please refer to the [releases page](https://github.com/minhdanh/lococ/releases) for the latest version.
```bash
wget https://github.com/minhdanh/lococ/releases/download/v0.1.1/lococ-job-v0.1.1-linux-amd64.tar.gz -O lococ-job.tar.gz
tar xvf lococ-job.tar.gz
chmod +x lococ-job
sudo mv lococ-job /usr/local/bin/
```
Then create a directory for the configuration file:
```bash
sudo mkdir /etc/lococ
```
You will need to put a file named `config.yaml` to this directory. Please refer to section [Configurations](#configurations) for the content of this file. Make sure the values of the fields are set correctly.

After that we need to create a cronjob to run `lococ-job` periodically. For example the following cronjob will run `lococ-job` hourly:
```cron
0 * * * * /usr/local/bin/lococ-job --config-dir=/etc/lococ
```

That's it. Now wait for messages to be sent to your Telegram channel at the beginning of every hour.

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
```yaml
- name: BBC Vietnamese
  url: "https://www.bbc.co.uk/vietnamese/index.xml"
- name: StatusCode Weekly
  url: "https://weekly.statuscode.com/rss/"
  telegram_channel: "-1001340592770"
```

### Using config file (config.yaml)

```yaml
retry:
  enabled: true
  count: 3

telegram:
  channel: "@lococ"
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

```bash
lococ-job --config-dir=/app
```

# Development
There're Dockerfile and docker-compose.yml.sample files to help get this app up and running in a local environment. Remember to set the correct values for the environment variables.

```bash
# Create your docker-compose.yml file
cp docker-compose.yml.sample docker-compose.yml

# Then update the environment variables in docker-compose.yml

# Then start the containers
docker compose up

# Then run the job
docker compose run --rm lococ /bin/lococ-job
```
