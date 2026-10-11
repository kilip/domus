package finance

import (
	"encoding/json"
	"net/url"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
)

const DefaultCookieName = "domus.fin"

func RequireFinanceContext(cookieName string) fiber.Handler {
	if cookieName == "" {
		cookieName = DefaultCookieName
	}

	return func(c fiber.Ctx) error {
		raw := c.Cookies(cookieName)
		if raw == "" {
			return ErrContextNotActive
		}

		// Support URL-encoded JSON or raw JSON in cookie
		unescaped, err := url.QueryUnescape(raw)
		if err != nil {
			unescaped = raw
		}

		var fc model.FinanceContext
		if err := json.Unmarshal([]byte(unescaped), &fc); err != nil {
			return ErrContextNotActive
		}

		if fc.ParishID <= 0 || fc.UnitID <= 0 {
			return ErrContextNotActive
		}

		c.Locals(model.FinanceContextKey, fc)
		c.SetContext(model.WithFinanceContext(c.Context(), fc))

		return c.Next()
	}
}

func FinanceParishScope(c fiber.Ctx) authz.Scope {
	fc, ok := model.FinanceContextFromContext(c)
	if !ok || fc.ParishID <= 0 {
		return authz.Scope{Type: authz.ScopeParish, ID: 0}
	}
	return authz.Scope{
		Type: authz.ScopeParish,
		ID:   int64(fc.ParishID),
	}
}
