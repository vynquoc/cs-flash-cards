package data

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/vynquoc/cs-flash-cards/internal/validator"
)

type CodeSnippet map[string]interface{}

type Card struct {
	ID             int64        `json:"id"`
	CreatedAt      time.Time    `json:"created_at"`
	Title          string       `json:"title"`
	Tags           []string     `json:"tags"`
	Content        string       `json:"content"`
	NextReviewDate time.Time    `json:"next_review_date"`
	CodeSnippet    *CodeSnippet `json:"code_snippet"`
	Description    string       `json:"description"`
}

type CardModel struct {
	DB *sql.DB
}

func (c CodeSnippet) Value() (driver.Value, error) {
	j, err := json.Marshal(c)
	return j, err
}

func (c *CodeSnippet) Scan(src interface{}) error {

	source, ok := src.([]byte)
	nullSrc := string(source)
	if nullSrc == "null" {
		*c = nil
		return nil
	}
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*c, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("type assertion .(map[string]interface{}) failed")
	}

	return nil
}

func ValidateCard(v *validator.Validator, card *Card) {
	v.Check(card.Title != "", "title", "must be provided")
	v.Check(card.Content != "", "content", "must be provided")
	v.Check(card.Tags != nil, "tags", "must be provided")
	v.Check(len(card.Tags) >= 1, "tags", "must contain at least 1 tag")
	v.Check(len(card.Tags) <= 5, "tags", "must not contain more than 5 tags")
}

func (c CardModel) Insert(card *Card) error {
	query := `
			INSERT INTO cards (title, content, tags, next_review_date, code_snippet, description)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at
		`
	args := []interface{}{card.Title, card.Content, pq.Array(card.Tags), card.NextReviewDate, card.CodeSnippet, card.Description}
	return c.DB.QueryRow(query, args...).Scan(&card.ID, &card.CreatedAt)
}

func (c CardModel) Get(id int64) (*Card, error) {
	query := `
		SELECT id, content, title, tags, code_snippet, created_at, next_review_date, description
		FROM cards
		WHERE id = $1
	`
	var card Card
	var s sql.NullString
	err := c.DB.QueryRow(query, id).Scan(
		&card.ID,
		&card.Content,
		&card.Title,
		pq.Array(&card.Tags),
		&card.CodeSnippet,
		&card.CreatedAt,
		&card.NextReviewDate,
		&s,
	)
	if s.Valid {
		card.Description = s.String
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &card, nil
}

func (c CardModel) Update(card *Card) error {
	query := `
		UPDATE cards
		SET title = $1, content = $2, tags = $3, code_snippet = $4, next_review_date = $5, description = $6
		WHERE id = $7
		RETURNING id
	`
	args := []interface{}{
		card.Title,
		card.Content,
		pq.Array(&card.Tags),
		card.CodeSnippet,
		card.NextReviewDate,
		card.Description,
		card.ID,
	}
	err := c.DB.QueryRow(query, args...).Scan(&card.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}

func (c CardModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM cards
		WHERE id = $1
	`

	result, err := c.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (c CardModel) GetAll(title string, tags []string, filters Filters) ([]*Card, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, next_review_date, tags, content, code_snippet, description
		FROM cards
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (tags @> $2 or $2 = '{}')
		ORDER BY %s %s, created_at DESC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())
	args := []interface{}{title, pq.Array(tags), filters.limit(), filters.offset()}
	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	cards := []*Card{}
	totalRecords := 0

	for rows.Next() {
		var card Card
		var s sql.NullString
		err := rows.Scan(
			&totalRecords,
			&card.ID,
			&card.CreatedAt,
			&card.Title,
			&card.NextReviewDate,
			pq.Array(&card.Tags),
			&card.Content,
			&card.CodeSnippet,
			&s,
		)
		if s.Valid {
			card.Description = s.String
		}
		if err != nil {
			return nil, Metadata{}, err
		}
		cards = append(cards, &card)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return cards, metadata, nil
}

func (c CardModel) GetReviewCards() ([]*Card, error) {
	query := `
		SELECT id, created_at, title, next_review_date, tags, content, code_snippet, description
		FROM cards
		WHERE next_review_date <= CURRENT_DATE
	`
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, err
	}
	cards := []*Card{}
	var s sql.NullString
	for rows.Next() {
		var card Card
		err := rows.Scan(
			&card.ID,
			&card.CreatedAt,
			&card.Title,
			&card.NextReviewDate,
			pq.Array(&card.Tags),
			&card.Content,
			&card.CodeSnippet,
			&s,
		)
		if err != nil {
			return nil, err
		}
		if s.Valid {
			card.Description = s.String
		}
		cards = append(cards, &card)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}

func (c CardModel) GetRandomCard() (*Card, error) {
	query := `
		SELECT id, content, title, tags, code_snippet, created_at, next_review_date, description
		FROM cards
		ORDER BY RANDOM() 
		LIMIT 1
	`
	var card Card
	var s sql.NullString
	err := c.DB.QueryRow(query).Scan(
		&card.ID,
		&card.Content,
		&card.Title,
		pq.Array(&card.Tags),
		&card.CodeSnippet,
		&card.CreatedAt,
		&card.NextReviewDate,
		&s,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	if s.Valid {
		card.Description = s.String
	}

	return &card, nil

}
