package services

import (
	"errors"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrUnavailable  = errors.New("dates unavailable")
	ErrNotFound     = errors.New("record not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
)
