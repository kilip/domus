package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFinanceWorkspaceRepo struct {
	calledRoles []string
	workspaces  []model.FinanceWorkspaceResponse
	err         error
}

func (m *mockFinanceWorkspaceRepo) GetAccessibleWorkspaces(ctx context.Context, userID uuid.UUID, roles []string) ([]model.FinanceWorkspaceResponse, error) {
	m.calledRoles = roles
	return m.workspaces, m.err
}

func TestFinanceWorkspaceService_GetAccessibleWorkspaces(t *testing.T) {
	expected := []model.FinanceWorkspaceResponse{
		{ID: 1, Name: "Workspace 1", ParishID: 10},
	}
	repo := &mockFinanceWorkspaceRepo{workspaces: expected}

	svc := service.NewFinanceWorkspaceService(repo, nil)
	userID := uuid.New()
	res, err := svc.GetAccessibleWorkspaces(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, expected, res)
	assert.ElementsMatch(t, []string{"finance-admin", "finance-writer", "finance-reader"}, repo.calledRoles)
}
