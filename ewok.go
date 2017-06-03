package main

import (
	"flag"
	"fmt"
	"github.com/paalka/ewok/config"
	"github.com/paalka/ewok/db"
	"github.com/paalka/ewok/feed"
	"html/template"
	"net/http"
	"strconv"
)

const (
	ITEMS_PER_PAGE = 10
)

func renderTemplate(w http.ResponseWriter, templates *template.Template, tmpl_name string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl_name+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeIndexHandler(config config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)

		feedItems := feed.GetAllFeedItems(db)
		feeds := feed.GetFeeds(db)
		renderTemplate(w, templates, "index", struct {
			FeedItems []feed.EwokItem
			Feeds     []feed.EwokFeed
		}{feedItems, feeds})
	}
}

func makePageHandler(config config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)
		possibleIndex := r.URL.Path[(len("/page/")):]

		if _, err := strconv.Atoi(possibleIndex); err != nil {
			http.Error(w, "Page not found!", http.StatusNotFound)
			return
		}

		index, err := strconv.ParseUint(possibleIndex, 10, 64)

		if err != nil {
			http.Error(w, "Page not found!", http.StatusNotFound)
			return
		}
		feedItems := feed.GetPaginatedFeeds(db, ITEMS_PER_PAGE, uint(index))
		feeds := feed.GetFeeds(db)
		renderTemplate(w, templates, "index", struct {
			FeedItems []feed.EwokItem
			Feeds     []feed.EwokFeed
		}{feedItems, feeds})
	}
}

func main() {
	portPtr := flag.Int("port", 8080, "The port to use when running the server")
	flag.Parse()
	config := config.LoadJsonConfig("config.json")
	templates := template.Must(template.ParseFiles("web/templates/index.html"))

	http.HandleFunc("/page/", makePageHandler(config, templates))
	http.HandleFunc("/", makeIndexHandler(config, templates))
	err := http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil)
	if err != nil {
		panic(err)
	}
}
