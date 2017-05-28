package main

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/ewok/config"
	"github.com/paalka/ewok/db"
	"github.com/paalka/ewok/feed"
	"github.com/paalka/ewok/feed_fetcher"
)

func main() {
	config := config.LoadConfig("../config.json")
	db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)

	feeds := feed.GetFeeds(db)
	SyncFeeds(db, feeds)
}

func SyncFeeds(db *sql.DB, feeds []feed.RSSFeed) {
	messages := make(chan feed.RSSFeed)
	feedFinished := make(chan bool)

	fp := gofeed.NewParser()
	for _, feed := range feeds {
		go feed_fetcher.FetchFeed(db, feed, fp, messages, feedFinished)
	}

	for c := 0; c < len(feeds); {
		select {
		case f := <-messages:
			fmt.Println(f.LastUpdated)
		case <-feedFinished:
			c++
		}
	}
}
