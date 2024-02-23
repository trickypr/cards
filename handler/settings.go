package handler

import (
	"cards/model"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type Settings struct {
	ApiKeys []model.ApiKey
}

func HandleSettingsGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, data, _ := jwtauth.FromContext(r.Context())

		settings := Settings{}

		apiKeys, err := model.UserApiKeys(db, data["id"].(string))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("fetching api keys", err)
			return
		}
		settings.ApiKeys = apiKeys

		tmpl := TmplFiles("./templates/base.htmx", "./templates/settings.htmx", "./templates/partials/apikeys.htmx")
		if err := tmpl.ExecuteTemplate(w, "base", settings); err != nil {
			slog.Error("rendering template", err)
			return
		}
	})
}

func HandleCreateApiKeys(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, data, _ := jwtauth.FromContext(r.Context())

		apikey := model.ApiKey{Owner: data["id"].(string), Name: r.FormValue("keyname")}
		err := apikey.Create(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("creating api key", err)
			return
		}

		tmpl := TmplFiles("./templates/partials/apikeys.htmx")
		if err := tmpl.ExecuteTemplate(w, "apikey", apikey); err != nil {
			slog.Error("rendering template", err)
			return
		}
	})
}

func HandleDeleteApiKey(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apikey := model.ApiKey{ID: chi.URLParam(r, "key")}
		err := apikey.Delete(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("deleting api key", err)
			return
		}
	})
}
