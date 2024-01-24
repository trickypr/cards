package handler

import (
	"cards/model"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type DeckTemplateData struct {
	Deck  model.Deck
	Cards []model.Card
}

func RenderDeck(db *sql.DB, deck model.Deck, w http.ResponseWriter, r *http.Request) {
	cards, err := deck.Cards(db)
	if err != nil {
		slog.Error("fetching deck", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := DeckTemplateData{
		deck,
		cards,
	}

	tmpl := template.Must(template.ParseFiles("./templates/base.htmx", "./templates/deck.htmx"))
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("executing template", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandleDeckGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deckId := chi.URLParam(r, "deckid")
		deck, err := model.GetDeck(db, deckId)
		if err != nil {
			slog.Error("fetching deck", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RenderDeck(db, deck, w, r)
	})
}

func HandleDeckPost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deckId := chi.URLParam(r, "deckid")
		deck, err := model.GetDeck(db, deckId)
		if err != nil {
			slog.Error("fetching deck", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, data, _ := jwtauth.FromContext(r.Context())

		card := model.Card{
			Deck: deckId,
			One:  r.FormValue("One"),
			Two:  r.FormValue("Two"),
		}
		card.Create(db, data["id"].(string))

		RenderDeck(db, deck, w, r)
	})
}

func HandleCreateCardGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deckId := chi.URLParam(r, "deckid")
		deck, err := model.GetDeck(db, deckId)
		if err != nil {
			slog.Error("fetching deck", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("./templates/base.htmx", "./templates/createCard.htmx"))
		if err := tmpl.ExecuteTemplate(w, "base", deck); err != nil {
			slog.Error("executing template", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
