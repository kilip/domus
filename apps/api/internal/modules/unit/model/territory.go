package model

type TerritoryResponse struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Code        *string            `json:"code,omitempty"`
	Kind        string             `json:"kind"`
	Description *string            `json:"description,omitempty"`
	Status      string             `json:"status"`
	ParishID    int                `json:"parishId"`
	ParentID    *int               `json:"parentId,omitempty"`
	UnitID      *int               `json:"unitId,omitempty"`
	Parish      *ParishResponse    `json:"parish,omitempty"`
	Parent      *TerritoryResponse `json:"parent,omitempty"`
}

type SearchTerritoryRequest struct {
	ParishID *int    `json:"parishId,omitempty"`
	ParentID *int    `json:"parentId,omitempty"`
	Kind     *string `json:"kind,omitempty"`
	Status   *string `json:"status,omitempty"`
	Cursor   *string `json:"cursor,omitempty"`
	Limit    *int    `json:"limit,omitempty"`
}

type CreateTerritoryRequest struct {
	Name        string  `json:"name" validate:"required"`
	Code        *string `json:"code,omitempty"`
	Kind        string  `json:"kind" validate:"required"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	ParishID    int     `json:"parishId" validate:"required"`
	ParentID    *int    `json:"parentId,omitempty"`
}

type UpdateTerritoryRequest struct {
	Name        string  `json:"name" validate:"required"`
	Code        *string `json:"code,omitempty"`
	Kind        string  `json:"kind" validate:"required"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	ParishID    int     `json:"parishId" validate:"required"`
	ParentID    *int    `json:"parentId,omitempty"`
}
