package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EnsureUserSnapshot performs an authenticated GET /ping to ensure
// that the current test user snapshot is stored in the database.
func EnsureUserSnapshot[T any](s *testutil.ApiTestSuite[T]) {
	s.T().Helper()
	s.Request("/ping", fiber.MethodGet, nil)
	s.OK()
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

type FinanceSmokeSuite struct {
	testutil.ApiTestSuite[any]
}

func (s *FinanceSmokeSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
}

func (s *FinanceSmokeSuite) TestHelpersAndRoleGrant_Smoke() {
	// 1. Ensure user snapshot exists in DB
	EnsureUserSnapshot(&s.ApiTestSuite)

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
