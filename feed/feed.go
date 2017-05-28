package feed

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"time"
)

const timeLayoutRSS string = time.RFC1123Z
const timeLayoutPSQL string = time.RFC3339

type EwokFeed struct {
	*gofeed.Feed
	Id uint
}

type FeedItem struct {
	Title       string
	Link        string
	Description string
	Published   string
}

func GetFeeds(db *sql.DB) []EwokFeed {
	rows, err := db.Query("SELECT id, title, url, last_updated FROM rss.rss_feed")

	if err != nil {
		panic(err)
	}

	var feeds []EwokFeed
	for rows.Next() {
		f := EwokFeed{&gofeed.Feed{}, 0}
		err = rows.Scan(&f.Id, &f.Title, &f.Link, &f.Updated)
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

func UpdateItems(db *sql.DB, newFeed *gofeed.Feed, oldFeed EwokFeed) {
	newFeedLastUpdated := newFeed.Updated
	if newFeedLastUpdated == "" && len(newFeed.Items) > 0 {
		newFeedLastUpdated = newFeed.Items[0].Published
	}
	newLastUpdatedTime := parseLastUpdated(timeLayoutRSS, newFeedLastUpdated)
	oldLastUpdatedTime := parseLastUpdated(timeLayoutPSQL, oldFeed.Updated)

	if newLastUpdatedTime.Before(oldLastUpdatedTime) {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	ins_stmt, err := db.Prepare("INSERT INTO rss.rss_item (title, description, link, publish_date) VALUES ($1, $2, $3, $4)")
	if err != nil {
		panic(err)
	}

	for _, item := range newFeed.Items {
		if item.PublishedParsed != nil && item.PublishedParsed.After(oldLastUpdatedTime) {
			_, err := tx.Stmt(ins_stmt).Exec(item.Title, item.Description, item.Link, item.Published)

			if err != nil {
				tx.Rollback()
				panic(err)
			}
		}
	}

	update_feed_stmt, err := db.Prepare("UPDATE rss.rss_feed SET last_updated = $1 WHERE id = $2")
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	_, err = tx.Stmt(update_feed_stmt).Exec(newFeedLastUpdated, oldFeed.Id)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	tx.Commit()
}
