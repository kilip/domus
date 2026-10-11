package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/membership"
	entworkspace "github.com/paroki/domus/api/ent/workspace"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type FinanceWorkspaceRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewFinanceWorkspaceRepository(entcli *ent.Client, log *slog.Logger) FinanceWorkspaceRepository {
	return FinanceWorkspaceRepository{
		entcli: entcli,
		log:    log,
	}
}

func (r FinanceWorkspaceRepository) GetAccessibleWorkspaces(ctx context.Context, userID uuid.UUID, roles []string) ([]model.FinanceWorkspaceResponse, error) {
	memberships, err := r.entcli.Membership.Query().
		Where(
			membership.UserIDEQ(userID.String()),
			membership.ScopeTypeEQ(membership.ScopeTypeWorkspace),
			membership.RoleIn(roles...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return []model.FinanceWorkspaceResponse{}, nil
	}

	workspaceIDs := make([]int, 0, len(memberships))
	for _, m := range memberships {
		workspaceIDs = append(workspaceIDs, int(m.ScopeID))
	}

	workspaces, err := r.entcli.Workspace.Query().
		Where(
			entworkspace.IDIn(workspaceIDs...),
			entworkspace.StatusEQ(entworkspace.StatusActive),
		).
		WithParish().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.FinanceWorkspaceResponse, 0, len(workspaces))
	for _, w := range workspaces {
		parishID := 0
		if w.Edges.Parish != nil {
			parishID = w.Edges.Parish.ID
		}
		result = append(result, model.FinanceWorkspaceResponse{
			ID:       w.ID,
			Name:     w.Name,
			ParishID: parishID,
		})
	}

	return result, nil
}
