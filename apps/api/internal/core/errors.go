package core

import "errors"

var (
	ErrItemNotFound = errors.New("NOT_FOUND")
	ErrInvalidID    = errors.New("INVALID_ID")
	ErrGenerateID   = errors.New("GENERATE_ID")
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrItemNotFound)
}
