package main

import (
	"crypto/md5"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/minhdanh/lococ/internal/config"
	"github.com/minhdanh/lococ/pkg/bitly"
	"github.com/minhdanh/lococ/pkg/hackernews"
	"github.com/minhdanh/lococ/pkg/telegram"
	"github.com/mmcdole/gofeed"
)

type ItemWrapper struct {
	Item            interface{}
	Prefix          string
	RssLinkCheckSum string
	TelegramChannel string
}

func getHNItem(itemsChan chan ItemWrapper, hnClient *hackernews.HNClient, rc *redis.Client, minScore int, itemId int, wg *sync.WaitGroup) {
	defer wg.Done()

	if checked, err := alreadyChecked(rc, strconv.Itoa(itemId)); err != nil {
		log.Println(err)
		return
	} else if checked {
		log.Printf("HackerNews item %v already checked", itemId)
		return
	}
	hnItem, err := hnClient.GetItem(itemId)
	if err != nil {
		log.Printf("Error getting HackerNews item %v: %v", itemId, err)
		return
	}
	if hnItem.Score >= minScore {
		itemsChan <- ItemWrapper{Item: hnItem}
	} else {
		log.Printf("HackerNews item %v doesn't have enough points (%v)", itemId, hnItem.Score)
	}
}

func runCheck(cfg *config.Config) {
	var items []ItemWrapper

	rc := cfg.RedisClient

	if cfg.HackerNewsConfig.Enabled {
		log.Println("Getting top stories from HackerNews")
		hnClient := hackernews.NewHNClient()
		hnItemIDs := hnClient.GetItemIDs()

		itemsChan := make(chan ItemWrapper)
		wg := new(sync.WaitGroup)

		for _, itemId := range hnItemIDs {
			wg.Add(1)
			go getHNItem(itemsChan, hnClient, rc, cfg.HackerNewsConfig.MinScore, itemId, wg)
		}

		go func() {
			wg.Wait()
			close(itemsChan)
		}()

		for item := range itemsChan {
			items = append(items, item)
		}
	}

	for _, rssChannel := range cfg.RSSChannels {
		log.Printf("Getting RSS content for %v", rssChannel.Name)

		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(rssChannel.URL)
		if err != nil {
			log.Println(err)
			// TODO: notify about error such as time out when connecting to bbc
			continue
		}

		log.Printf("RSS channel %v has %v items", rssChannel.Name, len(feed.Items))
		for _, item := range feed.Items {
			md5Sum := md5.Sum([]byte(strings.TrimSpace(item.Link)))
			linkHash := string(md5Sum[:])
			if checked, err := alreadyChecked(rc, linkHash); err != nil {
				log.Println(err)
				continue
			} else if checked {
				log.Printf("RSS item \"%v\" already checked", item.Title)
				continue
			}
			items = append(items, ItemWrapper{Item: item, Prefix: rssChannel.Name, RssLinkCheckSum: linkHash, TelegramChannel: rssChannel.TelegramChannel})
		}
	}

	log.Printf("Processing %v items", len(items))

	t := telegram.NewClient(cfg.TelegramApiToken, cfg.TelegramChannel, cfg.TelegramPreviewLink, cfg.HackerNewsConfig.YcombinatorLink)
	for _, item := range items {
		var url, redisKey string
		switch value := item.Item.(type) {
		case *hackernews.HNItem:
			log.Printf("Sending Telegram message, HackerNews item: %v", value.ID)
			url = value.URL
			redisKey = strconv.Itoa(value.ID)
		case *gofeed.Item:
			log.Printf("Sending Telegram message, RSS item: \"%v\"", value.Title)
			url = strings.TrimSpace(value.Link)
			redisKey = item.RssLinkCheckSum
		}
		if cfg.BitLyEnabled {
			bitly := bitly.NewClient(cfg.BitLyApiToken)
			url = bitly.ShortenUrl(url)
		}
		if !cfg.DryRun {
			for i := 0; i < cfg.RetryCount+1; i++ {
				if i > 0 {
					log.Println("Retrying message...")
				}
				_, err := t.SendMessageForItem(item.Item, url, item.Prefix, item.TelegramChannel)

				if err != nil {
					log.Println(err)
					err := err.(tgbotapi.Error)
					// wait to retry only if rate limited
					if err.Code == 429 {
						log.Printf("Rate limited. Waiting for %v seconds before retrying.", err.ResponseParameters.RetryAfter)
						time.Sleep(time.Second * time.Duration(err.ResponseParameters.RetryAfter+1))
					}
				} else {
					rc.Set(redisKey, "", 0)
					break
				}

				if !cfg.RetryEnabled {
					break
				}
			}
		} else {
			log.Println("dry-run mode is enabled. Not sending messages.")
		}
	}
}

func main() {
	cfg := config.NewConfig()

	if cfg.Once {
		log.Println("Running lococ once...")
		runCheck(cfg)
		log.Println("Check completed.")
		return
	}

	log.Printf("Starting lococ service with interval %v...", cfg.Interval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Executing initial check on startup...")
	runCheck(cfg)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Executing scheduled check...")
			runCheck(cfg)
		case sig := <-sigChan:
			log.Printf("Received signal %v. Shutting down gracefully...", sig)
			return
		}
	}
}

func alreadyChecked(rc *redis.Client, key string) (bool, error) {
	if _, err := rc.Get(key).Result(); err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
