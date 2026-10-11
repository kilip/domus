package model

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

var (
	ErrTerritoryDifferentParish    = errors.New("TERRITORY_DIFFERENT_PARISH")
	ErrTerritorySelfParent         = errors.New("TERRITORY_SELF_PARENT")
	ErrTerritoryCycleDetected      = errors.New("TERRITORY_CYCLE_DETECTED")
	ErrTerritoryDeleteWithChildren = errors.New("TERRITORY_DELETE_WITH_CHILDREN")
)

func init() {
	core.RegisterError(ErrTerritoryDifferentParish, fiber.StatusBadRequest, "parent territory belongs to a different parish")
	core.RegisterError(ErrTerritorySelfParent, fiber.StatusBadRequest, "territory cannot be its own parent")
	core.RegisterError(ErrTerritoryCycleDetected, fiber.StatusBadRequest, "cycle detected: territory cannot have its descendant as parent")
	core.RegisterError(ErrTerritoryDeleteWithChildren, fiber.StatusBadRequest, "cannot delete territory with children; deactivate it instead")
}
