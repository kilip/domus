package model

type ActivateFinanceRequest struct {
	ParishID int `json:"parishId" validate:"required,min=1"`
	UnitID   int `json:"unitId" validate:"required,min=1"`
}

type ActivateFinanceResponse struct {
	ActiveParishID int `json:"activeParishId"`
	ActiveUnitID   int `json:"activeUnitId"`
}

type FinanceUnitResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Code     *string `json:"code,omitempty"`
	Kind     string  `json:"kind"`
	ParishID int     `json:"parishId"`
}
