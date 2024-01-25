package main

import (
	"cards/database"
	"cards/handler"
	"cards/model"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"

	_ "github.com/mattn/go-sqlite3"
	cp "github.com/otiai10/copy"
)

func main() {
	if _, err := os.Stat("templates"); errors.Is(err, os.ErrNotExist) {
		err := cp.Copy("/templates", "./templates/")
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Copied templates directory")
	}

	if _, err := os.Stat("static"); errors.Is(err, os.ErrNotExist) {
		err := cp.Copy("/static", "./static/")
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Copied static directory")
	}

	r := chi.NewRouter()

	db := database.InitializeDatabase()
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(tokenAuth))

	r.Route("/auth", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token, _, err := jwtauth.FromContext(r.Context())
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				if token != nil && jwt.Validate(token, tokenAuth.ValidateOptions()...) == nil {
					http.Redirect(w, r, "/decks", 303)
					return
				}

				next.ServeHTTP(w, r)
			})
		})

		r.Get("/login", handler.HandleLoginGet(db))
		r.Post("/login", handler.HandleLoginPost(db, tokenAuth))

		r.Get("/signup", handler.HandleSignupGet(db))
		r.Post("/signup", handler.HandleSignupPost(db, tokenAuth))
	})

	r.Route("/decks", func(r chi.Router) {
		r.Use(AlwaysHTML)
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/", handler.HandleDecksGet(db))
		r.Post("/", handler.HandleDecksPost(db))
		r.Get("/create", handler.HandleDecksCreateGet(db))
		r.Route("/{deckid}", func(r chi.Router) {
			r.Use(AuthenticatorDeck(db, tokenAuth))

			r.Get("/", handler.HandleDeckGet(db))
			r.Post("/", handler.HandleDeckPost(db))
			r.Get("/create", handler.HandleCreateCardGet(db))
			r.Get("/learn", handler.HandleLearnGet(db))
			r.Get("/review", handler.HandleReviewGet(db))
			r.Get("/complete", handler.HandleCompleteGet(db))

			r.Route("/card/{cardid}", func(r chi.Router) {
				r.Use(AuthenticatorCard(db, tokenAuth))
				r.Get("/", handler.HandleCardGet(db))
				r.Put("/", handler.HandleCardPut(db))
			})
		})
	})

	r.Route("/cards/{cardid}", func(r chi.Router) {
		r.Use(AlwaysHTML)
		r.Use(AuthenticatorCard(db, tokenAuth))

		r.Patch("/evaluation", handler.HandleCardEvaluationPatch(db))
		r.Patch("/review", handler.HandleCardReviewPatch(db))
	})

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}

func AlwaysHTML(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		h.ServeHTTP(w, r)
	})
}

func AuthenticatorCard(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			cardid := chi.URLParam(r, "cardid")
			token, data, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if token == nil || jwt.Validate(token, ja.ValidateOptions()...) != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if !model.IsCardOwner(db, data["id"].(string), cardid) {
				http.Error(w, "You do not own this card", http.StatusUnauthorized)
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func AuthenticatorDeck(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			deckid := chi.URLParam(r, "deckid")
			token, data, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if token == nil || jwt.Validate(token, ja.ValidateOptions()...) != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if !model.IsDeckOwner(db, data["id"].(string), deckid) {
				http.Error(w, "You do not own this deck", http.StatusUnauthorized)
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
