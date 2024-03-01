package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Cards CardModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Cards: CardModel{DB: db},
	}
}
