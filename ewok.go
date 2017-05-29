package main

import (
	"github.com/paalka/ewok/config"
	"github.com/paalka/ewok/db"
	"github.com/paalka/ewok/feed"
	"html/template"
	"net/http"
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
		renderTemplate(w, templates, "index", feedItems)
	}
}

func main() {
	config := config.LoadConfig("config.json")
	templates := template.Must(template.ParseFiles("templates/index.html"))

	http.HandleFunc("/", makeIndexHandler(config, templates))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}
