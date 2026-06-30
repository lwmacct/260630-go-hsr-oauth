package repository

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("repository not found")

func WrapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
