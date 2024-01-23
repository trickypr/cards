package handler

import (
	"cards/model"
	"database/sql"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
)

func TmplFiles(files ...string) *template.Template {
	return template.Must(template.New("internal").Funcs(template.FuncMap{
		"arr": func(els ...any) []any {
			return els
		},
	}).ParseFiles(files...))
}

func HandleLearnGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deckid := chi.URLParam(r, "deckid")
		cards, err := model.CardsToLearn(db, deckid)
		if err != nil {
			slog.Error("fetching cards to learn", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(cards) == 0 {
			http.Redirect(w, r, "/decks/"+deckid+"/review", 307)
			return
		}

		tmpl := TmplFiles("./templates/base.htmx", "./templates/learn.htmx", "./templates/partials/card.htmx")
		if err := tmpl.ExecuteTemplate(w, "base", cards); err != nil {
			slog.Error("render template", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func HandleReviewGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deckid := chi.URLParam(r, "deckid")
		cards, err := model.CardsToReview(db, deckid)
		if err != nil {
			slog.Error("fetching cards to review", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// if len(cards) == 0 {
		// 	http.Redirect(w, r, "/decks/"+deckid+"/review", 307)
		// 	return
		// }

		tmpl := TmplFiles("./templates/base.htmx", "./templates/review.htmx", "./templates/partials/card.htmx")
		if err := tmpl.ExecuteTemplate(w, "base", cards); err != nil {
			slog.Error("render template", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
