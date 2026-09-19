package config

import (
	"encoding/base64"
	"flag"
	"log"
	"os"
	"strings"

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
	Port                int
	DryRun              bool
	RetryEnabled        bool
	RetryCount          int
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

	flag.String("config-dir", "/etc/lococ", "Default config directory")
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

	// defaults
	viper.SetDefault("Port", 3000)
	viper.SetDefault("RetryEnabled", false)
	viper.SetDefault("RetryCount", 1)

	config.Port = viper.GetInt("port")

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
	redisCloudUrl := os.Getenv("REDISCLOUD_URL")
	redisOptions, err := redis.ParseURL(redisCloudUrl)
	if err != nil {
		log.Printf("Warning: %v", err)
	} else {
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
