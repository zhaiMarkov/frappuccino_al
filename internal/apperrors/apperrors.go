package apperrors

import "errors"

// Самые встречаемые проблемы
var (
	ErrInvalidInput     = errors.New("Error: invalid input: missing required field")
	ErrExistConflict    = errors.New("Error: already exist")
	ErrNotExistConflict = errors.New("Error: doesn't exist")
	ErrOrderClosed      = errors.New("Error: the order is already closed")
)
