package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type UnitService interface {
	Search(c context.Context, req model.SearchUnitRequest) (*model.PaginatedResult[model.UnitResponse], error)
	GetByID(c context.Context, id int) (*model.UnitResponse, error)
	Create(c context.Context, req model.CreateUnitRequest) (*model.UnitResponse, error)
	Update(c context.Context, id int, req model.UpdateUnitRequest) (*model.UnitResponse, error)
	Delete(c context.Context, id int) error
}

type UnitController struct {
	svc UnitService
}

func NewUnitController(svc UnitService) UnitController {
	return UnitController{svc}
}

// Search godoc
//
//	@Summary		Search units
//	@Description	Search units with filters and cursor pagination
//	@Tags			Unit
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.SearchUnitRequest	true	"Search unit request"
//	@Success		200		{object}	httpx.WebResponse[[]model.UnitResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/units/search [post]
func (uc UnitController) Search(c fiber.Ctx) error {
	var req model.SearchUnitRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
	}
	res, err := uc.svc.Search(c, req)
	if err != nil {
		return err
	}
	return httpx.WithCursor(c, res.Data, res.NextCursor)
}

// GetByID godoc
//
//	@Summary		Get unit by ID
//	@Description	Get unit details by ID
//	@Tags			Unit
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Unit ID"
//	@Success		200	{object}	httpx.WebResponse[model.UnitResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/units/{id} [get]
func (uc UnitController) GetByID(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	data, err := uc.svc.GetByID(c, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Create godoc
//
//	@Summary		Create unit
//	@Description	Create a new unit
//	@Tags			Unit
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateUnitRequest	true	"Create unit request"
//	@Success		201		{object}	httpx.WebResponse[model.UnitResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/units [post]
func (uc UnitController) Create(c fiber.Ctx) error {
	var req model.CreateUnitRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := uc.svc.Create(c, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, data)
}

// Update godoc
//
//	@Summary		Update unit
//	@Description	Update unit by ID
//	@Tags			Unit
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Unit ID"
//	@Param			request	body		model.UpdateUnitRequest		true	"Update unit request"
//	@Success		200		{object}	httpx.WebResponse[model.UnitResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/units/{id} [put]
func (uc UnitController) Update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req model.UpdateUnitRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := uc.svc.Update(c, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Delete godoc
//
//	@Summary		Delete unit
//	@Description	Delete unit by ID
//	@Tags			Unit
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Unit ID"
//	@Success		204
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/units/{id} [delete]
func (uc UnitController) Delete(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := uc.svc.Delete(c, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

