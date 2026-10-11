package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type ParishService interface {
	GetAll(c context.Context) ([]model.ParishResponse, error)
	GetByID(c context.Context, id int) (*model.ParishResponse, error)
	Create(c context.Context, req model.CreateParishRequest) (*model.ParishResponse, error)
	Update(c context.Context, id int, req model.UpdateParishRequest) (*model.ParishResponse, error)
	Delete(c context.Context, id int) error
}

type ParishController struct {
	svc ParishService
}

func NewParishController(svc ParishService) ParishController {
	return ParishController{svc}
}

// GetAll godoc
//
//	@Summary		Get all parishes
//	@Description	Get all parishes list
//	@Tags			Parish
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	httpx.WebResponse[[]model.ParishResponse]
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Router			/parishes [get]
func (pc ParishController) GetAll(c fiber.Ctx) error {
	data, err := pc.svc.GetAll(c)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// GetByID godoc
//
//	@Summary		Get parish by ID
//	@Description	Get parish details by ID
//	@Tags			Parish
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Parish ID"
//	@Success		200	{object}	httpx.WebResponse[model.ParishResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/parishes/{id} [get]
func (pc ParishController) GetByID(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	data, err := pc.svc.GetByID(c, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Create godoc
//
//	@Summary		Create parish
//	@Description	Create a new parish
//	@Tags			Parish
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateParishRequest	true	"Create parish request"
//	@Success		201		{object}	httpx.WebResponse[model.ParishResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/parishes [post]
func (pc ParishController) Create(c fiber.Ctx) error {
	var req model.CreateParishRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := pc.svc.Create(c, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, data)
}

// Update godoc
//
//	@Summary		Update parish
//	@Description	Update parish by ID
//	@Tags			Parish
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Parish ID"
//	@Param			request	body		model.UpdateParishRequest	true	"Update parish request"
//	@Success		200		{object}	httpx.WebResponse[model.ParishResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/parishes/{id} [put]
func (pc ParishController) Update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req model.UpdateParishRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := pc.svc.Update(c, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Delete godoc
//
//	@Summary		Delete parish
//	@Description	Delete parish by ID
//	@Tags			Parish
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Parish ID"
//	@Success		204
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/parishes/{id} [delete]
func (pc ParishController) Delete(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := pc.svc.Delete(c, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
