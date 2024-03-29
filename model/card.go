package model

import (
	"database/sql"
	"math"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Card struct {
	ID         string
	Deck       string
	One        string
	Two        string
	LastReview string

	RepititionNumber      int
	InterrepitionInterval int
	Easiness              float32
	LastQuality           int8
}

func IsCardOwner(db *sql.DB, userid string, cardid string) bool {
	count := -1
	row := db.QueryRow("SELECT COUNT(*) FROM card WHERE id = $1 AND owner = $2", cardid, userid)
	row.Scan(&count)

	return count > 0
}

func (c *Card) Create(db *sql.DB, owner string) error {
	id, err := gonanoid.New()
	if err != nil {
		return err
	}

	c.ID = id

	s, err := db.Prepare(
		`
      INSERT INTO card (id, deck, one, two, owner)
      VALUES ($1, $2, $3, $4, $5)
      RETURNING id
    `,
	)
	if err != nil {
		return err
	}

	return s.QueryRow(c.ID, c.Deck, c.One, c.Two, owner).Scan(&c.ID)
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

func (c *Card) Review(db *sql.DB, quality int8) error {
	s, err := db.Prepare(`
    UPDATE card
    SET last_quality = $1, last_review = datetime('now')
    WHERE id = $2
    RETURNING last_review
  `)
	if err != nil {
		return err
	}

	c.LastQuality = quality
	return s.QueryRow(quality, c.ID).Scan(&c.LastReview)
}

func (c *Card) Learn(db *sql.DB, quality int8) error {
	// https://en.wikipedia.org/wiki/SuperMemo#Description_of_SM-2_algorithm
	c.LastQuality = quality

	if quality >= 3 {
		if c.RepititionNumber == 0 {
			c.InterrepitionInterval = 1
		} else if c.RepititionNumber == 1 {
			c.InterrepitionInterval = 6
		} else {
			c.InterrepitionInterval = int(math.Round(float64(c.InterrepitionInterval) * float64(c.Easiness)))
		}

		c.RepititionNumber += 1
	} else {
		c.RepititionNumber = 0
		// Deviation from algorithm
		if quality != 0 {
			c.InterrepitionInterval = 1
		} else {
			c.InterrepitionInterval = 0
		}
	}

	qf := float32(quality)
	c.Easiness = c.Easiness + (0.1 - (5.0-qf)*(0.08+(5.0-qf)*0.02))
	if c.Easiness < 1.3 {
		c.Easiness = 1.3
	}

	s, err := db.Prepare(`
    UPDATE card
    SET repition_number = $1, easiness = $2, interrepition_interval = $3, last_review = datetime('now'), last_quality = $4
    WHERE id = $5
    RETURNING last_review
  `)
	if err != nil {
		return err
	}

	return s.QueryRow(c.RepititionNumber, c.Easiness, c.InterrepitionInterval, quality, c.ID).Scan(&c.LastReview)
}

func GetCard(db *sql.DB, id string) (Card, error) {
	card := Card{}

	s, err := db.Prepare(
		`
      SELECT id, deck, one, two, repition_number, 
              easiness, interrepition_interval, last_review, last_quality
      FROM card
      WHERE id = $1
    `,
	)
	if err != nil {
		return card, err
	}

	err = s.QueryRow(id).Scan(&card.ID, &card.Deck, &card.One, &card.Two,
		&card.RepititionNumber, &card.Easiness, &card.InterrepitionInterval,
		&card.LastReview, &card.LastQuality)
	if err != nil {
		return card, err
	}

	return card, nil
}

func CardsToLearn(db *sql.DB, deckid string) ([]Card, error) {
	s, err := db.Prepare(`
    SELECT id, deck, one, two, repition_number, 
            easiness, interrepition_interval, last_review, last_quality
    FROM card
    WHERE deck = $1 
          AND (julianday("now") - julianday(last_review)) >= interrepition_interval
  `)
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(deckid)
	if err != nil {
		return nil, err
	}

	return CardsFromRows(rows)
}

func CardsToReview(db *sql.DB, deckid string) ([]Card, error) {
	s, err := db.Prepare(`
    SELECT id, deck, one, two, repition_number,
            easiness, interrepition_interval, last_review, last_quality
    FROM card
    WHERE deck = $1 
          AND last_quality < 4
  `)
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(deckid)
	if err != nil {
		return nil, err
	}

	return CardsFromRows(rows)
}

func CardsFromRows(rows *sql.Rows) ([]Card, error) {
	cards := []Card{}

	for rows.Next() {
		card := Card{}

		if err := rows.Scan(&card.ID, &card.Deck, &card.One, &card.Two,
			&card.RepititionNumber, &card.Easiness, &card.InterrepitionInterval,
			&card.LastReview, &card.LastQuality); err != nil {
			return cards, err
		}

		cards = append(cards, card)
	}

	return cards, nil
}
