package dberrors

import (
	"errors"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a lookup yields no rows.
var ErrNotFound = errors.New("db: not found")

// IsNoRows reports whether err is GORM's record-not-found error.
func IsNoRows(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}