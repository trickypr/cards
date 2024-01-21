package database

import (
	"database/sql"
	"log"
)

func InitializeDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "file:./db.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`
    CREATE TABLE IF NOT EXISTS deck (
      id TEXT PRIMARY KEY,
      name TEXT,
      description TEXT,
      side_one_lang TEXT,
      side_two_lang TEXT
    )
    `,
	)

	if err != nil {
		log.Print("DB error")
		log.Fatal(err)
	}
	return db
}
