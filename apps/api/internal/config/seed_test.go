package config

import (
	"context"
	"io"
	"log/slog"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paroki/domus/api/ent/enttest"
	"github.com/paroki/domus/api/ent/membership"
	"github.com/paroki/domus/api/ent/user"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedSuperadmin_EmptyEmail_Skips(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:seed_empty_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ConfigureEntCli(client)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := Config{SuperadminEmail: ""}

	err := SeedSuperadmin(context.Background(), client, cfg, logger)
	require.NoError(t, err)

	userCount, err := client.User.Query().Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, userCount)

	membershipCount, err := client.Membership.Query().Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, membershipCount)
}

func TestSeedSuperadmin_IdempotentRunTwice(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:seed_twice_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ConfigureEntCli(client)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := Config{
		SuperadminEmail: "admin@example.com",
		SuperadminName:  "Test Super Admin",
	}

	enforcer, err := authz.New()
	require.NoError(t, err)

	ctx := context.Background()

	// Run 1: Should create user (status=invited) and membership
	err = SeedSuperadmin(ctx, client, cfg, logger, enforcer)
	require.NoError(t, err)

	u1, err := client.User.Query().Where(user.EmailEqualFold("admin@example.com")).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "admin@example.com", u1.Email)
	assert.Equal(t, "Test Super Admin", u1.Name)
	assert.Equal(t, user.StatusInvited, u1.Status)

	mems, err := client.Membership.Query().Where(membership.UserIDEQ(u1.ID.String())).All(ctx)
	require.NoError(t, err)
	require.Len(t, mems, 1)
	assert.Equal(t, membership.ScopeTypeSystem, mems[0].ScopeType)
	assert.Equal(t, int64(0), mems[0].ScopeID)
	assert.Equal(t, authz.RoleSuperadmin, mems[0].Role)

	// Verify Casbin authorization
	canWrite, err := enforcer.Can(ctx, u1.ID.String(), authz.System, authz.ResourceUser, authz.ActionWrite)
	require.NoError(t, err)
	assert.True(t, canWrite)

	// Run 2: Seed again with the same config - must be idempotent
	err = SeedSuperadmin(ctx, client, cfg, logger, enforcer)
	require.NoError(t, err)

	userCount, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, userCount, "running seed twice must produce exactly one user")

	membershipCount, err := client.Membership.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, membershipCount, "running seed twice must produce exactly one membership")
}

func TestSeedSuperadmin_DoesNotDowngradeActiveUser(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:seed_active_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ConfigureEntCli(client)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	// Pre-create an active user
	existingID := core.GenerateID()
	_, err := client.User.Create().
		SetID(existingID).
		SetName("Existing Active User").
		SetEmail("super@example.com").
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	cfg := Config{
		SuperadminEmail: "SUPER@EXAMPLE.COM", // uppercase test
		SuperadminName:  "",                  // defaults to Super Admin
	}

	err = SeedSuperadmin(ctx, client, cfg, logger)
	require.NoError(t, err)

	u, err := client.User.Query().Where(user.EmailEqualFold("super@example.com")).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, existingID, u.ID)
	assert.Equal(t, user.StatusActive, u.Status, "must not downgrade active status to invited")

	mems, err := client.Membership.Query().Where(membership.UserIDEQ(u.ID.String())).All(ctx)
	require.NoError(t, err)
	require.Len(t, mems, 1)
	assert.Equal(t, membership.ScopeTypeSystem, mems[0].ScopeType)
	assert.Equal(t, int64(0), mems[0].ScopeID)
	assert.Equal(t, authz.RoleSuperadmin, mems[0].Role)
}

