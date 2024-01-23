package main

import (
	"cards/database"
	"cards/handler"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/brotli/go/cbrotli"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	r := chi.NewRouter()

	db := database.InitializeDatabase()

	compressor := middleware.NewCompressor(5, "/*")
	compressor.SetEncoder("br", func(w io.Writer, level int) io.Writer {
		return cbrotli.NewWriter(w, cbrotli.WriterOptions{
			Quality: level,
			LGWin:   0,
		})
	})
	r.Use(middleware.Compress(5))

	r.Route("/decks", func(r chi.Router) {
		r.Use(AlwaysHTML)

		r.Get("/", handler.HandleDecksGet(db))
		r.Post("/", handler.HandleDecksPost(db))
		r.Get("/create", handler.HandleDecksCreateGet(db))
		r.Route("/{deckid}", func(r chi.Router) {
			r.Get("/", handler.HandleDeckGet(db))
			r.Post("/", handler.HandleDeckPost(db))
			r.Get("/create", handler.HandleCreateCardGet(db))
			r.Get("/learn", handler.HandleLearnGet(db))
			r.Get("/review", handler.HandleReviewGet(db))
			r.Get("/complete", handler.HandleCompleteGet(db))

			r.Route("/card", func(r chi.Router) {
				r.Get("/{cardid}", handler.HandleCardGet(db))
				r.Put("/{cardid}", handler.HandleCardPut(db))
			})
		})
	})

	r.Route("/cards/{cardid}", func(r chi.Router) {
		r.Use(AlwaysHTML)

		r.Patch("/evaluation", handler.HandleCardEvaluationPatch(db))
		r.Patch("/review", handler.HandleCardReviewPatch(db))
	})

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":8080", r))
}

func AlwaysHTML(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		h.ServeHTTP(w, r)
	})
}
