package model

import (
	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Card struct {
	ID   string
	Deck string
	One  string
	Two  string
}

func (c *Card) Create(db *sql.DB) error {
	id, err := gonanoid.New()
	if err != nil {
		return err
	}

	c.ID = id

	s, err := db.Prepare(
		`
      INSERT INTO card (id, deck, one, two)
      VALUES ($1, $2, $3, $4)
      RETURNING id
    `,
	)
	if err != nil {
		return err
	}

	return s.QueryRow(c.ID, c.Deck, c.One, c.Two).Scan(&c.ID)
}
