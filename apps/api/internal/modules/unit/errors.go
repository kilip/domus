package unit

import "github.com/paroki/domus/api/internal/modules/unit/model"

var (
	ErrTerritoryDifferentParish       = model.ErrTerritoryDifferentParish
	ErrTerritorySelfParent            = model.ErrTerritorySelfParent
	ErrTerritoryCycleDetected         = model.ErrTerritoryCycleDetected
	ErrTerritoryDeleteWithChildren    = model.ErrTerritoryDeleteWithChildren
	ErrUnitRequiresTerritory          = model.ErrUnitRequiresTerritory
	ErrUnitNonTerritorialHasTerritory = model.ErrUnitNonTerritorialHasTerritory
	ErrUnitTerritoryDifferentParish   = model.ErrUnitTerritoryDifferentParish
	ErrUnitParentDifferentParish      = model.ErrUnitParentDifferentParish
	ErrUnitSelfParent                 = model.ErrUnitSelfParent
	ErrUnitCycleDetected              = model.ErrUnitCycleDetected
	ErrUnitDeleteWithChildren         = model.ErrUnitDeleteWithChildren
)
