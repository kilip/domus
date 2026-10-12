package controller

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/user/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

// UserService defines the business logic methods exposed to UserController.
type UserService interface {
	Invite(ctx context.Context, caller core.AuthenticatedUser, req model.InviteUserRequest) (*model.UserInviteResponse, bool, error)
	Resolve(ctx context.Context, email string) (*model.ResolveUserResponse, error)
}

// UserController handles HTTP requests for user endpoints.
type UserController struct {
	svc      UserService
	validate *validator.Validate
}

// NewUserController constructs a new UserController.
func NewUserController(svc UserService) UserController {
	return UserController{
		svc:      svc,
		validate: validator.New(),
	}
}

// Invite godoc
//
//	@Summary		Invite user or add memberships
//	@Description	Invite a new user with initial memberships, or add memberships to an existing user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.InviteUserRequest	true	"Invite user request"
//	@Success		200		{object}	httpx.WebResponse[model.UserInviteResponse]
//	@Success		201		{object}	httpx.WebResponse[model.UserInviteResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/users [post]
func (uc UserController) Invite(c fiber.Ctx) error {
	var req model.InviteUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	if err := uc.validate.Struct(req); err != nil {
		return err
	}

	caller := core.UserFromContext(c)
	resp, isNew, err := uc.svc.Invite(c, caller, req)
	if err != nil {
		return err
	}

	if isNew {
		return httpx.Created(c, resp)
	}

	return httpx.OK(c, resp)
}

func (uc UserController) Resolve(c fiber.Ctx) error {
	var req model.ResolveUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	if err := uc.validate.Struct(req); err != nil {
		return err
	}

	resp, err := uc.svc.Resolve(c, req.Email)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
