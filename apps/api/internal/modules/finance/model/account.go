package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateAccountRequest struct {
	Code        string     `json:"code" validate:"required,max=30"`
	Name        string     `json:"name" validate:"required,max=200"`
	Type        string     `json:"type" validate:"required,oneof=asset liability equity income expense"`
	ParentID    *uuid.UUID `json:"parentId,omitempty"`
	UnitID      int        `json:"unitId" validate:"required"`
	Postable    *bool      `json:"postable,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

type UpdateAccountRequest struct {
	Code        string     `json:"code" validate:"required,max=30"`
	Name        string     `json:"name" validate:"required,max=200"`
	Type        string     `json:"type" validate:"required,oneof=asset liability equity income expense"`
	ParentID    *uuid.UUID `json:"parentId,omitempty"`
	Postable    *bool      `json:"postable,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

type AccountResponse struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	NormalBalance string     `json:"normalBalance"`
	Postable      bool       `json:"postable"`
	Status        string     `json:"status"`
	Description   *string    `json:"description,omitempty"`
	ParishID      int        `json:"parishId"`
	ParentID      *uuid.UUID `json:"parentId,omitempty"`
	UnitID        int        `json:"unitId"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type ListAccountsQuery struct {
	Type     *string `query:"type" validate:"omitempty,oneof=asset liability equity income expense"`
	Status   *string `query:"status" validate:"omitempty,oneof=active inactive"`
	Postable *bool   `query:"postable"`
}

// DeriveNormalBalance returns "debit" for asset and expense, and "credit" for liability, equity, and income.
func DeriveNormalBalance(accountType string) string {
	switch accountType {
	case "asset", "expense":
		return "debit"
	case "liability", "equity", "income":
		return "credit"
	default:
		return "debit"
	}
}
