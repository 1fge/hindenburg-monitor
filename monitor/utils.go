package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/1fge/hindenburg-monitor/internal-http-pkgs/discordwebhook"
	http "github.com/1fge/hindenburg-monitor/internal-http-pkgs/fhttp"
	"github.com/tidwall/gjson"

	tls_client "github.com/1fge/hindenburg-monitor/internal-http-pkgs/tls-client"
	"github.com/1fge/hindenburg-monitor/internal-http-pkgs/tls-client/profiles"
)

// headers we'll use with each request
var defaultGetHeaders http.Header = http.Header{
	"sec-ch-ua":                 {"\"Chromium\";v=\"122\", \"Not(A:Brand\";v=\"24\", \"Google Chrome\";v=\"122\""},
	"sec-ch-ua-mobile":          {"?0"},
	"sec-ch-ua-platform":        {"\"Windows\""},
	"upgrade-insecure-requests": {"1"},
	"user-agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"},
	"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	"dnt":                       {"1"},
	"sec-fetch-site":            {"same-origin"},
	"sec-fetch-mode":            {"navigate"},
	"sec-fetch-user":            {"?1"},
	"sec-fetch-dest":            {"document"},
	"referer":                   {"https://hindenburgresearch.com/"},
	"accept-encoding":           {"gzip, deflate, br, zstd"},
	"accept-language":           {"en-US,en;q=0.9"},
	http.HeaderOrderKey: {
		"sec-ch-ua",
		"sec-ch-ua-mobile",
		"sec-ch-ua-platform",
		"upgrade-insecure-requests",
		"user-agent",
		"accept",
		"dnt",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-user",
		"sec-fetch-dest",
		"referer",
		"accept-encoding",
		"accept-language",
	},
}

// structs used for parsing xml sitemap
type UrlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Urls    []Url    `xml:"url"`
}

type Url struct {
	Loc string `xml:"loc"`
}

// create a new client with new jar, proxy, etc.
func createNewClient() (tls_client.HttpClient, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	// add charles proxy to debug if flag set
	if debug {
		options = append(options, tls_client.WithProxyUrl("http://127.0.0.1:8888"))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// cacheBustURL adds timestamp to our fetch to avoid caches from kinsta
func cacheBustURL(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	params := parsedURL.Query()
	params.Add("ts", timestamp)

	parsedURL.RawQuery = params.Encode()
	return parsedURL.String(), nil
}

// extractWPJSONNewLoad returns a map[string]ContentItem on fresh starts for monitors
// this func works with posts & media since they are in the same format
func wpjsonExtractFresh(data gjson.Result) map[string]ContentItem {
	var freshMap map[string]ContentItem = make(map[string]ContentItem)

	for _, entry := range data.Array() {
		entryURL := entry.Get("link").String()
		entryTitle := entry.Get("title.rendered").String()

		freshMap[entryURL] = ContentItem{
			URL:   entryURL,
			title: entryTitle,
		}
	}

	// on first load, set existing data w/o triggering monitor
	return freshMap
}

// sitemapExtractFresh pulls all sitemap entries when our current
// cache for entries is empty
func sitemapExtractFresh(resp []byte) (map[string]ContentItem, error) {
	var freshMap map[string]ContentItem = make(map[string]ContentItem)

	allURLs, err := pullAllSitemapItems(resp)
	if err != nil {
		return nil, err
	}

	for _, URL := range allURLs {
		freshMap[URL] = ContentItem{
			URL:   URL,
			title: URL,
		}
	}

	return freshMap, nil
}

// pullAllSitemapItems returns a slice of URLs we pull from sitemap page
func pullAllSitemapItems(resp []byte) ([]string, error) {
	var urlSet UrlSet
	var parsedURLs []string

	if err := xml.Unmarshal([]byte(resp), &urlSet); err != nil {
		return nil, fmt.Errorf("Error parsing XML: %v\n", err)
	}

	for _, url := range urlSet.Urls {
		parsedURLs = append(parsedURLs, url.Loc)
	}

	return parsedURLs, nil
}

// sendDiscordAlert sends discord webhook with the new content found, along with where we found it
func sendDiscordAlert(URL, title, source string) {
	color := "1294635"

	message := discordwebhook.Message{
		Username: "Hindenburg Site Update",
		Content:  "",
		Embeds: []discordwebhook.Embed{
			discordwebhook.Embed{
				Title:       "Update Via " + source,
				Url:         URL,
				Description: "Content Title: " + title,
				Color:       color,
				Fields: []discordwebhook.Field{
					discordwebhook.Field{
						Name:   "Raw URL", // creation of our *SnipeItem
						Value:  fmt.Sprintf("`%s`", URL),
						Inline: true,
					},
				},
			},
		},
	}

	if err := discordwebhook.SendMessage(webhookURL, message); err != nil {
		fmt.Println("Failed sending webhook", err)
	}
}

func printErrAndWait(source string, err error, delay time.Duration) {
	fmt.Printf("[%s] - %s - Error: %s\n", conciseTime(), source, err.Error())
	time.Sleep(delay)
}

func conciseTime() string {
	now := time.Now()
	conciseDateTime := now.Format("01-02-2006 15:04:05")
	return conciseDateTime
}
