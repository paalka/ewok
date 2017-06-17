package main

import (
	"flag"
	"fmt"
	"github.com/paalka/ewok/config"
	"github.com/paalka/ewok/db"
	"github.com/paalka/ewok/feed"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"html/template"
	"net/http"
	"strconv"
	"time"
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
		for _, f := range feeds {
			f.Updated = feed.ParseTime(feed.TimeLayoutPSQL, f.Updated).Format(time.RFC1123)
		}
		renderTemplate(w, templates, "index", struct {
			FeedItems []feed.EwokItem
			Feeds     []feed.EwokFeed
		}{feedItems, feeds})
	}
}

func makePageHandler(config config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		possibleIndex := chi.URLParam(r, "paginationIndex")
		db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)

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
		for _, f := range feeds {
			f.Updated = feed.ParseTime(feed.TimeLayoutPSQL, f.Updated).Format(time.RFC1123)
		}
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
	baseRouter := chi.NewRouter()

	baseRouter.Use(middleware.RequestID)
	baseRouter.Use(middleware.RealIP)
	baseRouter.Use(middleware.Logger)
	baseRouter.Use(middleware.Recoverer)
	baseRouter.Use(middleware.CloseNotify)
	baseRouter.Use(middleware.Timeout(60 * time.Second))

	baseRouter.Get("/page/:paginationIndex", makePageHandler(config, templates))
	baseRouter.Get("/", makeIndexHandler(config, templates))
	baseRouter.FileServer("/static/", http.Dir("web/static"))
	err := http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), baseRouter)
	if err != nil {
		panic(err)
	}
}
