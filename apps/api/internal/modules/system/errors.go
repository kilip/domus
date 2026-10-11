package system

import "github.com/paroki/domus/api/internal/modules/system/model"

var (
	ErrTerritoryDifferentParish    = model.ErrTerritoryDifferentParish
	ErrTerritorySelfParent         = model.ErrTerritorySelfParent
	ErrTerritoryCycleDetected      = model.ErrTerritoryCycleDetected
	ErrTerritoryDeleteWithChildren = model.ErrTerritoryDeleteWithChildren
)
