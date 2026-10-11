package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustCan(t *testing.T, e *Enforcer, user string, scope Scope, obj Resource, act Action) bool {
	t.Helper()
	ok, err := e.Can(context.Background(), user, scope, obj, act)
	require.NoError(t, err)
	return ok
}

func newEnforcer(t *testing.T) *Enforcer {
	t.Helper()
	e, err := New()
	require.NoError(t, err)
	return e
}

func TestSuperadmin(t *testing.T) {
	e := newEnforcer(t)
	require.NoError(t, e.AddMembership("u1", RoleSuperadmin, System))

	for _, obj := range []Resource{ResourceDiocese, ResourceParish} {
		require.True(t, mustCan(t, e, "u1", System, obj, ActionRead))
		require.True(t, mustCan(t, e, "u1", System, obj, ActionWrite))
	}
	require.False(t, mustCan(t, e, "u1", System, ResourceFinance, ActionRead))
}

func TestModuleRolesAreScoped(t *testing.T) {
	e := newEnforcer(t)
	parish12 := Scope{ScopeParish, 12}
	parish13 := Scope{ScopeParish, 13}

	require.NoError(t, e.AddMembership("w", ModuleRole(ResourceFinance, LevelWriter), parish12))
	require.NoError(t, e.AddMembership("r", ModuleRole(ResourceFinance, LevelReader), parish12))

	require.True(t, mustCan(t, e, "w", parish12, ResourceFinance, ActionWrite))
	require.True(t, mustCan(t, e, "w", parish12, ResourceFinance, ActionRead))
	require.False(t, mustCan(t, e, "w", parish13, ResourceFinance, ActionWrite))
	require.False(t, mustCan(t, e, "w", parish12, ResourceSacrament, ActionRead))

	require.True(t, mustCan(t, e, "r", parish12, ResourceFinance, ActionRead))
	require.False(t, mustCan(t, e, "r", parish12, ResourceFinance, ActionWrite))
}

func TestNoScopeInheritance(t *testing.T) {
	e := newEnforcer(t)
	require.NoError(t, e.AddMembership("u", ModuleRole(ResourceFinance, LevelAdmin), Scope{ScopeDiocese, 1}))

	require.True(t, mustCan(t, e, "u", Scope{ScopeDiocese, 1}, ResourceFinance, ActionWrite))
	require.False(t, mustCan(t, e, "u", Scope{ScopeParish, 12}, ResourceFinance, ActionWrite))
}

func TestRemoveMembership(t *testing.T) {
	e := newEnforcer(t)
	role := ModuleRole(ResourceSacrament, LevelWriter)
	scope := Scope{ScopeUnit, 5}

	require.NoError(t, e.AddMembership("u", role, scope))
	require.True(t, mustCan(t, e, "u", scope, ResourceSacrament, ActionWrite))

	require.NoError(t, e.RemoveMembership("u", role, scope))
	require.False(t, mustCan(t, e, "u", scope, ResourceSacrament, ActionWrite))
}

func TestUnknownUserDenied(t *testing.T) {
	e := newEnforcer(t)
	require.False(t, mustCan(t, e, "nobody", System, ResourceDiocese, ActionRead))
}
