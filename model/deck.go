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
	id, err := gonanoid.New(10)
	if err != nil {
		return err
	}

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

	return s.QueryRow(d.ID, d.Name, d.Description, d.SideOneLang, d.SideTwoLang).Scan(&d.ID)
}

func GetDeck(db *sql.DB, id string) (Deck, error) {
	deck := Deck{}

	s, err := db.Prepare("SELECT id, name, description, side_one_lang, side_two_lang FROM deck WHERE id = $1")
	if err != nil {
		return deck, err
	}

	err = s.QueryRow(id).Scan(&deck.ID, &deck.Name, &deck.Description, &deck.SideOneLang, &deck.SideTwoLang)
	if err != nil {
		return deck, err
	}

	return deck, nil
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

func (d *Deck) Cards(db *sql.DB) ([]Card, error) {
	s, err := db.Prepare(`
    SELECT id, deck, one, two, repition_number, 
            easiness, interrepition_interval, last_review, last_quality
    FROM card
    WHERE deck = $1
  `)
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(&d.ID)
	if err != nil {
		return nil, err
	}

	return CardsFromRows(rows)
}
