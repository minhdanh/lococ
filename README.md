# lococ

A Telegram bot daemon that monitors HackerNews top stories and RSS feeds, sending notifications directly to configured Telegram channels.

Here's how a message sent to your Telegram channel will look like:

![Sample Message](/sample_message.png)

# Features
- Filter HackerNews posts by point/score
- Shorten links with [bitly](https://bitly.com/)
- Send to multiple Telegram channels
- Retry messages if rate limited by Telegram
- Self-scheduling background daemon with configurable execution interval
- Native systemd service support and graceful shutdown

# Installation

`lococ` requires a Redis server to track sent items and prevent duplicated messages.

### 1. Download and Install Binary
Download the latest `lococ` binary from the [releases page](https://github.com/minhdanh/lococ/releases):

```bash
wget https://github.com/minhdanh/lococ/releases/download/v0.2.0/lococ-v0.2.0-linux-amd64.tar.gz -O lococ.tar.gz
tar xvf lococ.tar.gz
chmod +x lococ
sudo mv lococ /usr/local/bin/
```

### 2. Configuration
Create a directory for the configuration file:
```bash
sudo mkdir /etc/lococ
```
Place your `config.yaml` inside `/etc/lococ/`. Refer to [Configurations](#configurations) below for available options.

### 3. Setup systemd Service
Copy the provided systemd service file:
```bash
sudo cp lococ.service /etc/systemd/system/lococ.service
sudo systemctl daemon-reload
sudo systemctl enable --now lococ
```

Check the service status:
```bash
sudo systemctl status lococ
```

View real-time logs using `journalctl`:
```bash
journalctl -u lococ -f
```

### Manual / One-off Execution
You can run a single check cycle without starting the continuous scheduler using the `--once` flag:
```bash
lococ --config-dir=/etc/lococ --once
```

# Configurations
You can configure `lococ` using environment variables, command-line flags, or a `config.yaml` file.

### Command-line Flags
- `--config-dir`: Directory containing `config.yaml` (default: `/etc/lococ`)
- `--interval`: Interval between checks in daemon mode, e.g. `1h`, `30m` (default: `1h`)
- `--once`: Run check once and exit immediately
- `--dry-run`: Fetch items without sending Telegram messages

### Using environment variables
- `INTERVAL`: Check interval (e.g. `1h`, `30m`)
- `HACKERNEWS_ENABLED`: Enable HackerNews notifications (`true`/`false`)
- `HACKERNEWS_MIN_SCORE`: The minimum score of a news item
- `HACKERNEWS_YCOMBINATOR_LINK`: Whether or not to include the link to HackerNews
- `TELEGRAM_CHANNEL`: The Telegram channel to send notifications to
- `TELEGRAM_API_TOKEN`: Telegram API token
- `BITLY_ENABLED`: Enable Bitly link shortening
- `BITLY_API_TOKEN`: Bitly API token
- `REDIS_URL`: Redis URL (e.g. `redis://localhost:6379`)
- `RSS_CONFIG_BASE64`: A list of RSS channels encoded in base64 format. For example:
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

# Interval between periodic checks
interval: 1h

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

# Development
`Dockerfile` and `docker-compose.yml.sample` files are provided for local development:

```bash
# Create your docker-compose.yml file
cp docker-compose.yml.sample docker-compose.yml

# Update the environment variables in docker-compose.yml

# Start containers (daemon mode)
docker compose up

# Or execute a one-off run
docker compose run --rm lococ /bin/lococ --once
```
