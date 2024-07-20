package data

import (
	"context"
	"errors"
	"time"
)

func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)

}

// Errors
var (
	ErrDublicateCategoryTranslation = errors.New("duplicate category translation")
	ErrDuplicateEmail               = errors.New("duplicate email")
	ErrRecordNotFound               = errors.New("record not found")
	ErrEditConflict                 = errors.New("edit conflict")
	ErrInvalidRuntimeFormat         = errors.New("invalid runtime format")
	ErrCantDeleteDefaultCategory    = errors.New("can't delete default category")
)
