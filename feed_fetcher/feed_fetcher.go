package feed_fetcher

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/ewok/feed"
)

func FetchFeed(db *sql.DB, oldFeed feed.EwokFeed, fp *gofeed.Parser, ch chan<- feed.EwokFeed, chFinished chan<- bool) {
	newFeed, err := fp.ParseURL(oldFeed.Link)

	if err != nil {
		fmt.Printf("%s %s", err, oldFeed.Link)
	}

	if newFeed != nil {
		feed.UpdateItems(db, newFeed, oldFeed)

		ch <- feed.EwokFeed{&gofeed.Feed{Title: oldFeed.Title, Updated: "sad", Link: oldFeed.Link}, oldFeed.Id}
	}

	chFinished <- true
}
