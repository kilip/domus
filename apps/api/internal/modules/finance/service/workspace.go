package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
)

type FinanceWorkspaceRepository interface {
	GetAccessibleWorkspaces(ctx context.Context, userID uuid.UUID, roles []string) ([]model.FinanceWorkspaceResponse, error)
}

type FinanceWorkspaceService struct {
	repo FinanceWorkspaceRepository
	log  *slog.Logger
}

func NewFinanceWorkspaceService(repo FinanceWorkspaceRepository, log *slog.Logger) FinanceWorkspaceService {
	return FinanceWorkspaceService{
		repo: repo,
		log:  log,
	}
}

func (s FinanceWorkspaceService) GetAccessibleWorkspaces(ctx context.Context, userID uuid.UUID) ([]model.FinanceWorkspaceResponse, error) {
	financeRoles := []string{
		authz.ModuleRole(authz.ResourceFinance, authz.LevelAdmin),
		authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter),
		authz.ModuleRole(authz.ResourceFinance, authz.LevelReader),
	}

	return s.repo.GetAccessibleWorkspaces(ctx, userID, financeRoles)
}
