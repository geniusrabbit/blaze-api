package directives

import "errors"

var (
	ErrValueIsNil       = errors.New("value is nil")
	ErrValueIsEmpty     = errors.New("value is empty")
	ErrValueOutOfRange  = errors.New("value is out of range")
	ErrValueIsNotNumber = errors.New("value is not a number")
)
