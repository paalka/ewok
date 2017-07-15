package feed

import (
	"database/sql"
	"fmt"
	"github.com/bcampbell/fuzzytime"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"math"
	"strings"
	"time"
)

const TimeLayout string = time.RFC3339

type EwokItem struct {
	*gofeed.Item
	ParentFeedId   uint
	ParentFeedName string
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
	rows, err := db.Query("SELECT rss_item.title, link, description, publish_date, parent_feed_id, rss_feed.title FROM rss_item JOIN rss_feed ON rss_feed.id = rss_item.parent_feed_id ORDER BY publish_date DESC OFFSET $1 LIMIT $2", nItems*offset, nItems)

	if err != nil {
		return nil, err
	}

	var feedItems []EwokItem
	for rows.Next() {
		item := EwokItem{&gofeed.Item{}, 0, ""}
		err = rows.Scan(&item.Title, &item.Link, &item.Description, &item.Published, &item.ParentFeedId, &item.ParentFeedName)
		if err != nil {
			return nil, err
		}
		item.Published = ParseTime(TimeLayout, item.Published).Format(time.RFC1123)
		feedItems = append(feedItems, item)
	}

	return feedItems, nil
}

func GetAllFeedItems(db *sql.DB) ([]EwokItem, error) {
	rows, err := db.Query("SELECT rss_item.title, link, description, publish_date, parent_feed_id, rss_feed.title FROM rss_item JOIN rss_feed ON rss_feed.id = rss_item.parent_feed_id ORDER BY publish_date DESC")

	if err != nil {
		return nil, err
	}

	var feedItems []EwokItem
	for rows.Next() {
		item := EwokItem{&gofeed.Item{}, 0, ""}
		err = rows.Scan(&item.Title, &item.Link, &item.Description, &item.Published, &item.ParentFeedId, &item.ParentFeedName)
		if err != nil {
			return nil, err
		}
		item.Published = ParseTime(TimeLayout, item.Published).Format(time.RFC1123)
		feedItems = append(feedItems, item)
	}

	return feedItems, nil
}

func GetNumFeedPages(db *sql.DB, itemsPerPage uint) (*int, error) {
	row := db.QueryRow("SELECT count(*) FROM rss.rss_item")

	var nItems int
	err := row.Scan(&nItems)
	if err != nil {
		return nil, err
	}

	nPages := int(math.Ceil(float64(nItems) / float64(itemsPerPage)))

	return &nPages, nil
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

	newFeedLastUpdatedDT, _, err := fuzzytime.Extract(newFeedLastUpdated)
	if err != nil {
		fmt.Println(err)
	}
	newLastUpdatedTime := ParseTime(TimeLayout, newFeedLastUpdatedDT.ISOFormat())
	oldLastUpdatedTime := ParseTime(TimeLayout, oldFeed.Updated)

	var newItems []*EwokItem

	if newLastUpdatedTime.After(oldLastUpdatedTime) {
		for _, item := range newFeed.Items {
			if item.PublishedParsed != nil && item.PublishedParsed.After(oldLastUpdatedTime) {
				newItems = append(newItems, &EwokItem{item, oldFeed.Id, oldFeed.Title})
			}
		}
	}

	return newItems, newFeedLastUpdated
}
