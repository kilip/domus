package model

import (
	"time"

	"github.com/google/uuid"
)

// MembershipInput represents a requested scope membership for a user.
type MembershipInput struct {
	// ScopeType defines the scope hierarchy level: system, diocese, parish, or workspace.
	ScopeType string `json:"scopeType" validate:"required,oneof=system diocese parish workspace"`
	// ScopeID is the numeric identifier of the scope (0 for system).
	ScopeID int64 `json:"scopeId" validate:"gte=0"`
	// Role is the role name to assign within the scope.
	Role string `json:"role" validate:"required"`
}

// InviteUserRequest represents the payload for inviting a user or updating memberships via POST /users.
type InviteUserRequest struct {
	// Email is the unique email address of the user.
	Email string `json:"email" validate:"required,email"`
	// Name is the display name of the user.
	Name string `json:"name" validate:"required"`
	// Memberships contains one or more initial scope memberships.
	Memberships []MembershipInput `json:"memberships" validate:"required,min=1,dive"`
}

// MembershipResponse represents a membership record in the response payload.
type MembershipResponse struct {
	// ID is the database record identifier for the membership.
	ID int `json:"id"`
	// ScopeType is the scope level: system, diocese, parish, or workspace.
	ScopeType string `json:"scopeType"`
	// ScopeID is the numeric identifier of the scope.
	ScopeID int64 `json:"scopeId"`
	// Role is the assigned role name.
	Role string `json:"role"`
	// CreatedAt is the timestamp when the membership was created.
	CreatedAt time.Time `json:"createdAt"`
}

// UserInviteResponse represents the user and membership details returned by the invite endpoint.
type UserInviteResponse struct {
	// ID is the unique identifier (UUID v7) of the user.
	ID uuid.UUID `json:"id"`
	// Name is the display name of the user.
	Name string `json:"name"`
	// Email is the lowercase email of the user.
	Email string `json:"email"`
	// Status is the user activation status (e.g., invited, active).
	Status string `json:"status"`
	// Memberships is the list of memberships associated with the user.
	Memberships []MembershipResponse `json:"memberships"`
}

