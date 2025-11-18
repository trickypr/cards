package model

import (
	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type ApiKey struct {
	ID     string
	Name   string
	Owner  string
	ApiKey string
}

func (a *ApiKey) Create(db *sql.DB) error {
	id, err := gonanoid.New(10)
	if err != nil {
		return err
	}

	a.ID = id

	apiKey, err := gonanoid.New(20)
	if err != nil {
		return err
	}

	a.ApiKey = apiKey

	s, err := db.Prepare(`INSERT INTO api_key (id, name, owner, api_key) VALUES ($1, $2, $3, $4) RETURNING id`)
	if err != nil {
		return err
	}

	return s.QueryRow(a.ID, a.Name, a.Owner, a.ApiKey).Scan(&a.ID)
}

func (a *ApiKey) Delete(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM api_key WHERE id = $1 RETURNING id`, a.ID)
	return err
}

func UserApiKeys(db *sql.DB, userid string) ([]ApiKey, error) {
	rows, err := db.Query("SELECT id, name, owner, api_key FROM api_key WHERE owner = $1", userid)
	if err != nil {
		return nil, err
	}

	apiKeys := []ApiKey{}

	for rows.Next() {
		apiKey := ApiKey{}

		if err := rows.Scan(&apiKey.ID, &apiKey.Name, &apiKey.Owner, &apiKey.ApiKey); err != nil {
			return apiKeys, err
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}
