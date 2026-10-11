package httpx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
)

// ScopeFunc resolves the casbin domain for a request.
type ScopeFunc func(c fiber.Ctx) authz.Scope

// SystemScope is for resources not owned by any diocese/parish/unit.
func SystemScope(fiber.Ctx) authz.Scope { return authz.System }

func RequirePermission(e *authz.Enforcer, scope ScopeFunc, obj authz.Resource, act authz.Action) fiber.Handler {
	return func(c fiber.Ctx) error {
		user := core.UserFromContext(c)

		ok, err := e.Can(c.Context(), user.ID.String(), scope(c), obj, act)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrForbidden
		}
		return c.Next()
	}
}
