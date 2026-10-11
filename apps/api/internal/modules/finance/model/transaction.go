package model

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
)

var (
	// TransactionNumberRegex matches valid custom transaction numbers (^[A-Za-z0-9._/-]{1,40}$).
	TransactionNumberRegex = regexp.MustCompile(`^[A-Za-z0-9._/-]{1,40}$`)
)

type PostingInput struct {
	AccountID uuid.UUID `json:"accountId" validate:"required"`
	Side      string    `json:"side" validate:"required,oneof=debit credit"`
	Amount    int64     `json:"amount" validate:"required,min=1,max=1000000000000000"`
	Memo      *string   `json:"memo,omitempty" validate:"omitempty,max=500"`
}

type CreateTransactionRequest struct {
	Number      *string        `json:"number,omitempty"`
	Date        string         `json:"date" validate:"required"`
	Kind        string         `json:"kind" validate:"required,oneof=income expense transfer adjustment"`
	Description *string        `json:"description,omitempty"`
	UnitID      *int           `json:"unitId,omitempty"`
	Postings    []PostingInput `json:"postings" validate:"required,min=2,max=200,dive"`
}

type SearchTransactionRequest struct {
	From      *string    `json:"from,omitempty"`
	To        *string    `json:"to,omitempty"`
	Kind      *string    `json:"kind,omitempty" validate:"omitempty,oneof=income expense transfer adjustment"`
	Status    *string    `json:"status,omitempty" validate:"omitempty,oneof=draft submitted approved posted void"`
	UnitID    *int       `json:"unitId,omitempty"`
	AccountID *uuid.UUID `json:"accountId,omitempty"`
	Q         *string    `json:"q,omitempty"`
	Cursor    *string    `json:"cursor,omitempty"`
	Limit     *int       `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

type PostingResponse struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"accountId"`
	AccountCode string    `json:"accountCode"`
	AccountName string    `json:"accountName"`
	Side        string    `json:"side"`
	Amount      int64     `json:"amount"`
	Memo        *string   `json:"memo,omitempty"`
}

type EntryResponse struct {
	ID          uuid.UUID         `json:"id"`
	Date        string            `json:"date"`
	Description *string           `json:"description,omitempty"`
	ReversesID  *uuid.UUID        `json:"reversesId,omitempty"`
	Postings    []PostingResponse `json:"postings"`
}

type TransactionResponse struct {
	ID          uuid.UUID            `json:"id"`
	Number      string               `json:"number"`
	Date        string               `json:"date"`
	Kind        string               `json:"kind"`
	Status      string               `json:"status"`
	Description *string              `json:"description,omitempty"`
	UnitID      *int                 `json:"unitId,omitempty"`
	Entries     []EntryResponse      `json:"entries"`
	Attachments []AttachmentResponse `json:"attachments"`
	CreatedBy   core.User            `json:"createdBy"`
	CreatedAt   time.Time            `json:"createdAt"`
}

type TransactionSummary struct {
	ID          uuid.UUID `json:"id"`
	Number      string    `json:"number"`
	Date        string    `json:"date"`
	Kind        string    `json:"kind"`
	Status      string    `json:"status"`
	Description *string   `json:"description,omitempty"`
	UnitID      *int      `json:"unitId,omitempty"`
	TotalAmount int64     `json:"totalAmount"`
}
