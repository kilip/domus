package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type TerritoryService interface {
	Search(c context.Context, req model.SearchTerritoryRequest) (*model.PaginatedResult[model.TerritoryResponse], error)
	GetByID(c context.Context, id int) (*model.TerritoryResponse, error)
	Create(c context.Context, req model.CreateTerritoryRequest) (*model.TerritoryResponse, error)
	Update(c context.Context, id int, req model.UpdateTerritoryRequest) (*model.TerritoryResponse, error)
	Delete(c context.Context, id int) error
}

type TerritoryController struct {
	svc TerritoryService
}

func NewTerritoryController(svc TerritoryService) TerritoryController {
	return TerritoryController{svc}
}

// Search godoc
//
//	@Summary		Search territories
//	@Description	Search territories with filters and cursor pagination
//	@Tags			Territory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.SearchTerritoryRequest	true	"Search territory request"
//	@Success		200		{object}	httpx.WebResponse[[]model.TerritoryResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/territories/search [post]
func (tc TerritoryController) Search(c fiber.Ctx) error {
	var req model.SearchTerritoryRequest
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
//	@Summary		Get territory by ID
//	@Description	Get territory details by ID
//	@Tags			Territory
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Territory ID"
//	@Success		200	{object}	httpx.WebResponse[model.TerritoryResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/territories/{id} [get]
func (tc TerritoryController) GetByID(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	data, err := tc.svc.GetByID(c, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Create godoc
//
//	@Summary		Create territory
//	@Description	Create a new territory
//	@Tags			Territory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateTerritoryRequest	true	"Create territory request"
//	@Success		201		{object}	httpx.WebResponse[model.TerritoryResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/territories [post]
func (tc TerritoryController) Create(c fiber.Ctx) error {
	var req model.CreateTerritoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := tc.svc.Create(c, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, data)
}

// Update godoc
//
//	@Summary		Update territory
//	@Description	Update territory by ID
//	@Tags			Territory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Territory ID"
//	@Param			request	body		model.UpdateTerritoryRequest	true	"Update territory request"
//	@Success		200		{object}	httpx.WebResponse[model.TerritoryResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/territories/{id} [put]
func (tc TerritoryController) Update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req model.UpdateTerritoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := tc.svc.Update(c, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Delete godoc
//
//	@Summary		Delete territory
//	@Description	Delete territory by ID
//	@Tags			Territory
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Territory ID"
//	@Success		204
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/territories/{id} [delete]
func (tc TerritoryController) Delete(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := tc.svc.Delete(c, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
