package fetcher

import (
	"fmt"
	"github.com/mmcdole/gofeed"
)

type Feed struct {
	LastPublished string
	FeedUrl       string
}

func ReadFeed(feed Feed, fp *gofeed.Parser, ch chan<- Feed, chFinished chan<- bool) {
	rssFeed, err := fp.ParseURL(feed.FeedUrl)

	if err != nil {
		fmt.Printf("%s %s", err, feed.FeedUrl)
	}

	if rssFeed != nil {
		ch <- Feed{LastPublished: rssFeed.Items[0].Published, FeedUrl: feed.FeedUrl}
	}

	chFinished <- true
}
