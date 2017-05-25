package fetcher

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/rss/pkg/feed"
)

func ReadFeed(f feed.RSSFeed, fp *gofeed.Parser, ch chan<- feed.RSSFeed, chFinished chan<- bool) {
	rssFeed, err := fp.ParseURL(f.Url)

	if err != nil {
		fmt.Printf("%s %s", err, f.Url)
	}

	if rssFeed != nil {
		ch <- feed.RSSFeed{LastUpdated: rssFeed.Items[0].Published, Url: f.Url}
	}

	chFinished <- true
}
