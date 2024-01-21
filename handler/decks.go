package handler

import (
	"cards/model"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"
)

func RenderDeckList(db *sql.DB, w http.ResponseWriter) {
	decks, err := model.ReadAllDecks(db)
	if err != nil {
		slog.Error("fetching decks", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/base.htmx", "./templates/decks.htmx"))
	if err = tmpl.ExecuteTemplate(w, "base", decks); err != nil {
		slog.Error("executing templates", err)
	}
}

func HandleDeckGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RenderDeckList(db, w)
	})
}

func HandleDeckPost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := model.Deck{
			Name:        r.FormValue("Name"),
			Description: r.FormValue("Description"),
		}

		if err := d.Create(db); err != nil {
			slog.Error("creating deck", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RenderDeckList(db, w)
	})
}

func HandleDeckCreateGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(
			template.ParseFiles("./templates/base.htmx", "./templates/createdeck.htmx"),
		)

		if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
			slog.Error("executing template", err)
		}
	})
}
