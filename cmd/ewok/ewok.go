package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/paalka/ewok/pkg/config"
	"github.com/paalka/ewok/pkg/db"
	"github.com/paalka/ewok/pkg/feed"
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

func makeIndexHandler(config config.Config, templates *template.Template, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedItems, err := feed.GetAllFeedItems(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		feeds, err := feed.GetFeeds(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, f := range feeds {
			f.Updated = feed.ParseTime(feed.TimeLayout, f.Updated).Format(time.RFC1123)
		}
		renderTemplate(w, templates, "index", struct {
			FeedItems []feed.EwokItem
			Feeds     []feed.EwokFeed
		}{feedItems, feeds})
	}
}

func makePageHandler(config config.Config, templates *template.Template, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		possibleIndex := chi.URLParam(r, "paginationIndex")

		if _, err := strconv.Atoi(possibleIndex); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		index, err := strconv.ParseUint(possibleIndex, 10, 64)

		if err != nil {
			http.Error(w, "Page not found!", http.StatusNotFound)
			return
		}

		feedItems, err := feed.GetPaginatedFeeds(db, ITEMS_PER_PAGE, uint(index))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		feeds, err := feed.GetFeeds(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, f := range feeds {
			f.Updated = feed.ParseTime(feed.TimeLayout, f.Updated).Format(time.RFC1123)
		}
		renderTemplate(w, templates, "index", struct {
			FeedItems []feed.EwokItem
			Feeds     []feed.EwokFeed
		}{feedItems, feeds})
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
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

	db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)
	baseRouter.Get("/page/:paginationIndex", makePageHandler(config, templates, db))
	baseRouter.Get("/", makeIndexHandler(config, templates, db))

	FileServer(baseRouter, "/static/", http.Dir("web/static"))

	port := *portPtr
	log.Printf("Listening to localhost:%d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), baseRouter)
	if err != nil {
		panic(err)
	}
}
