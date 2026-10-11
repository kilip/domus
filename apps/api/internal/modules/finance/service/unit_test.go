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

type mockFinanceUnitRepo struct {
	calledRoles []string
	units       []model.FinanceUnitResponse
	err         error
}

func (m *mockFinanceUnitRepo) GetAccessibleUnits(ctx context.Context, userID uuid.UUID, roles []string) ([]model.FinanceUnitResponse, error) {
	m.calledRoles = roles
	return m.units, m.err
}

func TestFinanceUnitService_GetAccessibleUnits(t *testing.T) {
	code := "U1"
	expected := []model.FinanceUnitResponse{
		{ID: 1, Name: "Unit 1", Code: &code, Kind: "category", ParishID: 10},
	}
	repo := &mockFinanceUnitRepo{units: expected}

	svc := service.NewFinanceUnitService(repo, nil)
	userID := uuid.New()
	res, err := svc.GetAccessibleUnits(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, expected, res)
	assert.ElementsMatch(t, []string{"finance-admin", "finance-writer", "finance-reader"}, repo.calledRoles)
}

