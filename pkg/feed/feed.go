package feed

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"time"
)

const timeLayoutRSS string = time.RFC1123Z
const timeLayoutPSQL string = time.RFC3339

type RSSFeed struct {
	Title       string
	Url         string
	LastUpdated string
}

func GetFeeds(db *sql.DB) []RSSFeed {
	rows, err := db.Query("SELECT title, url, last_updated FROM rss.rss_feed")

	if err != nil {
		panic(err)
	}

	var feeds []RSSFeed
	for rows.Next() {
		var f RSSFeed
		err = rows.Scan(&f.Title, &f.Url, &f.LastUpdated)
		if err != nil {
			panic(err)
		}
		feeds = append(feeds, f)
	}

	return feeds
}

func parseLastUpdated(timeLayout string, timeString string) time.Time {
	t, err := time.Parse(timeLayout, timeString)
	if err != nil {
		fmt.Println(err)
	}

	return t
}

