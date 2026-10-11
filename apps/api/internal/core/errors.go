package core

import "errors"

var (
	ErrItemNotFound = errors.New("NOT_FOUND")
	ErrInvalidID    = errors.New("INVALID_ID")
	ErrForbidden    = errors.New("FORBIDDEN")
	ErrUnauthorized = errors.New("UNAUTHORIZED")
	ErrInternal     = errors.New("INTERNAL")
	ErrUnknown      = errors.New("UNKNOWN")
	ErrGenerateID   = errors.New("GENERATE_ID")
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrItemNotFound)
}

type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

type ErrorItem struct {
	Status  int          `json:"status"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}
