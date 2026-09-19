package config

import (
	"encoding/base64"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
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
	DryRun              bool
	RetryEnabled        bool
	RetryCount          int
	Interval            time.Duration
	Once                bool
}

type HackerNewsConfig struct {
	Enabled         bool
	MinScore        int
	YcombinatorLink bool
}

type RSSChannel struct {
	Name            string
	URL             string
	TelegramChannel string `mapstructure:"telegram_channel" yaml:"telegram_channel"`
}

func NewConfig() *Config {
	configDir := ""

	if flag.Lookup("config-dir") == nil {
		flag.String("config-dir", "/etc/lococ", "Default config directory")
	}
	if pflag.Lookup("dry-run") == nil {
		pflag.Bool("dry-run", false, "Do not send real Telegram messages")
	}
	if pflag.Lookup("interval") == nil {
		pflag.String("interval", "1h", "Execution interval for periodic checks (e.g. 1h, 30m)")
	}
	if pflag.Lookup("once") == nil {
		pflag.Bool("once", false, "Execute check once and exit immediately")
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	if !pflag.Parsed() {
		pflag.Parse()
	}
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

	// defaults
	viper.SetDefault("RetryEnabled", false)
	viper.SetDefault("RetryCount", 1)
	viper.SetDefault("Interval", "1h")

	config.TelegramChannel = viper.GetString("telegram.channel")
	config.TelegramApiToken = viper.GetString("telegram.api_token")
	config.TelegramPreviewLink = viper.GetBool("telegram.preview_link")

	// retry
	config.RetryEnabled = viper.GetBool("retry.enabled")
	config.RetryCount = viper.GetInt("retry.count")

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

	rssBase64 := os.Getenv("RSS_CONFIG_BASE64")
	if rssBase64 != "" {
		log.Println("Env var RSS_CONFIG_BASE64 detected, will be used for RSS channels config.")
		sDec, err := base64.StdEncoding.DecodeString(rssBase64)
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			err := yaml.Unmarshal([]byte(sDec), &rssChannels)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}

	config.RSSChannels = rssChannels

	// redis
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = os.Getenv("REDISCLOUD_URL")
	}

	var redisOptions *redis.Options
	if redisUrl != "" {
		var err error
		redisOptions, err = redis.ParseURL(redisUrl)
		if err != nil {
			log.Printf("Warning: error parsing Redis URL: %v", err)
		} else {
			log.Println("Using Redis config from URL")
		}
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

	// interval
	intervalStr := viper.GetString("interval")
	parsedInterval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Printf("Warning: invalid interval '%v', falling back to 1h: %v", intervalStr, err)
		parsedInterval = 1 * time.Hour
	}
	config.Interval = parsedInterval

	// flags
	config.DryRun = viper.GetBool("dry-run")
	config.Once = viper.GetBool("once")

	return &config
}
