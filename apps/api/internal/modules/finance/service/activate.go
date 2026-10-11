package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
)

type ActivateRepository interface {
	ParishExists(ctx context.Context, parishID int) (bool, error)
	UnitExistsInParish(ctx context.Context, unitID int, parishID int) (bool, error)
}

type ActivateService struct {
	repo  ActivateRepository
	authz *authz.Enforcer
	log   *slog.Logger
}

func NewActivateService(repo ActivateRepository, authz *authz.Enforcer, log *slog.Logger) ActivateService {
	return ActivateService{
		repo:  repo,
		authz: authz,
		log:   log,
	}
}

func (s ActivateService) Activate(ctx context.Context, userID uuid.UUID, req model.ActivateFinanceRequest) (*model.ActivateFinanceResponse, error) {
	// 1. Verify parish exists
	parishExists, err := s.repo.ParishExists(ctx, req.ParishID)
	if err != nil {
		return nil, err
	}
	if !parishExists {
		return nil, model.ErrParishNotFound
	}

	// 2. Verify unit exists and belongs to the parish
	unitExists, err := s.repo.UnitExistsInParish(ctx, req.UnitID, req.ParishID)
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
