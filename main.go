package main

import (
	"cards/database"
	"cards/handler"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	r := chi.NewRouter()

	db := database.InitializeDatabase()

	r.Route("/decks", func(r chi.Router) {
		r.Get("/", handler.HandleDecksGet(db))
		r.Post("/", handler.HandleDecksPost(db))
		r.Get("/create", handler.HandleDecksCreateGet(db))
		r.Route("/{deckid}", func(r chi.Router) {
			r.Get("/", handler.HandleDeckGet(db))
			r.Post("/", handler.HandleDeckPost(db))
			r.Get("/create", handler.HandleCreateCardGet(db))
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
