package model

import (
	"github.com/google/uuid"
)

type LedgerQuery struct {
	AccountID uuid.UUID `query:"accountId" validate:"required"`
	From      string    `query:"from" validate:"required"`
	To        string    `query:"to" validate:"required"`
	Cursor    string    `query:"cursor"`
	Limit     int       `query:"limit"`
}

type LedgerAccount struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	NormalBalance string    `json:"normalBalance"`
}

type LedgerRow struct {
	Date              string    `json:"date"`
	EntryID           uuid.UUID `json:"entryId"`
	PostingID         uuid.UUID `json:"postingId"`
	TransactionID     uuid.UUID `json:"transactionId"`
	TransactionNumber string    `json:"transactionNumber"`
	Description       *string   `json:"description,omitempty"`
	Debit             int64     `json:"debit"`
	Credit            int64     `json:"credit"`
	Balance           int64     `json:"balance"`
}

type LedgerResponse struct {
	Account        LedgerAccount `json:"account"`
	OpeningBalance int64         `json:"openingBalance"`
	Rows           []LedgerRow   `json:"rows"`
	ClosingBalance int64         `json:"closingBalance"`
}

type TrialBalanceQuery struct {
	AsOf string `query:"asOf" validate:"required"`
}

type TrialBalanceRow struct {
	AccountID uuid.UUID `json:"accountId"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Debit     int64     `json:"debit"`
	Credit    int64     `json:"credit"`
	Balance   int64     `json:"balance"`
}

type TrialBalanceTotals struct {
	Debit  int64 `json:"debit"`
	Credit int64 `json:"credit"`
}

type TrialBalanceResponse struct {
	AsOf   string             `json:"asOf"`
	Rows   []TrialBalanceRow  `json:"rows"`
	Totals TrialBalanceTotals `json:"totals"`
}

type CollectionsQuery struct {
	From            string     `query:"from" validate:"required"`
	To              string     `query:"to" validate:"required"`
	ParentAccountID *uuid.UUID `query:"parentAccountId"`
}

type CollectionsRow struct {
	AccountID uuid.UUID `json:"accountId"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Total     int64     `json:"total"`
}

type CollectionsResponse struct {
	From       string           `json:"from"`
	To         string           `json:"to"`
	Rows       []CollectionsRow `json:"rows"`
	GrandTotal int64            `json:"grandTotal"`
}
