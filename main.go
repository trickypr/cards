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
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	gonanoid "github.com/matoous/go-nanoid/v2"

	_ "github.com/mattn/go-sqlite3"
	cp "github.com/otiai10/copy"
)

func main() {
	slog.Info("Main has run")

	if _, err := os.Stat("data"); errors.Is(err, os.ErrNotExist) {
		log.Fatal("You need to create a data folder!")
	}

	if _, err := os.Stat(".git"); errors.Is(err, os.ErrNotExist) {
		err := cp.Copy("/templates", "./templates/")
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Copied templates directory")

		err = cp.Copy("/static", "./static/")
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Copied static directory")
	} else {
		slog.Info("Skipping non-dev prep")
	}

	secret, err := os.ReadFile("./data/secret")
	if err != nil {
		// TODO: Better random generator
		slog.Info("Generating new secret")
		secret = []byte(gonanoid.Must(32))
		os.WriteFile("./data/secret", secret, 0666)
	}

	r := chi.NewRouter()

	db := database.InitializeDatabase()
	tokenAuth := jwtauth.New("HS256", secret, nil)

	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(tokenAuth))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/decks", 303)
	})

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

	r.Route("/settings", func(r chi.Router) {
		r.Use(AlwaysHTML)
		r.Use(Authenticator(db, tokenAuth))

		r.Get("/", handler.HandleSettingsGet(db))
		r.Post("/apikeys", handler.HandleCreateApiKeys(db))
		r.Delete("/apikeys/{key}", handler.HandleDeleteApiKey(db))
	})

	r.Route("/decks", func(r chi.Router) {
		r.Use(AlwaysHTML)
		r.Use(Authenticator(db, tokenAuth))

		r.Get("/", handler.HandleDecksGet(db))
		r.Post("/", handler.HandleDecksPost(db))
		r.Get("/create", handler.HandleDecksCreateGet(db))
		r.Route("/{deckid}", func(r chi.Router) {
			r.Use(AuthenticatorDeck(db, tokenAuth))

			r.Get("/", handler.HandleDeckGet(db))
			r.Post("/", handler.HandleDeckPost(db))
			r.Put("/", handler.HandleDeckPut(db))
			r.Get("/edit", handler.HandleDeckEditGet(db))
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

	r.Get("/favicon.ico", static("static/favicon.ico"))
	r.Get("/apple-touch-icon.png", static("static/icon512.png"))

	fs := http.FileServer(http.Dir("static/lib"))
	r.With(CacheFor(30)).Handle("/lib/*", http.StripPrefix("/lib/", fs))

	fs = http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	if _, err := os.Stat("full-cert.crt"); errors.Is(err, os.ErrNotExist) {
		slog.Info("Serve using http")
		log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
	} else {
		slog.Info("Serve using https")
		log.Fatal(http.ListenAndServeTLS("0.0.0.0:8080", "full-cert.crt", "private-key.key", r))
	}
}

func AlwaysHTML(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		h.ServeHTTP(w, r)
	})
}

func CacheFor(days int) func(http.Handler) http.Handler {
	seconds := days * 24 * 60 * 60
	ss := strconv.Itoa(seconds)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", "max-age="+ss+",s-maxage="+ss)
			h.ServeHTTP(w, r)
		})
	}
}

func static(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func AuthenticatorInternals(allowed func(db *sql.DB, r *http.Request, data map[string]interface{}) bool) func(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			hfn := func(w http.ResponseWriter, r *http.Request) {
				token, data, err := jwtauth.FromContext(r.Context())
				if err != nil || token == nil || jwt.Validate(token, ja.ValidateOptions()...) != nil {
					http.Redirect(w, r, "/auth/login", 303)
					return
				}

				if !allowed(db, r, data) {
					http.Error(w, "You do not own this item", http.StatusUnauthorized)
					return
				}

				// Token is authenticated, pass it through
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(hfn)
		}
	}
}

func Authenticator(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return AuthenticatorInternals(func(db *sql.DB, r *http.Request, data map[string]interface{}) bool {
		return model.UserExists(db, data["id"].(string))
	})(db, ja)
}

func AuthenticatorCard(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return AuthenticatorInternals(func(db *sql.DB, r *http.Request, data map[string]interface{}) bool {
		cardid := chi.URLParam(r, "cardid")
		return model.IsCardOwner(db, data["id"].(string), cardid)
	})(db, ja)
}

func AuthenticatorDeck(db *sql.DB, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return AuthenticatorInternals(func(db *sql.DB, r *http.Request, data map[string]interface{}) bool {
		deckid := chi.URLParam(r, "deckid")
		return model.IsDeckOwner(db, data["id"].(string), deckid)
	})(db, ja)
}
