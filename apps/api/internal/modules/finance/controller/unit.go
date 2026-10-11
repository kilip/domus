package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/service"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type FinanceUnitController struct {
	svc service.FinanceUnitService
}

func NewFinanceUnitController(svc service.FinanceUnitService) FinanceUnitController {
	return FinanceUnitController{svc: svc}
}

// GetAccessibleUnits godoc
//
//	@Summary		Get accessible finance units
//	@Description	Get list of units accessible to the current user for finance module
//	@Tags			Finance
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	httpx.WebResponse[[]model.FinanceUnitResponse]
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Router			/fin/units [get]
func (uc FinanceUnitController) GetAccessibleUnits(c fiber.Ctx) error {
	user := core.UserFromContext(c)
	if user.ID == uuid.Nil {
		return core.ErrUnauthorized
	}

	var units []model.FinanceUnitResponse
	units, err := uc.svc.GetAccessibleUnits(c.Context(), user.ID)
	if err != nil {
		return err
	}

	return httpx.OK(c, units)
}
