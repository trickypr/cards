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

func (c *Card) UpdateContents(db *sql.DB) error {
	s, err := db.Prepare(
		`
      UPDATE card
      SET one = $1, two = $2
      WHERE id = $3
      RETURNING id
    `,
	)
	if err != nil {
		return err
	}

	return s.QueryRow(c.One, c.Two, c.ID).Scan(&c.ID)
}

func GetCard(db *sql.DB, id string) (Card, error) {
	card := Card{}

	s, err := db.Prepare(
		`
      SELECT id, deck, one, two
      FROM card
      WHERE id = $1
    `,
	)
	if err != nil {
		return card, err
	}

	err = s.QueryRow(id).Scan(&card.ID, &card.Deck, &card.One, &card.Two)
	if err != nil {
		return card, err
	}

	return card, nil
}
