package test

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FinanceUnitTestSuite struct {
	testutil.ApiTestSuite[[]model.FinanceUnitResponse]
}

func (s *FinanceUnitTestSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *FinanceUnitTestSuite) TestGetAccessibleUnits_EmptyWhenNoRoles() {
	s.Request("/fin/units", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	assert.Empty(s.T(), res.Data)
}

func (s *FinanceUnitTestSuite) TestGetAccessibleUnits_Success() {
	st := testutil.GetState()
	parish := CreateTestParish(s.T())
	unit1 := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	_ = CreateTestUnit(s.T(), parish.ID, s.User.ID) // unit2 without membership

	// Grant finance-writer membership to unit1 directly in Membership table
	_, err := st.Ent.Membership.Create().
		SetUserID(s.User.ID.String()).
		SetScopeType("unit").
		SetScopeID(int64(unit1.ID)).
		SetRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter)).
		Save(context.Background())
	require.NoError(s.T(), err)

	s.Request("/fin/units", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	require.Len(s.T(), res.Data, 1)
	assert.Equal(s.T(), unit1.ID, res.Data[0].ID)
	assert.Equal(s.T(), unit1.Name, res.Data[0].Name)
	assert.Equal(s.T(), parish.ID, res.Data[0].ParishID)
}

func (s *FinanceUnitTestSuite) TestGetAccessibleUnits_InactiveUnitExcluded() {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	parish := CreateTestParish(s.T())
	unitInactive := CreateTestUnit(s.T(), parish.ID, s.User.ID)

	// Set unit to inactive
	err := st.Ent.Unit.UpdateOneID(unitInactive.ID).
		SetStatus(unit.StatusInactive).
		Exec(ctx)
	require.NoError(s.T(), err)

	_, err = st.Ent.Membership.Create().
		SetUserID(s.User.ID.String()).
		SetScopeType("unit").
		SetScopeID(int64(unitInactive.ID)).
		SetRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelAdmin)).
		Save(context.Background())
	require.NoError(s.T(), err)

	s.Request("/fin/units", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	assert.Empty(s.T(), res.Data)
}

func TestFinanceUnitTestSuite(t *testing.T) {
	suite.Run(t, new(FinanceUnitTestSuite))
}

