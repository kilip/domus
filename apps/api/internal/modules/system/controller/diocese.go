package controller

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/system/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type DioceseService interface {
	GetAll(c context.Context) ([]model.DioceseResponse, error)
	GetByID(c context.Context, id int) (*model.DioceseResponse, error)
	Create(c context.Context, creatorID uuid.UUID, req model.CreateDioceseRequest) (*model.DioceseResponse, error)
	Update(c context.Context, updaterID uuid.UUID, id int, req model.UpdateDioceseRequest) (*model.DioceseResponse, error)
	Delete(c context.Context, id int) error
}

type DioceseController struct {
	dios DioceseService
}

func NewDioceseController(dios DioceseService) DioceseController {
	return DioceseController{dios}
}

func parseID(c fiber.Ctx) (int, error) {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, core.ErrInvalidID
	}
	return id, nil
}

// GetAll godoc
//
//	@Summary		Get all dioceses
//	@Description	Get all dioceses list
//	@Tags			Diocese
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	httpx.WebResponse[[]model.DioceseResponse]
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Router			/dioceses [get]
func (dc DioceseController) GetAll(c fiber.Ctx) error {
	data, err := dc.dios.GetAll(c)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// GetByID godoc
//
//	@Summary		Get diocese by ID
//	@Description	Get diocese details by ID
//	@Tags			Diocese
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Diocese ID"
//	@Success		200	{object}	httpx.WebResponse[model.DioceseResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/dioceses/{id} [get]
func (dc DioceseController) GetByID(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	data, err := dc.dios.GetByID(c, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Create godoc
//
//	@Summary		Create diocese
//	@Description	Create a new diocese
//	@Tags			Diocese
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateDioceseRequest	true	"Create diocese request"
//	@Success		201		{object}	httpx.WebResponse[model.DioceseResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/dioceses [post]
func (dc DioceseController) Create(c fiber.Ctx) error {
	var req model.CreateDioceseRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	user := core.UserFromContext(c)
	data, err := dc.dios.Create(c, user.ID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, data)
}

// Update godoc
//
//	@Summary		Update diocese
//	@Description	Update diocese by ID
//	@Tags			Diocese
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Diocese ID"
//	@Param			request	body		model.UpdateDioceseRequest	true	"Update diocese request"
//	@Success		200		{object}	httpx.WebResponse[model.DioceseResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/dioceses/{id} [put]
func (dc DioceseController) Update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req model.UpdateDioceseRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	user := core.UserFromContext(c)
	data, err := dc.dios.Update(c, user.ID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Delete godoc
//
//	@Summary		Delete diocese
//	@Description	Delete diocese by ID
//	@Tags			Diocese
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Diocese ID"
//	@Success		204
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/dioceses/{id} [delete]
func (dc DioceseController) Delete(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := dc.dios.Delete(c, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
