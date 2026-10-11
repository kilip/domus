package attachment

import (
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
)

// Response is the public representation of an attachment. It is generic:
// OwnerType/OwnerID point at whatever record the file is attached to.
type Response struct {
	ID          uuid.UUID `json:"id"`
	OwnerType   string    `json:"ownerType"`
	OwnerID     uuid.UUID `json:"ownerId"`
	WorkspaceID int       `json:"workspaceId"`
	Filename    string    `json:"filename"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	CreatedBy   core.User `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DownloadResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}
