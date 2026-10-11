package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidDate = errors.New("invalid date format: expected YYYY-MM-DD")
)

const DateFormat = "2006-01-02"

// ParseDate strictly parses a date string in YYYY-MM-DD format and returns midnight in UTC.
func ParseDate(s string) (time.Time, error) {
	if len(s) != 10 {
		return time.Time{}, ErrInvalidDate
	}
	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// FormatDate formats the time in UTC as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	return t.UTC().Format(DateFormat)
}
