package main

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/paalka/ewok/config"
	"github.com/paalka/ewok/db"
	"github.com/paalka/ewok/feed"
)

func main() {
	config := config.LoadConfig("../config.json")
	db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)

	feeds := feed.GetFeeds(db)
	SyncFeeds(db, feeds)
}

func fetchNewFeedItems(db *sql.DB, oldFeed feed.EwokFeed, fp *gofeed.Parser, feedDiffsCh chan<- feed.EwokFeed, chFinished chan<- bool) {
	newFeed, err := fp.ParseURL(oldFeed.Link)

	if err != nil {
		fmt.Printf("%s %s", err, oldFeed.Link)
	}

	if newFeed != nil {
		newItems, newFeedLastUpdated := feed.GetNewItems(db, newFeed, oldFeed)
		feedDiffsCh <- feed.EwokFeed{&gofeed.Feed{Updated: newFeedLastUpdated, Items: newItems}, oldFeed.Id}
	}

	chFinished <- true
}

func SyncFeeds(db *sql.DB, feeds []feed.EwokFeed) {
	feedDiffsCh := make(chan feed.EwokFeed)
	feedFinished := make(chan bool)

	fp := gofeed.NewParser()
	for _, feed := range feeds {
		go fetchNewFeedItems(db, feed, fp, feedDiffsCh, feedFinished)
	}

	var feedDiffs []feed.EwokFeed
	for c := 0; c < len(feeds); {
		select {
		case feedDiff := <-feedDiffsCh:
			feedDiffs = append(feedDiffs, feedDiff)
		case <-feedFinished:
			c++
		}
	}

	for _, feedDiff := range feedDiffs {
		feed.UpdateFeedFromDiff(db, feedDiff)
	}
}
