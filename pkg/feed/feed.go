package feed

import (
	"database/sql"
)

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
