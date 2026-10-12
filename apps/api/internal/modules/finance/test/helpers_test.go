package test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/user"
	entworkspace "github.com/paroki/domus/api/ent/workspace"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// GrantFinanceRole grants the finance role to all workspaces belonging to the parish, as well as the parish scope.
func GrantFinanceRole[T any](s *testutil.ApiTestSuite[T], parishID int, role string) {
	s.T().Helper()
	st := testutil.GetState()
	workspaces, _ := st.Ent.Workspace.Query().Where(entworkspace.HasParishWith(parish.IDEQ(parishID))).All(context.Background())
	for _, w := range workspaces {
		s.GrantRole(role, authz.Scope{Type: authz.ScopeWorkspace, ID: int64(w.ID)})
	}
	s.GrantRole(role, authz.Scope{Type: authz.ScopeParish, ID: int64(parishID)})
}

// SetFinanceCookie sets the domus.fin cookie in the test suite for active parish and workspace.
func SetFinanceCookie[T any](s *testutil.ApiTestSuite[T], parishID int, workspaceID int) {
	s.T().Helper()
	cookieVal := fmt.Sprintf(`{"activeParishId":%d,"activeWorkspaceId":%d}`, parishID, workspaceID)
	s.SetCookie(&http.Cookie{
		Name:  "domus.fin",
		Value: url.QueryEscape(cookieVal),
		Path:  "/",
	})
}

// EnsureUser ensures that the current test user is stored in the database.
func EnsureUser[T any](s *testutil.ApiTestSuite[T]) {
	s.T().Helper()
	st := testutil.GetState()
	exists, err := st.Ent.User.Query().Where(user.IDEQ(s.User.ID)).Exist(context.Background())
	require.NoError(s.T(), err)
	if !exists {
		_, err = st.Ent.User.Create().
			SetID(s.User.ID).
			SetName(s.User.Name).
			SetEmail(s.User.Email).
			SetAuthUserID(s.User.ID).
			SetStatus(user.StatusActive).
			Save(context.Background())
		require.NoError(s.T(), err)
	}
}

// CreateTestParish creates a fresh parish with a unique name directly via Ent client.
func CreateTestParish(t *testing.T) *ent.Parish {
	t.Helper()
	st := testutil.GetState()
	uniqueSuffix := uuid.New().String()[:8]
	parish, err := st.Ent.Parish.Create().
		SetName(fmt.Sprintf("Parish-%s", uniqueSuffix)).
		Save(context.Background())
	require.NoError(t, err)
	return parish
}

// CreateTestWorkspace creates a fresh workspace belonging to the specified parish directly via Ent client.
func CreateTestWorkspace(t *testing.T, parishID int, userID uuid.UUID) *ent.Workspace {
	t.Helper()
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: userID})
	uniqueSuffix := uuid.New().String()[:8]
	workspace, err := st.Ent.Workspace.Create().
		SetName(fmt.Sprintf("Workspace-%s", uniqueSuffix)).
		SetParishID(parishID).
		Save(ctx)
	require.NoError(t, err)
	return workspace
}

// CreateTestUnit alias for backwards compatibility during refactor
func CreateTestUnit(t *testing.T, parishID int, userID uuid.UUID) *ent.Workspace {
	return CreateTestWorkspace(t, parishID, userID)
}

type FinanceSmokeSuite struct {
	testutil.ApiTestSuite[any]
}

func (s *FinanceSmokeSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
}

func (s *FinanceSmokeSuite) TestHelpersAndRoleGrant_Smoke() {
	// 1. Ensure user  exists in DB
	EnsureUser(&s.ApiTestSuite)

	// Verify user exists in ent DB
	st := testutil.GetState()
	userExists, err := st.Ent.User.Get(context.Background(), s.User.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.User.Email, userExists.Email)

	// 2. Create fresh parish per test
	parish := CreateTestParish(s.T())
	assert.Greater(s.T(), parish.ID, 0)

	// 3. Grant finance-writer role in parish scope
	parishScope := authz.Scope{Type: authz.ScopeParish, ID: int64(parish.ID)}
	role := authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter)
	s.GrantRole(role, parishScope)

	// 4. Validate enforcer permissions
	canWrite, err := st.Authz.Can(context.Background(), s.User.ID.String(), parishScope, authz.ResourceFinance, authz.ActionWrite)
	require.NoError(s.T(), err)
	assert.True(s.T(), canWrite, "user should have write permission in granted parish")

	canRead, err := st.Authz.Can(context.Background(), s.User.ID.String(), parishScope, authz.ResourceFinance, authz.ActionRead)
	require.NoError(s.T(), err)
	assert.True(s.T(), canRead, "finance-writer should also have read permission")

	// Another parish should NOT have access
	otherParishScope := authz.Scope{Type: authz.ScopeParish, ID: int64(parish.ID + 9999)}
	canOther, err := st.Authz.Can(context.Background(), s.User.ID.String(), otherParishScope, authz.ResourceFinance, authz.ActionWrite)
	require.NoError(s.T(), err)
	assert.False(s.T(), canOther, "user should not have permission in another parish")
}

func TestFinanceSmokeSuite(t *testing.T) {
	suite.Run(t, new(FinanceSmokeSuite))
}
