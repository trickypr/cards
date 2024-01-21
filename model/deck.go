package model

import (
	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Deck struct {
	ID          string
	Name        string
	Description string
	SideOneLang string
	SideTwoLang string
}

func (d *Deck) Create(db *sql.DB) error {
	id, err := gonanoid.New()
	d.ID = id

	s, err := db.Prepare(
		`
      INSERT INTO deck (id, name, description, side_one_lang, side_two_lang)
      VALUES ($1, $2, $3, $4, $5)
      RETURNING id
    `,
	)
	if err != nil {
		return err
	}

	return s.QueryRow(d.ID, d.Name, d.Description, d.SideOneLang, d.SideTwoLang).Scan()
}

func ReadAllDecks(db *sql.DB) ([]Deck, error) {
	decks := []Deck{}
	s, err := db.Prepare("SELECT id, name, description, side_one_lang, side_two_lang FROM deck")
	if err != nil {
		return decks, err
	}

	rows, err := s.Query()
	if err != nil {
		return decks, err
	}

	for rows.Next() {
		deck := Deck{}

		if err := rows.Scan(&deck.ID, &deck.Name, &deck.Description, &deck.SideOneLang, &deck.SideTwoLang); err != nil {
			return decks, err
		}

		decks = append(decks, deck)
	}

	return decks, nil
}
