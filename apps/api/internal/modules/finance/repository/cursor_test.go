package repository_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorRoundTrip(t *testing.T) {
	now := time.Date(2026, 10, 11, 14, 30, 15, 123456000, time.UTC)
	id := uuid.New()

	cursor := repository.EncodeCursor(now, id)
	require.NotEmpty(t, cursor)

	gotDate, gotID, err := repository.DecodeCursor(cursor)
	require.NoError(t, err)
	assert.Equal(t, now, gotDate)
	assert.Equal(t, id, gotID)
}

func TestDecodeCursor_Errors(t *testing.T) {
	tests := []struct {
		name   string
		cursor string
	}{
		{
			name:   "empty string",
			cursor: "",
		},
		{
			name:   "invalid base64",
			cursor: "not-valid-base64!@#$",
		},
		{
			name:   "no delimiter",
			cursor: base64.RawURLEncoding.EncodeToString([]byte("2026-10-11T00:00:00Z")),
		},
		{
			name:   "tampered invalid date",
			cursor: base64.RawURLEncoding.EncodeToString([]byte("not-a-date|" + uuid.New().String())),
		},
		{
			name:   "tampered invalid UUID",
			cursor: base64.RawURLEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339) + "|not-a-uuid")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := repository.DecodeCursor(tt.cursor)
			assert.Error(t, err)
		})
	}
}
