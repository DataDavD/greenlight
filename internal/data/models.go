package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound is returned when a movie record doesn't exist in database.
var ErrRecordNotFound = errors.New("record not found")

// Models struct is a single convenient container to hold and represent all our database models.
type Models struct {
	Movies MovieModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
