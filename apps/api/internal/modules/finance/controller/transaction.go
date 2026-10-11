package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type TransactionService interface {
	Create(ctx context.Context, req model.CreateTransactionRequest) (*model.TransactionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.TransactionResponse, error)
	Search(ctx context.Context, req model.SearchTransactionRequest) (*model.PaginatedResult[model.TransactionSummary], error)
	Void(ctx context.Context, id uuid.UUID, req model.VoidTransactionRequest) (*model.TransactionResponse, error)
}

type TransactionController struct {
	svc TransactionService
}

func NewTransactionController(svc TransactionService) TransactionController {
	return TransactionController{svc: svc}
}

// Create godoc
//
//	@Summary		Create transaction
//	@Description	Create and post a new transaction with single-step journal entry
//	@Tags			Finance Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateTransactionRequest	true	"Create transaction request"
//	@Success		201		{object}	httpx.WebResponse[model.TransactionResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		409		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/fin/transactions [post]
func (tc TransactionController) Create(c fiber.Ctx) error {
	var req model.CreateTransactionRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	res, err := tc.svc.Create(c, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// Search godoc
//
//	@Summary		Search transactions
//	@Description	Search transactions with filters and cursor pagination
//	@Tags			Finance Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.SearchTransactionRequest	false	"Search transaction request"
//	@Success		200		{object}	httpx.WebResponse[[]model.TransactionSummary]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/fin/transactions/search [post]
func (tc TransactionController) Search(c fiber.Ctx) error {
	var req model.SearchTransactionRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
	}
	res, err := tc.svc.Search(c, req)
	if err != nil {
		return err
	}
	return httpx.WithCursor(c, res.Data, res.NextCursor)
}

// GetByID godoc
//
//	@Summary		Get transaction by ID
//	@Description	Get transaction details including journal entries and postings
//	@Tags			Finance Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Transaction ID (UUID)"
//	@Success		200	{object}	httpx.WebResponse[model.TransactionResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/fin/transactions/{id} [get]
func (tc TransactionController) GetByID(c fiber.Ctx) error {
	id, err := httpx.GetID(c)
	if err != nil {
		return err
	}
	res, err := tc.svc.GetByID(c, *id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Void godoc
//
//	@Summary		Void transaction
//	@Description	Void a posted transaction by creating a reversal entry and setting transaction status to void
//	@Tags			Finance Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Transaction ID (UUID)"
//	@Param			request	body		model.VoidTransactionRequest	false	"Void transaction request"
//	@Success		200		{object}	httpx.WebResponse[model.TransactionResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		409		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/fin/transactions/{id}/void [post]
func (tc TransactionController) Void(c fiber.Ctx) error {
	id, err := httpx.GetID(c)
	if err != nil {
		return err
	}
	var req model.VoidTransactionRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
	}
	res, err := tc.svc.Void(c, *id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}
