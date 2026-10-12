package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

type UserService interface {
	Validate(c context.Context, user core.AuthenticatedUser) error
}

func UserValidate(svc UserService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions || strings.HasPrefix(c.Path(), "/internal") {
			return c.Next()
		}
		user := core.UserFromContext(c)

		if err := svc.Validate(c, user); err != nil {
			return err
		}

		return c.Next()
	}
}
