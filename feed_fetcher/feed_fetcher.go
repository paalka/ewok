package feed_fetcher

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/ewok/feed"
)

func FetchFeed(db *sql.DB, f feed.RSSFeed, fp *gofeed.Parser, ch chan<- feed.RSSFeed, chFinished chan<- bool) {
	rssFeed, err := fp.ParseURL(f.Url)

	if err != nil {
		fmt.Printf("%s %s", err, f.Url)
	}

	if rssFeed != nil {
		feedLastUpdated := rssFeed.Updated
		if feedLastUpdated == "" && len(rssFeed.Items) > 0 {
			feedLastUpdated = rssFeed.Items[0].Published
		}

		feed.UpdateItems(db, f.Id, rssFeed, feedLastUpdated, f.LastUpdated)

		ch <- feed.RSSFeed{Title: f.Title, LastUpdated: feedLastUpdated, Url: f.Url}
	}

	chFinished <- true
}
