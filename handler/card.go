package handler

import (
	"cards/model"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleCardGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cardid := chi.URLParam(r, "cardid")
		card, err := model.GetCard(db, cardid)
		if err != nil {
			slog.Error("fetching card", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("./templates/base.htmx", "./templates/editcard.htmx"))
		if err := tmpl.ExecuteTemplate(w, "base", card); err != nil {
			slog.Error("execute template", err)
		}
	})
}

func HandleCartPut(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deckid := chi.URLParam(r, "deckid")
		cardid := chi.URLParam(r, "cardid")

		card := model.Card{
			ID:  cardid,
			One: r.FormValue("One"),
			Two: r.FormValue("Two"),
		}
		err := card.UpdateContents(db)
		if err != nil {
			slog.Error("updating card", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/decks/"+deckid, 303)
	})
}
