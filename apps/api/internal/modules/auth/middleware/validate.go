package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
)

type UserService interface {
	Validate(c context.Context, user core.AuthenticatedUser) (*model.UserResponse, error)
}

func UserValidate(svc UserService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions || strings.HasPrefix(c.Path(), "/internal") {
			return c.Next()
		}
		user := core.UserFromContext(c)

		dbUser, err := svc.Validate(c, user)
		if err != nil {
			return err
		}

		user.ID = dbUser.ID
		user.Name = dbUser.Name
		user.Email = dbUser.Email
		if dbUser.Avatar != "" {
			user.Avatar = dbUser.Avatar
		}
		c.Locals(core.AUTH_USER_CONTEXT_KEY, user)

		return c.Next()
	}
}
