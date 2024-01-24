package model

import (
	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
	Username string

	PasswordHash string
}

func (u *User) Create(db *sql.DB, password string) error {
	id, err := gonanoid.New()
	if err != nil {
		return err
	}
	u.ID = id

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)

	s, err := db.Prepare(`INSERT INTO user (id, username, password_hash) VALUES ($1, $2, $3) RETURNING id`)
	if err != nil {
		return err
	}
	return s.QueryRow(u.ID, u.Username, u.PasswordHash).Scan(&u.ID)
}

func (u *User) Exists(db *sql.DB) bool {
	row := db.QueryRow("SELECT COUNT(*) FROM user WHERE username = $1", u.Username)
	count := 0
	row.Scan(&count)
	return count != 0
}

func (u *User) GetFromId(db *sql.DB) error {
	s, err := db.Prepare(`SELECT username, password_hash FROM user WHERE id = $1`)
	if err != nil {
		return err
	}

	return s.QueryRow(u.ID).Scan(&u.Username, &u.PasswordHash)
}

func (u *User) GetFromUsername(db *sql.DB) error {
	s, err := db.Prepare(`SELECT id, password_hash FROM user WHERE username  = $1`)
	if err != nil {
		return err
	}

	return s.QueryRow(u.Username).Scan(&u.ID, &u.PasswordHash)
}

func (u *User) PasswordMatches(db *sql.DB, password string) (bool, error) {
	err := u.GetFromUsername(db)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil, nil
}

func UsernameIsValid(db *sql.DB, username string) string {
	if len(username) < 4 || len(username) > 20 {
		return "Your username is not between 4 and 20 characters"
	}

	row := db.QueryRow("SELECT COUNT(*) FROM user WHERE username = $1", username)
	count := 0
	row.Scan(&count)

	if count != 0 {
		return "Your username is taken"
	}

	return ""
}
