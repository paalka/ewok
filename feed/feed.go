package feed

import (
	"database/sql"
	"fmt"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"strings"
	"time"
)

const timeLayoutRSS string = time.RFC1123Z
const timeLayoutPSQL string = time.RFC3339

type EwokItem struct {
	*gofeed.Item
	ParentFeedId uint
}

type EwokFeed struct {
	*gofeed.Feed
	Items []*EwokItem
	Id    uint
}

func UpdateFeedFromDiff(db *sql.DB, feedDiff EwokFeed) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	ins_stmt, err := db.Prepare("INSERT INTO rss_item (title, description, link, publish_date, parent_feed_id) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		panic(err)
	}

	for _, item := range feedDiff.Items {
		var description string
		if len(item.Description) > 350 {
			description = item.Description[:350]

		}

		strippedText, err := html2text.FromString(description)

		if !strings.HasSuffix(strippedText, " […]") {
			strippedText = strippedText + " […]"
		}
		if err != nil {
			panic(err)
		}

		_, err = tx.Stmt(ins_stmt).Exec(item.Title, strippedText, item.Link, item.Published, feedDiff.Id)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	update_feed_stmt, err := db.Prepare("UPDATE rss_feed SET last_updated = $1 WHERE id = $2")
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	_, err = tx.Stmt(update_feed_stmt).Exec(feedDiff.Updated, feedDiff.Id)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	tx.Commit()
}

func GetPaginatedFeeds(db *sql.DB, nItems uint, offset uint) []EwokItem {
	rows, err := db.Query("SELECT title, link, description, publish_date, parent_feed_id FROM rss_item ORDER BY publish_date DESC OFFSET $1 LIMIT $2", nItems*offset, nItems)

	if err != nil {
		panic(err)
	}

	var feedItems []EwokItem
	for rows.Next() {
		item := EwokItem{&gofeed.Item{}, 0}
		err = rows.Scan(&item.Title, &item.Link, &item.Description, &item.Published, &item.ParentFeedId)
		if err != nil {
			panic(err)
		}
		item.Published = ParseTime(timeLayoutPSQL, item.Published).Format(time.RFC1123)
		feedItems = append(feedItems, item)
	}

	return feedItems
}

func GetAllFeedItems(db *sql.DB) []EwokItem {
	rows, err := db.Query("SELECT title, link, description, publish_date, parent_feed_id FROM rss_item ORDER BY publish_date DESC")

	if err != nil {
		panic(err)
	}

	var feedItems []EwokItem
	for rows.Next() {
		item := EwokItem{&gofeed.Item{}, 0}
		err = rows.Scan(&item.Title, &item.Link, &item.Description, &item.Published, &item.ParentFeedId)
		if err != nil {
			panic(err)
		}
		item.Published = ParseTime(timeLayoutPSQL, item.Published).Format(time.RFC1123)
		feedItems = append(feedItems, item)
	}

	return feedItems
}

func GetFeeds(db *sql.DB) []EwokFeed {
	rows, err := db.Query("SELECT id, title, url, last_updated FROM rss_feed")

	if err != nil {
		panic(err)
	}

	var feeds []EwokFeed
	for rows.Next() {
		var tmpItems []*EwokItem
		f := EwokFeed{&gofeed.Feed{}, tmpItems, 0}
		err = rows.Scan(&f.Id, &f.Title, &f.Link, &f.Updated)
		if err != nil {
			panic(err)
		}
		feeds = append(feeds, f)
	}

	return feeds
}

func ParseTime(timeLayout string, timeString string) time.Time {
	t, err := time.Parse(timeLayout, timeString)
	if err != nil {
		fmt.Println(err)
	}

	return t
}

func GetNewItems(db *sql.DB, newFeed *gofeed.Feed, oldFeed EwokFeed) ([]*EwokItem, string) {
	newFeedLastUpdated := newFeed.Updated
	if newFeedLastUpdated == "" && len(newFeed.Items) > 0 {
		newFeedLastUpdated = newFeed.Items[0].Published
	}
	newLastUpdatedTime := ParseTime(timeLayoutRSS, newFeedLastUpdated)
	oldLastUpdatedTime := ParseTime(timeLayoutPSQL, oldFeed.Updated)

	var newItems []*EwokItem

	if newLastUpdatedTime.After(oldLastUpdatedTime) {
		for _, item := range newFeed.Items {
			if item.PublishedParsed != nil && item.PublishedParsed.After(oldLastUpdatedTime) {
				newItems = append(newItems, &EwokItem{item, oldFeed.Id})
			}
		}
	}

	return newItems, newFeedLastUpdated
}
