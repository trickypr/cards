package handler

import (
	"cards/model"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

func HandleLoginGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		error := query.Get("error")

		tmpl := TmplFiles("./templates/base.htmx", "./templates/login.htmx")
		if err := tmpl.ExecuteTemplate(w, "base", error); err != nil {
			slog.Error("rendering template", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func HandleLoginPost(db *sql.DB, token *jwtauth.JWTAuth) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := r.FormValue("password")
		username := r.FormValue("username")

		user := model.User{
			Username: username,
		}

		if !user.Exists(db) {
			http.Redirect(w, r, "/auth/login?error=Username or password was incorrect", 303)
		}

		valid_pass, err := user.PasswordMatches(db, password)
		if err != nil {
			slog.Error("creating user", err)
			http.Redirect(w, r, "/auth/login?error=An unknown internal error occured", 303)
			return
		}

		if !valid_pass {
			http.Redirect(w, r, "/auth/login?error=password was incorrect", 303)
			return
		}

		_, jwt, _ := token.Encode(map[string]interface{}{"id": user.ID, "username": user.Username})
		http.SetCookie(w, &http.Cookie{Name: "jwt", Value: jwt, Path: "/"})
		http.Redirect(w, r, "/decks", 303)
	})
}

func HandleSignupGet(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		error := query.Get("error")

		tmpl := TmplFiles("./templates/base.htmx", "./templates/signup.htmx")
		if err := tmpl.ExecuteTemplate(w, "base", error); err != nil {
			slog.Error("rendering template", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func HandleSignupPost(db *sql.DB, token *jwtauth.JWTAuth) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password1 := r.FormValue("password1")
		password2 := r.FormValue("password2")
		username := r.FormValue("username")

		usernameErr := model.UsernameIsValid(db, username)
		if usernameErr != "" {
			http.Redirect(w, r, "/auth/signup?error="+usernameErr, 303)
			return
		}

		if password1 != password2 {
			http.Redirect(w, r, "/auth/signup?error=Your passwords did not match", 303)
			return
		}

		if len(password1) < 6 {
			http.Redirect(w, r, "/auth/signup?error=Make sure your password is longer than 6 characters", 303)
			return
		}

		if len([]byte(password1)) > 72 {
			http.Redirect(w, r, "/auth/signup?error=Your password cannot be longer than 72 bytes", 303)
			return
		}

		user := model.User{
			Username: username,
		}
		err := user.Create(db, password1)
		if err != nil {
			slog.Error("creating user", err)
			http.Redirect(w, r, "/auth/signup?error=An unknown internal error occured", 303)
			return
		}

		_, jwt, _ := token.Encode(map[string]interface{}{"id": user.ID, "username": user.Username})
		http.SetCookie(w, &http.Cookie{Name: "jwt", Value: jwt, Path: "/"})
		http.Redirect(w, r, "/decks", 303)
	})
}
