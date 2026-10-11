package controller

import (
	"encoding/json"
	"net/url"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/service"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type ActivateController struct {
	svc        service.ActivateService
	cookieName string
}

func NewActivateController(svc service.ActivateService, cookieName string) ActivateController {
	if cookieName == "" {
		cookieName = "domus.fin"
	}
	return ActivateController{
		svc:        svc,
		cookieName: cookieName,
	}
}

// Activate sets the active parish and workspace context in the cookie.
//
//	@Summary		Activate finance context
//	@Description	Set active parish and workspace for finance module in user cookie
//	@Tags			Finance
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.ActivateFinanceRequest	true	"Activate finance request"
//	@Success		200		{object}	httpx.WebResponse[model.ActivateFinanceResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Router			/fin/activate [post]
func (ac ActivateController) Activate(c fiber.Ctx) error {
	user := core.UserFromContext(c)
	if user.ID == uuid.Nil {
		return core.ErrUnauthorized
	}

	var req model.ActivateFinanceRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	res, err := ac.svc.Activate(c, user.ID, req)
	if err != nil {
		return err
	}

	cookieBytes, err := json.Marshal(res)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = ac.cookieName
	cookie.Value = url.QueryEscape(string(cookieBytes))
	cookie.Path = "/"
	cookie.MaxAge = 20 * 365 * 24 * 3600 // 20 years persistent cookie
	c.Cookie(cookie)

	return httpx.OK(c, res)
}
