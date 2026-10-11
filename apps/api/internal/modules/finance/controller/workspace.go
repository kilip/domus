package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/service"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type FinanceWorkspaceController struct {
	svc service.FinanceWorkspaceService
}

func NewFinanceWorkspaceController(svc service.FinanceWorkspaceService) FinanceWorkspaceController {
	return FinanceWorkspaceController{svc: svc}
}

// GetAccessibleWorkspaces godoc
//
//	@Summary		Get accessible finance workspaces
//	@Description	Get list of workspaces accessible to the current user for finance module
//	@Tags			Finance
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	httpx.WebResponse[[]model.FinanceWorkspaceResponse]
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Router			/fin/workspaces [get]
func (wc FinanceWorkspaceController) GetAccessibleWorkspaces(c fiber.Ctx) error {
	user := core.UserFromContext(c)
	if user.ID == uuid.Nil {
		return core.ErrUnauthorized
	}

	var workspaces []model.FinanceWorkspaceResponse
	workspaces, err := wc.svc.GetAccessibleWorkspaces(c.Context(), user.ID)
	if err != nil {
		return err
	}

	return httpx.OK(c, workspaces)
}
