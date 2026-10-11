package model

type ActivateFinanceRequest struct {
	ParishID int `json:"parishId" validate:"required,min=1"`
	UnitID   int `json:"unitId" validate:"required,min=1"`
}

type ActivateFinanceResponse struct {
	ActiveParishID int `json:"activeParishId"`
	ActiveUnitID   int `json:"activeUnitId"`
}
