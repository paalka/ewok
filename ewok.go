package main

import (
	"github.com/paalka/ewok/config"
	"github.com/paalka/ewok/db"
	"github.com/paalka/ewok/feed"
	"html/template"
	"net/http"
)

func makeIndexHandler(config config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := db.GetDatabaseConnection(config.DB_NAME, config.DB_USER, config.DB_PASS)

		feedItems := feed.GetAllFeedItems(db)

		err := templates.ExecuteTemplate(w, "index.html", feedItems)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func main() {
	config := config.LoadConfig("config.json")
	templates := template.Must(template.ParseFiles("templates/index.html"))

	http.HandleFunc("/", makeIndexHandler(config, templates))
	http.ListenAndServe(":8000", nil)
}
