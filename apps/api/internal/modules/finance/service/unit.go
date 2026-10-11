package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
)

type FinanceUnitRepository interface {
	GetAccessibleUnits(ctx context.Context, userID uuid.UUID, roles []string) ([]model.FinanceUnitResponse, error)
}

type FinanceUnitService struct {
	repo FinanceUnitRepository
	log  *slog.Logger
}

func NewFinanceUnitService(repo FinanceUnitRepository, log *slog.Logger) FinanceUnitService {
	return FinanceUnitService{
		repo: repo,
		log:  log,
	}
}

func (s FinanceUnitService) GetAccessibleUnits(ctx context.Context, userID uuid.UUID) ([]model.FinanceUnitResponse, error) {
	financeRoles := []string{
		authz.ModuleRole(authz.ResourceFinance, authz.LevelAdmin),
		authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter),
		authz.ModuleRole(authz.ResourceFinance, authz.LevelReader),
	}

	return s.repo.GetAccessibleUnits(ctx, userID, financeRoles)
}
