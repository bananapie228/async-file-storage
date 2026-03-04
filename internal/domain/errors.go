package domain

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input data")
	ErrInternal     = errors.New("internal server error")
)
