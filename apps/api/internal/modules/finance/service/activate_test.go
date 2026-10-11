package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/service"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockActivateRepo struct {
	parishExists       bool
	parishExistsErr    error
	unitExistsInParish bool
	unitExistsErr      error
}

func (m *mockActivateRepo) ParishExists(ctx context.Context, parishID int) (bool, error) {
	return m.parishExists, m.parishExistsErr
}

func (m *mockActivateRepo) UnitExistsInParish(ctx context.Context, unitID int, parishID int) (bool, error) {
	return m.unitExistsInParish, m.unitExistsErr
}

func TestActivateService_ParishNotFound(t *testing.T) {
	repo := &mockActivateRepo{parishExists: false}
	az, err := authz.New()
	require.NoError(t, err)

	svc := service.NewActivateService(repo, az, nil)
	_, err = svc.Activate(context.Background(), uuid.New(), model.ActivateFinanceRequest{
		ParishID: 1,
		UnitID:   2,
	})

	assert.ErrorIs(t, err, model.ErrParishNotFound)
}

func TestActivateService_UnitNotFound(t *testing.T) {
	repo := &mockActivateRepo{parishExists: true, unitExistsInParish: false}
	az, err := authz.New()
	require.NoError(t, err)

	svc := service.NewActivateService(repo, az, nil)
	_, err = svc.Activate(context.Background(), uuid.New(), model.ActivateFinanceRequest{
		ParishID: 1,
		UnitID:   2,
	})

	assert.ErrorIs(t, err, model.ErrUnitNotFound)
}

func TestActivateService_Forbidden(t *testing.T) {
	repo := &mockActivateRepo{parishExists: true, unitExistsInParish: true}
	az, err := authz.New()
	require.NoError(t, err)

	svc := service.NewActivateService(repo, az, nil)
	userID := uuid.New()
	_, err = svc.Activate(context.Background(), userID, model.ActivateFinanceRequest{
		ParishID: 1,
		UnitID:   2,
	})

	assert.ErrorIs(t, err, core.ErrForbidden)
}

func TestActivateService_Success(t *testing.T) {
	repo := &mockActivateRepo{parishExists: true, unitExistsInParish: true}
	az, err := authz.New()
	require.NoError(t, err)

	userID := uuid.New()
	err = az.AddMembership(userID.String(), authz.ModuleRole(authz.ResourceFinance, authz.LevelReader), authz.Scope{
		Type: authz.ScopeUnit,
		ID:   2,
	})
	require.NoError(t, err)

	svc := service.NewActivateService(repo, az, nil)
	resp, err := svc.Activate(context.Background(), userID, model.ActivateFinanceRequest{
		ParishID: 1,
		UnitID:   2,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, resp.ActiveParishID)
	assert.Equal(t, 2, resp.ActiveUnitID)
}

