package model

type UnitResponse struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Code        *string            `json:"code,omitempty"`
	Kind        string             `json:"kind"`
	Description *string            `json:"description,omitempty"`
	Status      string             `json:"status"`
	ParishID    int                `json:"parishId"`
	ParentID    *int               `json:"parentId,omitempty"`
	TerritoryID *int               `json:"territoryId,omitempty"`
	Parish      *ParishResponse    `json:"parish,omitempty"`
	Parent      *UnitResponse      `json:"parent,omitempty"`
	Territory   *TerritoryResponse `json:"territory,omitempty"`
}

type SearchUnitRequest struct {
	ParishID    *int    `json:"parishId,omitempty"`
	ParentID    *int    `json:"parentId,omitempty"`
	TerritoryID *int    `json:"territoryId,omitempty"`
	Kind        *string `json:"kind,omitempty"`
	Status      *string `json:"status,omitempty"`
	Cursor      *string `json:"cursor,omitempty"`
	Limit       *int    `json:"limit,omitempty"`
}

type CreateUnitRequest struct {
	Name        string  `json:"name" validate:"required"`
	Code        *string `json:"code,omitempty"`
	Kind        string  `json:"kind" validate:"required,oneof=council territorial category section committee"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	ParishID    int     `json:"parishId" validate:"required"`
	ParentID    *int    `json:"parentId,omitempty"`
	TerritoryID *int    `json:"territoryId,omitempty"`
}

type UpdateUnitRequest struct {
	Name        string  `json:"name" validate:"required"`
	Code        *string `json:"code,omitempty"`
	Kind        string  `json:"kind" validate:"required,oneof=council territorial category section committee"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	ParishID    int     `json:"parishId" validate:"required"`
	ParentID    *int    `json:"parentId,omitempty"`
	TerritoryID *int    `json:"territoryId,omitempty"`
}
