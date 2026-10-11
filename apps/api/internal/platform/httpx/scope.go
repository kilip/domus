package httpx

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
)

// ParishScope extracts the parish ID from route params and returns a Scope with Type ScopeParish.
func ParishScope(param string) ScopeFunc {
	return func(c fiber.Ctx) authz.Scope {
		id, err := strconv.ParseInt(c.Params(param), 10, 64)
		if err != nil || id <= 0 {
			return authz.Scope{Type: authz.ScopeParish, ID: 0}
		}
		return authz.Scope{
			Type: authz.ScopeParish,
			ID:   id,
		}
	}
}

// RequireIntParam validates that the specified route parameter is a positive integer (> 0).
// If invalid, zero, or negative, it returns core.ErrInvalidID.
func RequireIntParam(name string) fiber.Handler {
	return func(c fiber.Ctx) error {
		val, err := strconv.Atoi(c.Params(name))
		if err != nil || val <= 0 {
			return core.ErrInvalidID
		}
		return c.Next()
	}
}

// GetIntParam retrieves and validates a positive integer from route parameters.
func GetIntParam(c fiber.Ctx, name string) (int, error) {
	val, err := strconv.Atoi(c.Params(name))
	if err != nil || val <= 0 {
		return 0, core.ErrInvalidID
	}
	return val, nil
}
