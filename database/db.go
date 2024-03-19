package database

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
)

func ExecCrash(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		log.Print("DB error")
		log.Fatal(err)
	}
}

func InitializeDatabase() *sql.DB {
	if _, err := os.Stat("./data/db.sqlite"); errors.Is(err, os.ErrNotExist) {
		os.Create("./data/db.sqlite")
		slog.Info("Created database file")
	}

	db, err := sql.Open("sqlite3", "file:./db.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	ExecCrash(db, `
    CREATE TABLE IF NOT EXISTS deck (
      id TEXT PRIMARY KEY,
      owner TEXT NOT NULL,
      name TEXT,
      description TEXT,
      side_one_lang TEXT,
      side_two_lang TEXT,

      CONSTRAINT fk_owner FOREIGN KEY (owner) REFERENCES user(id)
    )
    `,
	)

	ExecCrash(db, `
    CREATE TABLE IF NOT EXISTS card (
      id TEXT PRIMARY KEY,
      deck TEXT NOT NULL,
      owner TEXT NOT NULL,
      one TEXT,
      two TEXT,

      repition_number         INT     DEFAULT 0,
      easiness                FLOAT   DEFAULT 2.5,
      interrepition_interval  INT     DEFAULT 0,
      last_review             TEXT    DEFAULT '2013-10-07 08:23:19.120',
      last_quality            TINYINT DEFAULT 0,
      
      CONSTRAINT fk_deck FOREIGN KEY (deck) REFERENCES deck(id),
      CONSTRAINT fk_owner FOREIGN KEY (owner) REFERENCES user(id)
    )
    `,
	)

	ExecCrash(db, `
    CREATE TABLE IF NOT EXISTS user (
      id       TEXT PRIMARY KEY,
      username TEXT NOT NULL,

      password_hash TEXT NOT NULL
    )
  `)

	ExecCrash(db, `
    CREATE TABLE IF NOT EXISTS api_key (
      id      TEXT PRIMARY KEY,
      name    TEXT NOT NULL,
      owner   TEXT NOT NULL,
      api_key TEXT NOT NULL,

      CONSTRAINT fk_owner FOREIGN KEY (owner) REFERENCES user(id)
    )
  `)

	// Clean up any left over empty cards & decks
	// TODO: Remove in the future
	ExecCrash(db, `
    DELETE FROM card WHERE trim(one) = '' AND trim(two) = ''
  `)
	ExecCrash(db, `
    DELETE FROM deck WHERE trim(name) = '' AND trim(description) = ''
  `)

	return db
}
