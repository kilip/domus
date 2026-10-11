package repository

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	entworkspace "github.com/paroki/domus/api/ent/workspace"
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

func (r ActivateRepository) WorkspaceExistsInParish(ctx context.Context, workspaceID int, parishID int) (bool, error) {
	return r.entcli.Workspace.Query().
		Where(
			entworkspace.IDEQ(workspaceID),
			entworkspace.HasParishWith(parish.IDEQ(parishID)),
		).
		Exist(ctx)
}
