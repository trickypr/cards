package database

import (
	"database/sql"
	"log"
)

func ExecCrash(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		log.Print("DB error")
		log.Fatal(err)
	}
}

func InitializeDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "file:./db.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	ExecCrash(db, `
    CREATE TABLE IF NOT EXISTS deck (
      id TEXT PRIMARY KEY,
      name TEXT,
      description TEXT,
      side_one_lang TEXT,
      side_two_lang TEXT
    )
    `,
	)

	ExecCrash(db, `
    CREATE TABLE IF NOT EXISTS card (
      id TEXT PRIMARY KEY,
      deck TEXT NOT NULL,
      one TEXT,
      two TEXT,

      repition_number         INT     DEFAULT 0,
      easiness                FLOAT   DEFAULT 2.5,
      interrepition_interval  INT     DEFAULT 0,
      last_review             TEXT    DEFAULT '2013-10-07 08:23:19.120',
      last_quality            TINYINT DEFAULT 0,
      
      CONSTRAINT fk_deck FOREIGN KEY (deck) REFERENCES deck(id)
    )
    `,
	)

	return db
}
