// Package dberrors defines shared database error sentinels and error classification helpers.
package dberrors

import (
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested database query yields no rows
var ErrNotFound = errors.New("db: not found")

// Reports whether err is GORM's record-not-found error
func IsNoRows(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
