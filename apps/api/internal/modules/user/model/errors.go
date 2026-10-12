package model

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

var (
	// ErrSuperadminNotAllowed indicates that granting the superadmin role via the invite endpoint is forbidden.
	ErrSuperadminNotAllowed = errors.New("SUPERADMIN_NOT_ALLOWED")
	// ErrScopeNotFound indicates that the requested scope target was not found in the database.
	ErrScopeNotFound = errors.New("SCOPE_NOT_FOUND")
	// ErrUserNotRegistered indicates that an email or user has not been invited or registered.
	ErrUserNotRegistered = errors.New("USER_NOT_REGISTERED")
)

func init() {
	core.RegisterError(ErrSuperadminNotAllowed, fiber.StatusForbidden, "granting superadmin role is not allowed")
	core.RegisterError(ErrScopeNotFound, fiber.StatusNotFound, "scope not found")
	core.RegisterError(ErrUserNotRegistered, fiber.StatusForbidden, "user is not registered")
}

