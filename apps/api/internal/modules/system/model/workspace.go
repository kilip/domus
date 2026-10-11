package model

type WorkspaceResponse struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Status      string          `json:"status"`
	ParishID    int             `json:"parishId"`
	Parish      *ParishResponse `json:"parish,omitempty"`
}

type SearchWorkspaceRequest struct {
	ParishID *int    `json:"parishId,omitempty"`
	Status   *string `json:"status,omitempty"`
	Cursor   *string `json:"cursor,omitempty"`
	Limit    *int    `json:"limit,omitempty"`
}

type CreateWorkspaceRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	ParishID    int     `json:"parishId" validate:"required"`
}

type UpdateWorkspaceRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	ParishID    int     `json:"parishId" validate:"required"`
}
