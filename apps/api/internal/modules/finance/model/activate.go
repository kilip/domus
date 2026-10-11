package model

type ActivateFinanceRequest struct {
	ParishID    int `json:"parishId" validate:"required,min=1"`
	WorkspaceID int `json:"workspaceId" validate:"required,min=1"`
}

type ActivateFinanceResponse struct {
	ActiveParishID    int `json:"activeParishId"`
	ActiveWorkspaceID int `json:"activeWorkspaceId"`
}

type FinanceWorkspaceResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParishID int    `json:"parishId"`
}
