package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type AccountService interface {
	List(ctx context.Context, filter model.ListAccountsQuery) ([]model.AccountResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.AccountResponse, error)
	Create(ctx context.Context, req model.CreateAccountRequest) (*model.AccountResponse, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateAccountRequest) (*model.AccountResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AccountController struct {
	svc AccountService
}

func NewAccountController(svc AccountService) AccountController {
	return AccountController{svc: svc}
}

// Create godoc
//
//	@Summary		Create account
//	@Description	Create a new account in chart of accounts
//	@Tags			Finance Account
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateAccountRequest	true	"Create account request"
//	@Success		201		{object}	httpx.WebResponse[model.AccountResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		409		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/fin/accounts [post]
func (ac AccountController) Create(c fiber.Ctx) error {
	var req model.CreateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	res, err := ac.svc.Create(c, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// List godoc
//
//	@Summary		List accounts
//	@Description	List accounts for active parish and unit with optional filters, sorted by code
//	@Tags			Finance Account
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type		query		string	false	"Account type"	Enums(asset, liability, equity, income, expense)
//	@Param			status		query		string	false	"Status"		Enums(active, inactive)
//	@Param			postable	query		bool	false	"Postable"
//	@Success		200			{object}	httpx.WebResponse[[]model.AccountResponse]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Failure		422			{object}	httpx.ErrorResponse
//	@Router			/fin/accounts [get]
func (ac AccountController) List(c fiber.Ctx) error {
	var query model.ListAccountsQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}
	if accType := c.Query("type"); accType != "" {
		query.Type = &accType
	}
	if status := c.Query("status"); status != "" {
		query.Status = &status
	}
	if postableStr := c.Query("postable"); postableStr != "" {
		postable := postableStr == "true" || postableStr == "1"
		query.Postable = &postable
	}
	res, err := ac.svc.List(c, query)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// GetByID godoc
//
//	@Summary		Get account by ID
//	@Description	Get account details by ID
//	@Tags			Finance Account
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Account ID (UUID)"
//	@Success		200	{object}	httpx.WebResponse[model.AccountResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/fin/accounts/{id} [get]
func (ac AccountController) GetByID(c fiber.Ctx) error {
	id, err := httpx.GetID(c)
	if err != nil {
		return err
	}
	res, err := ac.svc.GetByID(c, *id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Update godoc
//
//	@Summary		Update account
//	@Description	Update an existing account
//	@Tags			Finance Account
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Account ID (UUID)"
//	@Param			request	body		model.UpdateAccountRequest	true	"Update account request"
//	@Success		200		{object}	httpx.WebResponse[model.AccountResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		409		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/fin/accounts/{id} [put]
func (ac AccountController) Update(c fiber.Ctx) error {
	id, err := httpx.GetID(c)
	if err != nil {
		return err
	}
	var req model.UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	res, err := ac.svc.Update(c, *id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Delete godoc
//
//	@Summary		Delete account
//	@Description	Soft delete an account (blocked if it has children or postings)
//	@Tags			Finance Account
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Account ID (UUID)"
//	@Success		204	"No Content"
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/fin/accounts/{id} [delete]
func (ac AccountController) Delete(c fiber.Ctx) error {
	id, err := httpx.GetID(c)
	if err != nil {
		return err
	}
	if err := ac.svc.Delete(c, *id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
