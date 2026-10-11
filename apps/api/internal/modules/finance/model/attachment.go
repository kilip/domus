package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
)

type AttachmentResponse struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transactionId"`
	Filename      string    `json:"filename"`
	MimeType      string    `json:"mimeType"`
	Size          int64     `json:"size"`
	Checksum      string    `json:"checksum"`
	CreatedBy     core.User `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
}

type DownloadResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}
