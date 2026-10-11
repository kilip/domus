package controller

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type ReportService interface {
	GetLedger(ctx context.Context, parishID int, query model.LedgerQuery) (model.LedgerResponse, string, error)
	GetTrialBalance(ctx context.Context, parishID int, query model.TrialBalanceQuery) (model.TrialBalanceResponse, error)
	GetCollections(ctx context.Context, parishID int, query model.CollectionsQuery) (model.CollectionsResponse, error)
}

type ReportController struct {
	svc ReportService
}

func NewReportController(svc ReportService) ReportController {
	return ReportController{svc: svc}
}

// GetLedger godoc
//
//	@Summary		Get account ledger
//	@Description	Get detailed transactions and running balance for an account
//	@Tags			Finance Report
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId	path		int		true	"Parish ID"
//	@Param			accountId	query		string	true	"Account ID (UUID)"
//	@Param			from		query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			to			query		string	true	"End date (YYYY-MM-DD)"
//	@Param			cursor		query		string	false	"Pagination cursor"
//	@Param			limit		query		int		false	"Items per page (default 50, max 200)"
//	@Success		200			{object}	httpx.WebResponse[model.LedgerResponse]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Failure		422			{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/reports/ledger [get]
func (rc ReportController) GetLedger(c fiber.Ctx) error {
	parishID, err := httpx.GetIntParam(c, "parishId")
	if err != nil {
		return err
	}
	var query model.LedgerQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}
	if accIDStr := c.Query("accountId"); accIDStr != "" {
		parsedID, err := uuid.Parse(accIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid accountId: "+err.Error())
		}
		query.AccountID = parsedID
	}
	if from := c.Query("from"); from != "" {
		query.From = from
	}
	if to := c.Query("to"); to != "" {
		query.To = to
	}
	if cursor := c.Query("cursor"); cursor != "" {
		query.Cursor = cursor
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = l
		}
	}

	res, nextCursor, err := rc.svc.GetLedger(c, parishID, query)
	if err != nil {
		return err
	}
	return httpx.WithCursor(c, res, nextCursor)
}

// GetTrialBalance godoc
//
//	@Summary		Get trial balance
//	@Description	Get trial balance report as of a specific date
//	@Tags			Finance Report
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId	path		int		true	"Parish ID"
//	@Param			asOf		query		string	true	"As of date (YYYY-MM-DD)"
//	@Success		200			{object}	httpx.WebResponse[model.TrialBalanceResponse]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Failure		422			{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/reports/trial-balance [get]
func (rc ReportController) GetTrialBalance(c fiber.Ctx) error {
	parishID, err := httpx.GetIntParam(c, "parishId")
	if err != nil {
		return err
	}
	var query model.TrialBalanceQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}
	if asOf := c.Query("asOf"); asOf != "" {
		query.AsOf = asOf
	}

	res, err := rc.svc.GetTrialBalance(c, parishID, query)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// GetCollections godoc
//
//	@Summary		Get collections report
//	@Description	Get summary of collections (kolekte) for income accounts within a date range
//	@Tags			Finance Report
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId		path		int		true	"Parish ID"
//	@Param			from			query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			to				query		string	true	"End date (YYYY-MM-DD)"
//	@Param			parentAccountId	query		string	false	"Filter by header account subtree (UUID)"
//	@Param			unitId			query		int		false	"Filter by transaction unit ID"
//	@Success		200				{object}	httpx.WebResponse[model.CollectionsResponse]
//	@Failure		400				{object}	httpx.ErrorResponse
//	@Failure		401				{object}	httpx.ErrorResponse
//	@Failure		403				{object}	httpx.ErrorResponse
//	@Failure		404				{object}	httpx.ErrorResponse
//	@Failure		422				{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/reports/collections [get]
func (rc ReportController) GetCollections(c fiber.Ctx) error {
	parishID, err := httpx.GetIntParam(c, "parishId")
	if err != nil {
		return err
	}
	var query model.CollectionsQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}
	if from := c.Query("from"); from != "" {
		query.From = from
	}
	if to := c.Query("to"); to != "" {
		query.To = to
	}
	if parentIDStr := c.Query("parentAccountId"); parentIDStr != "" {
		pID, err := uuid.Parse(parentIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid parentAccountId: "+err.Error())
		}
		query.ParentAccountID = &pID
	}
	if unitIDStr := c.Query("unitId"); unitIDStr != "" {
		uID, err := strconv.Atoi(unitIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid unitId: "+err.Error())
		}
		query.UnitID = &uID
	}

	res, err := rc.svc.GetCollections(c, parishID, query)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

