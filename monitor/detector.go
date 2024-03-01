package main

import (
	"fmt"
	"io"
	"time"

	http "github.com/1fge/hindenburg-monitor/internal-http-pkgs/fhttp"
	tls_client "github.com/1fge/hindenburg-monitor/internal-http-pkgs/tls-client"
	"github.com/tidwall/gjson"
)

// ContentItem stores metadata about each URL found on the page, allowing us to track new content
type ContentItem struct {
	URL   string
	title string
}

// ChangeDetectionFunc defines a function type for detecting changes in content
type ChangeDetectionFunc func([]byte, map[string]ContentItem) (map[string]ContentItem, *ContentItem, error)

// RouteMonitor holds the configuration for monitoring a specific route.
type RouteMonitor struct {
	source       string
	URL          string
	delay        time.Duration
	detectChange ChangeDetectionFunc

	client      tls_client.HttpClient
	lastContent map[string]ContentItem
}

func newRouteMonitor(source, URL string, delay time.Duration, changeFunc ChangeDetectionFunc) (*RouteMonitor, error) {
	client, err := createNewClient()
	if err != nil {
		return nil, err
	}

	return &RouteMonitor{
		source:       source,
		URL:          URL,
		delay:        delay,
		detectChange: changeFunc,

		client:      client,
		lastContent: make(map[string]ContentItem),
	}, nil
}

// fetchEndpoint retrieves data from the struct's specified endpoint.
func (r *RouteMonitor) fetchEndpoint() ([]byte, error) {
	fmt.Printf("[%s] - %s - Fetching Endpoint\n", r.source, conciseTime())

	modifiedURL, err := cacheBustURL(r.URL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", modifiedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = defaultGetHeaders

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad Status Fetching: %d", resp.StatusCode)
	}

	return body, nil
}

func (r *RouteMonitor) Monitor() {
	for {
		resp, err := r.fetchEndpoint()
		if err != nil {
			printErrAndWait(r.source, err, r.delay)
			continue
		}

		allContent, newContentItem, err := r.detectChange(resp, r.lastContent)
		if err != nil {
			printErrAndWait(r.source, err, r.delay)
			continue
		}

		if newContentItem != nil {
			fmt.Printf("[%s] - %s - Detected New Content: %s\n", r.source, conciseTime(), newContentItem.title)
			go sendDiscordAlert(newContentItem.URL, newContentItem.title, r.source)
		}

		// update our last content to be new contents of site
		r.lastContent = allContent
		time.Sleep(r.delay)
	}
}

// wpjsonUpdateDetector is used to find changes on both /posts & /media since the structure
// for both endpoints is identical -- they simply handle separate data.
// detects changes from /wp-json/wp/v2/posts or /wp-json/wp/v2/media
func wpjsonUpdateDetector(resp []byte, lastContent map[string]ContentItem) (map[string]ContentItem, *ContentItem, error) {
	var data gjson.Result = gjson.ParseBytes(resp)
	if len(lastContent) == 0 {
		lastContent = wpjsonExtractFresh(data)
		return lastContent, nil, nil
	}

	for _, entry := range data.Array() {
		entryURL := entry.Get("link").String()
		entryTitle := entry.Get("title.rendered").String()

		// if we already have the item, skip it
		if _, ok := lastContent[entryURL]; ok {
			continue
		}

		// otherwise, add it to our content cache & return new content item
		newContentItem := ContentItem{
			URL:   entryURL,
			title: entryTitle,
		}

		lastContent[entryURL] = newContentItem
		return lastContent, &newContentItem, nil
	}

	// if we reach here, we found no new content
	return lastContent, nil, nil
}

// sitemapUpdateDetector handles detecting changes from the sitemap found at
// https://hindenburgresearch.com/wp-sitemap-posts-post-1.xml
func sitemapUpdateDetector(resp []byte, lastContent map[string]ContentItem) (map[string]ContentItem, *ContentItem, error) {
	if len(lastContent) == 0 {
		lastContent, err := sitemapExtractFresh(resp)
		return lastContent, nil, err
	}

	sitemapURLs, err := pullAllSitemapItems(resp)
	if err != nil {
		return lastContent, nil, err
	}

	for _, URL := range sitemapURLs {
		if _, ok := lastContent[URL]; ok {
			continue
		}

		newContentItem := ContentItem{
			URL:   URL,
			title: URL,
		}

		lastContent[URL] = newContentItem
		return lastContent, &newContentItem, nil
	}

	return lastContent, nil, nil
}
