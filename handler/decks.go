package handler

import (
	"cards/model"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
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

func HandleDecksGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RenderDeckList(db, w)
	})
}

func HandleDecksPost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := model.Deck{
			Name:        r.FormValue("Name"),
			Description: r.FormValue("Description"),
		}

		_, data, _ := jwtauth.FromContext(r.Context())
		if err := d.Create(db, data["id"].(string)); err != nil {
			slog.Error("creating deck", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RenderDeckList(db, w)
	})
}

func HandleDecksCreateGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(
			template.ParseFiles("./templates/base.htmx", "./templates/createdeck.htmx"),
		)

		if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
			slog.Error("executing template", err)
		}
	})
}
