package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/membership"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type FinanceUnitRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewFinanceUnitRepository(entcli *ent.Client, log *slog.Logger) FinanceUnitRepository {
	return FinanceUnitRepository{
		entcli: entcli,
		log:    log,
	}
}

func (r FinanceUnitRepository) GetAccessibleUnits(ctx context.Context, userID uuid.UUID, roles []string) ([]model.FinanceUnitResponse, error) {
	memberships, err := r.entcli.Membership.Query().
		Where(
			membership.UserIDEQ(userID.String()),
			membership.ScopeTypeEQ(membership.ScopeTypeUnit),
			membership.RoleIn(roles...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return []model.FinanceUnitResponse{}, nil
	}

	unitIDs := make([]int, 0, len(memberships))
	for _, m := range memberships {
		unitIDs = append(unitIDs, int(m.ScopeID))
	}

	units, err := r.entcli.Unit.Query().
		Where(
			entunit.IDIn(unitIDs...),
			entunit.StatusEQ(entunit.StatusActive),
		).
		WithParish().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.FinanceUnitResponse, 0, len(units))
	for _, u := range units {
		parishID := 0
		if u.Edges.Parish != nil {
			parishID = u.Edges.Parish.ID
		}
		result = append(result, model.FinanceUnitResponse{
			ID:       u.ID,
			Name:     u.Name,
			Code:     u.Code,
			Kind:     string(u.Kind),
			ParishID: parishID,
		})
	}

	return result, nil
}

