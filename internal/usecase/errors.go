package usecase

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

type BusinessError struct {
	Code string
	Msg  string
}

func (e BusinessError) Error() string {
	if e.Msg == "" {
		return e.Code
	}
	return e.Msg
}
