package repository

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	entunit "github.com/paroki/domus/api/ent/unit"
)

type ActivateRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewActivateRepository(entcli *ent.Client, log *slog.Logger) ActivateRepository {
	return ActivateRepository{
		entcli: entcli,
		log:    log,
	}
}

func (r ActivateRepository) ParishExists(ctx context.Context, parishID int) (bool, error) {
	return r.entcli.Parish.Query().
		Where(parish.IDEQ(parishID)).
		Exist(ctx)
}

func (r ActivateRepository) UnitExistsInParish(ctx context.Context, unitID int, parishID int) (bool, error) {
	return r.entcli.Unit.Query().
		Where(
			entunit.IDEQ(unitID),
			entunit.HasParishWith(parish.IDEQ(parishID)),
		).
		Exist(ctx)
}

