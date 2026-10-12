package test

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent/workspace"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FinanceWorkspaceTestSuite struct {
	testutil.ApiTestSuite[[]model.FinanceWorkspaceResponse]
}

func (s *FinanceWorkspaceTestSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUser(&s.ApiTestSuite)
}

func (s *FinanceWorkspaceTestSuite) TestGetAccessibleWorkspaces_EmptyWhenNoRoles() {
	s.Request("/fin/workspaces", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	assert.Empty(s.T(), res.Data)
}

func (s *FinanceWorkspaceTestSuite) TestGetAccessibleWorkspaces_Success() {
	st := testutil.GetState()
	parish := CreateTestParish(s.T())
	ws1 := CreateTestWorkspace(s.T(), parish.ID, s.User.ID)
	_ = CreateTestWorkspace(s.T(), parish.ID, s.User.ID) // ws2 without membership

	// Grant finance-writer membership to ws1 directly in Membership table
	_, err := st.Ent.Membership.Create().
		SetUserID(s.User.ID.String()).
		SetScopeType("workspace").
		SetScopeID(int64(ws1.ID)).
		SetRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter)).
		Save(context.Background())
	require.NoError(s.T(), err)

	s.Request("/fin/workspaces", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	require.Len(s.T(), res.Data, 1)
	assert.Equal(s.T(), ws1.ID, res.Data[0].ID)
	assert.Equal(s.T(), ws1.Name, res.Data[0].Name)
	assert.Equal(s.T(), parish.ID, res.Data[0].ParishID)
}

func (s *FinanceWorkspaceTestSuite) TestGetAccessibleWorkspaces_InactiveWorkspaceExcluded() {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	parish := CreateTestParish(s.T())
	wsInactive := CreateTestWorkspace(s.T(), parish.ID, s.User.ID)

	// Set workspace to inactive
	err := st.Ent.Workspace.UpdateOneID(wsInactive.ID).
		SetStatus(workspace.StatusInactive).
		Exec(ctx)
	require.NoError(s.T(), err)

	_, err = st.Ent.Membership.Create().
		SetUserID(s.User.ID.String()).
		SetScopeType("workspace").
		SetScopeID(int64(wsInactive.ID)).
		SetRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelAdmin)).
		Save(context.Background())
	require.NoError(s.T(), err)

	s.Request("/fin/workspaces", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	assert.Empty(s.T(), res.Data)
}

func TestFinanceWorkspaceTestSuite(t *testing.T) {
	suite.Run(t, new(FinanceWorkspaceTestSuite))
}
