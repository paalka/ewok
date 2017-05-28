package feed_fetcher

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/ewok/feed"
)

func FetchFeed(db *sql.DB, f feed.EwokFeed, fp *gofeed.Parser, ch chan<- feed.EwokFeed, chFinished chan<- bool) {
	rssFeed, err := fp.ParseURL(f.Link)

	if err != nil {
		fmt.Printf("%s %s", err, f.Link)
	}

	if rssFeed != nil {
		feedLastUpdated := rssFeed.Updated
		if feedLastUpdated == "" && len(rssFeed.Items) > 0 {
			feedLastUpdated = rssFeed.Items[0].Published
		}

		feed.UpdateItems(db, f.Id, rssFeed, feedLastUpdated, f.Updated)

		ch <- feed.EwokFeed{&gofeed.Feed{Title: f.Title, Updated: feedLastUpdated, Link: f.Link}, f.Id}
	}

	chFinished <- true
}
