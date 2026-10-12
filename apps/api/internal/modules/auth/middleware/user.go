package middleware

import (
	"encoding/json"
	"slices"
	"strings"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
)

func UserInjector(issuer string, restAudience string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions || strings.HasPrefix(c.Path(), "/internal") {
			return c.Next()
		}
		token := jwtware.FromContext(c)
		if token == nil {
			return fiber.ErrUnauthorized
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.ErrUnauthorized
		}

		if iss, err := claims.GetIssuer(); err != nil || iss != issuer {
			return fiber.ErrUnauthorized
		}

		expectedAudience := restAudience

		if aud, err := claims.GetAudience(); err != nil || !slices.Contains(aud, expectedAudience) {
			return fiber.ErrUnauthorized
		}

		b, err := json.Marshal(claims)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		var user core.AuthenticatedUser
		if err := json.Unmarshal(b, &user); err != nil {
			return fiber.ErrInternalServerError
		}

		if sub, err := claims.GetSubject(); err == nil && sub != "" {
			if parsedSub, err := uuid.Parse(sub); err == nil {
				user.AuthID = &parsedSub
				if user.ID == uuid.Nil {
					user.ID = parsedSub
				}
			}
		}

		if user.AuthID == nil && user.ID != uuid.Nil {
			authID := user.ID
			user.AuthID = &authID
		}

		if ev, ok := claims["email_verified"].(bool); ok {
			user.EmailVerified = ev
		} else if ev, ok := claims["emailVerified"].(bool); ok {
			user.EmailVerified = ev
		}

		c.Locals(core.AUTH_USER_CONTEXT_KEY, user)

		return c.Next()
	}

}
