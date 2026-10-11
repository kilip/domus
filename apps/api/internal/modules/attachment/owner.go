package attachment

import (
	"context"

	"github.com/google/uuid"
)

// Owner is the record an attachment belongs to, as resolved by the owning
// module. WorkspaceID is the Casbin scope used to authorize access; because an
// owner's workspace is immutable it is safe to copy onto the attachment row.
type Owner struct {
	ID          uuid.UUID
	ParishID    int
	WorkspaceID int
}

// OwnerResolver loads the owner for (parishID, ownerID). It must return
// core.ErrItemNotFound when the owner does not exist or belongs to another
// parish, and a 422 error when the owner has no workspace.
type OwnerResolver func(ctx context.Context, parishID int, ownerID uuid.UUID) (*Owner, error)
