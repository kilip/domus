package model

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

var (
	ErrTerritoryDifferentParish       = errors.New("TERRITORY_DIFFERENT_PARISH")
	ErrTerritorySelfParent            = errors.New("TERRITORY_SELF_PARENT")
	ErrTerritoryCycleDetected         = errors.New("TERRITORY_CYCLE_DETECTED")
	ErrTerritoryDeleteWithChildren    = errors.New("TERRITORY_DELETE_WITH_CHILDREN")
	ErrUnitRequiresTerritory          = errors.New("UNIT_REQUIRES_TERRITORY")
	ErrUnitNonTerritorialHasTerritory = errors.New("UNIT_NON_TERRITORIAL_HAS_TERRITORY")
	ErrUnitTerritoryDifferentParish   = errors.New("UNIT_TERRITORY_DIFFERENT_PARISH")
	ErrUnitParentDifferentParish      = errors.New("UNIT_PARENT_DIFFERENT_PARISH")
	ErrUnitSelfParent                 = errors.New("UNIT_SELF_PARENT")
	ErrUnitCycleDetected              = errors.New("UNIT_CYCLE_DETECTED")
	ErrUnitDeleteWithChildren         = errors.New("UNIT_DELETE_WITH_CHILDREN")
)

func init() {
	core.RegisterError(ErrTerritoryDifferentParish, fiber.StatusBadRequest, "parent territory belongs to a different parish")
	core.RegisterError(ErrTerritorySelfParent, fiber.StatusBadRequest, "territory cannot be its own parent")
	core.RegisterError(ErrTerritoryCycleDetected, fiber.StatusBadRequest, "cycle detected: territory cannot have its descendant as parent")
	core.RegisterError(ErrTerritoryDeleteWithChildren, fiber.StatusBadRequest, "cannot delete territory with children; deactivate it instead")
	core.RegisterError(ErrUnitRequiresTerritory, fiber.StatusBadRequest, "territorial unit requires a territory")
	core.RegisterError(ErrUnitNonTerritorialHasTerritory, fiber.StatusBadRequest, "non-territorial unit cannot have a territory")
	core.RegisterError(ErrUnitTerritoryDifferentParish, fiber.StatusBadRequest, "territory belongs to a different parish")
	core.RegisterError(ErrUnitParentDifferentParish, fiber.StatusBadRequest, "parent unit belongs to a different parish")
	core.RegisterError(ErrUnitSelfParent, fiber.StatusBadRequest, "unit cannot be its own parent")
	core.RegisterError(ErrUnitCycleDetected, fiber.StatusBadRequest, "cycle detected: unit cannot have its descendant as parent")
	core.RegisterError(ErrUnitDeleteWithChildren, fiber.StatusBadRequest, "cannot delete unit with children; deactivate it instead")
}
