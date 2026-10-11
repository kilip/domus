package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
)

type ActivateService struct {
	entcli *ent.Client
	authz  *authz.Enforcer
	log    *slog.Logger
}

func NewActivateService(entcli *ent.Client, authz *authz.Enforcer, log *slog.Logger) ActivateService {
	return ActivateService{
		entcli: entcli,
		authz:  authz,
		log:    log,
	}
}

func (s ActivateService) Activate(ctx context.Context, userID uuid.UUID, req model.ActivateFinanceRequest) (*model.ActivateFinanceResponse, error) {
	// 1. Verify parish exists
	parishExists, err := s.entcli.Parish.Query().
		Where(parish.IDEQ(req.ParishID)).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !parishExists {
		return nil, model.ErrParishNotFound
	}

	// 2. Verify unit exists and belongs to the parish
	unitExists, err := s.entcli.Unit.Query().
		Where(
			entunit.IDEQ(req.UnitID),
			entunit.HasParishWith(parish.IDEQ(req.ParishID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !unitExists {
		return nil, model.ErrUnitNotFound
	}

	// 3. Verify user has finance permission in this unit
	scope := authz.Scope{Type: authz.ScopeUnit, ID: int64(req.UnitID)}
	can, err := s.authz.Can(ctx, userID.String(), scope, authz.ResourceFinance, authz.ActionRead)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, core.ErrForbidden
	}

	return &model.ActivateFinanceResponse{
		ActiveParishID: req.ParishID,
		ActiveUnitID:   req.UnitID,
	}, nil
}
