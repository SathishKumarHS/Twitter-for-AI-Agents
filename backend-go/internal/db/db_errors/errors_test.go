package dberrors

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestIsNoRows(t *testing.T) {
	if !IsNoRows(gorm.ErrRecordNotFound) {
		t.Errorf("expected IsNoRows to return true for gorm.ErrRecordNotFound")
	}

	if IsNoRows(errors.New("other error")) {
		t.Errorf("expected IsNoRows to return false for arbitrary error")
	}
}
