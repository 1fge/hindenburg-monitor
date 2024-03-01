package main

import (
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	postsRoute   = "https://hindenburgresearch.com/wp-json/wp/v2/posts"
	mediaRoute   = "https://hindenburgresearch.com/wp-json/wp/v2/media"
	sitemapRoute = "https://hindenburgresearch.com/wp-sitemap-posts-post-1.xml"
)

var (
	debug        = false
	refreshDelay = 2500 * time.Millisecond
	webhookURL   string

	ErrLoadingEnv        = errors.New("Loading Environment Values")
	ErrInvalidWebhookURL = errors.New("Invalid Webhook URL")
)

func main() {
	if err := loadWebhookURL(); err != nil {
		panic(err)
	}

	postsMonitor, err := newRouteMonitor("Posts", postsRoute, refreshDelay, wpjsonUpdateDetector)
	if err != nil {
		panic(err)
	}

	mediaMonitor, err := newRouteMonitor("Media", mediaRoute, refreshDelay, wpjsonUpdateDetector)
	if err != nil {
		panic(err)
	}

	sitemapMonitor, err := newRouteMonitor("Sitemap", sitemapRoute, refreshDelay, sitemapUpdateDetector)
	if err != nil {
		panic(err)
	}

	go postsMonitor.Monitor()
	go mediaMonitor.Monitor()
	go sitemapMonitor.Monitor()

	select {}
}

// loadWebhookURL pulls the discord webhook from ENV and ensures it is a proper URL
func loadWebhookURL() error {
	if err := godotenv.Load(); err != nil {
		return ErrLoadingEnv
	}

	webhookURL = os.Getenv("WEBHOOKURL")

	parsedHook, err := url.ParseRequestURI(webhookURL)
	if err != nil {
		return ErrInvalidWebhookURL
	}

	if webhookURL == "" || parsedHook.Scheme == "" || parsedHook.Host == "" {
		return ErrInvalidWebhookURL
	}

	return nil
}
