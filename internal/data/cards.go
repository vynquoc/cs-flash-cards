package data

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/vynquoc/cs-flash-cards/internal/validator"
)

type CodeSnippet struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type Card struct {
	ID             int64       `json:"id"`
	CreatedAt      time.Time   `json:"created_at"`
	Title          string      `json:"title"`
	Tags           []string    `json:"tags"`
	Content        string      `json:"content"`
	NextReviewDate time.Time   `json:"next_review_date"`
	CodeSnippet    CodeSnippet `json:"code_snippet"`
}

type CardModel struct {
	DB *sql.DB
}

func (c CodeSnippet) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CodeSnippet) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

func ValidateCard(v *validator.Validator, card *Card) {
	v.Check(card.Title != "", "title", "must be provided")
	v.Check(card.Content != "", "content", "must be provided")
	v.Check(card.Tags != nil, "tags", "must be provided")
	v.Check(len(card.Tags) >= 1, "tags", "must contain at least 1 tag")
	v.Check(len(card.Tags) <= 5, "tags", "must not contain more than 5 tags")
}

func (m CardModel) Insert(card *Card) error {
	query := `
		INSERT INTO cards (title, content, tags, next_review_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	args := []interface{}{card.Title, card.Content, pq.Array(card.Tags), card.NextReviewDate}
	if card.CodeSnippet.Code != "" && card.CodeSnippet.Language != "" {
		query = `
			INSERT INTO cards (title, content, tags, next_review_date, code_snippet)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at
		`
		args = append(args, card.CodeSnippet)
	}
	return m.DB.QueryRow(query, args...).Scan(&card.ID, &card.CreatedAt)
}

func (m CardModel) Get(id int64) (*Card, error) {
	return nil, nil
}

func (m CardModel) Update(card *Card) error {
	return nil
}

func (m CardModel) Delete(id int64) error {
	return nil
}
