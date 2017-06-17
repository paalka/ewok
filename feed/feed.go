package feed

import (
	"database/sql"
	"fmt"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"strings"
	"time"
)

const TimeLayoutRSS string = time.RFC1123Z
const TimeLayoutPSQL string = time.RFC3339

type EwokItem struct {
	*gofeed.Item
	ParentFeedId uint
}

type EwokFeed struct {
	*gofeed.Feed
	Items []*EwokItem
	Id    uint
}

func UpdateFeedFromDiff(db *sql.DB, feedDiff EwokFeed) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	ins_stmt, err := db.Prepare("INSERT INTO rss_item (title, description, link, publish_date, parent_feed_id) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}

	for _, item := range feedDiff.Items {
		description := item.Description
		if len(description) > 350 {
			description = description[:350]

		}

		strippedText, err := html2text.FromString(description)

		if !strings.HasSuffix(strippedText, " […]") {
			strippedText = strippedText + " […]"
		}
		if err != nil {
			return err
		}

		_, err = tx.Stmt(ins_stmt).Exec(item.Title, strippedText, item.Link, item.Published, feedDiff.Id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	update_feed_stmt, err := db.Prepare("UPDATE rss_feed SET last_updated = $1 WHERE id = $2")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Stmt(update_feed_stmt).Exec(feedDiff.Updated, feedDiff.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func GetPaginatedFeeds(db *sql.DB, nItems uint, offset uint) ([]EwokItem, error) {
	rows, err := db.Query("SELECT title, link, description, publish_date, parent_feed_id FROM rss_item ORDER BY publish_date DESC OFFSET $1 LIMIT $2", nItems*offset, nItems)

	if err != nil {
		return nil, err
	}

	var feedItems []EwokItem
	for rows.Next() {
		item := EwokItem{&gofeed.Item{}, 0}
		err = rows.Scan(&item.Title, &item.Link, &item.Description, &item.Published, &item.ParentFeedId)
		if err != nil {
			return nil, err
		}
		item.Published = ParseTime(TimeLayoutPSQL, item.Published).Format(time.RFC1123)
		feedItems = append(feedItems, item)
	}

	return feedItems, nil
}

func GetAllFeedItems(db *sql.DB) ([]EwokItem, error) {
	rows, err := db.Query("SELECT title, link, description, publish_date, parent_feed_id FROM rss_item ORDER BY publish_date DESC")

	if err != nil {
		return nil, err
	}

	var feedItems []EwokItem
	for rows.Next() {
		item := EwokItem{&gofeed.Item{}, 0}
		err = rows.Scan(&item.Title, &item.Link, &item.Description, &item.Published, &item.ParentFeedId)
		if err != nil {
			return nil, err
		}
		item.Published = ParseTime(TimeLayoutPSQL, item.Published).Format(time.RFC1123)
		feedItems = append(feedItems, item)
	}

	return feedItems, nil
}

func GetFeeds(db *sql.DB) ([]EwokFeed, error) {
	rows, err := db.Query("SELECT id, title, url, last_updated FROM rss_feed")

	if err != nil {
		return nil, err
	}

	var feeds []EwokFeed
	for rows.Next() {
		var tmpItems []*EwokItem
		f := EwokFeed{&gofeed.Feed{}, tmpItems, 0}
		err = rows.Scan(&f.Id, &f.Title, &f.Link, &f.Updated)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}

	return feeds, nil
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
	newLastUpdatedTime := ParseTime(TimeLayoutRSS, newFeedLastUpdated)
	oldLastUpdatedTime := ParseTime(TimeLayoutPSQL, oldFeed.Updated)

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
