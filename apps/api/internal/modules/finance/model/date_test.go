package model_test

import (
	"testing"
	"time"

	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "valid date",
			input:   "2026-10-11",
			want:    time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "valid leap year date",
			input:   "2024-02-29",
			want:    time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid month 13",
			input:   "2026-13-01",
			wantErr: true,
		},
		{
			name:    "invalid day February 30",
			input:   "2026-02-30",
			wantErr: true,
		},
		{
			name:    "invalid day non-leap year Feb 29",
			input:   "2025-02-29",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "with time part RFC3339",
			input:   "2026-10-11T00:00:00Z",
			wantErr: true,
		},
		{
			name:    "with time part space",
			input:   "2026-10-11 12:00:00",
			wantErr: true,
		},
		{
			name:    "invalid delimiter",
			input:   "2026/10/11",
			wantErr: true,
		},
		{
			name:    "partial date",
			input:   "2026-10",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, time.UTC, got.Location())
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	d := time.Date(2026, 10, 11, 15, 30, 0, 0, time.FixedZone("WITA", 8*3600))
	assert.Equal(t, "2026-10-11", model.FormatDate(d))
}
