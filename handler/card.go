package handler

import (
	"cards/model"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

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

func HandleCardPut(db *sql.DB) http.HandlerFunc {
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

func HandleCardEvaluationPatch(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		quality, err := strconv.Atoi(query.Get("q"))
		last := query.Get("last") == "true"
		if err != nil {
			slog.Error("parsing quality", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cardid := chi.URLParam(r, "cardid")
		card, err := model.GetCard(db, cardid)
		if err != nil {
			slog.Error("fetching card", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = card.Learn(db, int8(quality))
		if err != nil {
			slog.Error("reviewing card", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !last {
			return
		}

		toLearn, err := model.CardsToLearn(db, card.Deck)
		if err != nil {
			slog.Error("fetching cards to learn", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl := TmplFiles("./templates/partials/card.htmx")
		if err := tmpl.ExecuteTemplate(w, "cards", toLearn); err != nil {
			slog.Error("rendering template", err)
		}

		if len(toLearn) == 0 {
			w.Header().Add("HX-Location", "/decks/"+card.Deck+"/review")
		}
	})
}

func HandleCardReviewPatch(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		quality, err := strconv.Atoi(query.Get("q"))
		last := query.Get("last") == "true"
		if err != nil {
			slog.Error("parsing quality", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cardid := chi.URLParam(r, "cardid")
		card, err := model.GetCard(db, cardid)
		if err != nil {
			slog.Error("fetching card", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = card.Review(db, int8(quality))
		if err != nil {
			slog.Error("reviewing card", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !last {
			return
		}

		toLearn, err := model.CardsToReview(db, card.Deck)
		if err != nil {
			slog.Error("fetching cards to learn", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl := TmplFiles("./templates/partials/card.htmx")
		if err := tmpl.ExecuteTemplate(w, "cards", toLearn); err != nil {
			slog.Error("rendering template", err)
		}

		if len(toLearn) == 0 {
			//		w.Header().Add("HX-Location", "/decks/"+card.Deck+"/review")
			slog.Info("todo: stats screen")
		}
	})
}
