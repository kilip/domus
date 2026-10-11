package repository

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/core"
)

// EncodeCursor produces an opaque base64url-encoded string from date and UUID.
func EncodeCursor(date time.Time, id uuid.UUID) string {
	raw := date.UTC().Format(time.RFC3339Nano) + "|" + id.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes an opaque cursor back into date and UUID.
// If the cursor is invalid or malformed, it returns an HTTP 400 fiber.Error.
func DecodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	if cursor == "" {
		return time.Time{}, uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid cursor: empty cursor")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid cursor: base64 decode failed")
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid cursor: malformed structure")
	}

	date, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		date, err = time.Parse(time.RFC3339, parts[0])
		if err != nil {
			return time.Time{}, uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid cursor: invalid date component")
		}
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid cursor: invalid id component")
	}

	return date.UTC(), id, nil
}

func toCoreUser(u *ent.User) core.User {
	if u == nil {
		return core.User{}
	}
	return core.User{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Avatar: u.Avatar,
	}
}
