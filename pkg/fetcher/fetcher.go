package fetcher

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/ewok/pkg/feed"
)

func FetchFeed(f feed.RSSFeed, fp *gofeed.Parser, ch chan<- feed.RSSFeed, chFinished chan<- bool) {
	rssFeed, err := fp.ParseURL(f.Url)

	if err != nil {
		fmt.Printf("%s %s", err, f.Url)
	}

	if rssFeed != nil {
		feedLastUpdated := rssFeed.Updated
		if feedLastUpdated == "" && len(rssFeed.Items) > 0 {
			feedLastUpdated = rssFeed.Items[0].Published
		}

		ch <- feed.RSSFeed{LastUpdated: feedLastUpdated, Url: f.Url}
	}

	chFinished <- true
}
