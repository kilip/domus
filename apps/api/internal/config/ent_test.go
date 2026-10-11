package config

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/diocese"
	"github.com/paroki/domus/api/ent/enttest"
	"github.com/paroki/domus/api/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureEntCli_AutoSetCreatorAndUpdater(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent_hook_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ConfigureEntCli(client)

	user1ID := core.GenerateID()
	user2ID := core.GenerateID()

	// Seed users first because Diocese has foreign key edges to User
	_, err := client.User.Create().
		SetID(user1ID).
		SetName("User One").
		SetEmail("user1@example.com").
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.User.Create().
		SetID(user2ID).
		SetName("User Two").
		SetEmail("user2@example.com").
		Save(context.Background())
	require.NoError(t, err)

	// Test 1: Create Diocese with User 1 in context (creator and updater auto-set)
	ctxUser1 := core.ContextWithUser(context.Background(), core.AuthenticatedUser{
		ID:    user1ID,
		Name:  "User One",
		Email: "user1@example.com",
	})

	diocese, err := client.Diocese.Create().
		SetName("Keuskupan Surabaya").
		Save(ctxUser1)
	require.NoError(t, err)
	assert.Equal(t, user1ID, diocese.CreatedBy)
	assert.Equal(t, user1ID, diocese.UpdatedBy)

	// Test 2: Update Diocese with User 2 in context (updater auto-set)
	ctxUser2 := core.ContextWithUser(context.Background(), core.AuthenticatedUser{
		ID:    user2ID,
		Name:  "User Two",
		Email: "user2@example.com",
	})

	updatedDiocese, err := client.Diocese.UpdateOneID(diocese.ID).
		SetName("Keuskupan Surabaya Updated").
		Save(ctxUser2)
	require.NoError(t, err)
	assert.Equal(t, user1ID, updatedDiocese.CreatedBy)
	assert.Equal(t, user2ID, updatedDiocese.UpdatedBy)
}

func TestConfigureEntCli_SoftDelete(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent_soft_delete_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ConfigureEntCli(client)

	userID := core.GenerateID()
	_, err := client.User.Create().
		SetID(userID).
		SetName("Admin User").
		SetEmail("admin@example.com").
		Save(context.Background())
	require.NoError(t, err)

	ctxUser := core.ContextWithUser(context.Background(), core.AuthenticatedUser{
		ID:    userID,
		Name:  "Admin User",
		Email: "admin@example.com",
	})

	// 1. Create a Diocese
	d, err := client.Diocese.Create().
		SetName("Keuskupan Malang").
		Save(ctxUser)
	require.NoError(t, err)
	assert.Nil(t, d.DeletedAt)

	// 2. Soft delete the Diocese
	err = client.Diocese.DeleteOneID(d.ID).Exec(ctxUser)
	require.NoError(t, err)

	// 3. Normal queries should not find the soft-deleted diocese
	_, err = client.Diocese.Get(context.Background(), d.ID)
	assert.True(t, ent.IsNotFound(err))

	exists, err := client.Diocese.Query().Where(diocese.IDEQ(d.ID)).Exist(context.Background())
	require.NoError(t, err)
	assert.False(t, exists)

	all, err := client.Diocese.Query().All(context.Background())
	require.NoError(t, err)
	assert.Empty(t, all)

	// 4. Querying with SkipSoftDelete should find it with DeletedAt set
	skipCtx := core.SkipSoftDelete(context.Background())
	found, err := client.Diocese.Get(skipCtx, d.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
	assert.Equal(t, userID, found.UpdatedBy)

	// 5. Deleting an already soft-deleted entity without skip should return NotFound
	err = client.Diocese.DeleteOneID(d.ID).Exec(context.Background())
	assert.True(t, ent.IsNotFound(err))

	// 6. Deleting with SkipSoftDelete should hard delete
	err = client.Diocese.DeleteOneID(d.ID).Exec(skipCtx)
	require.NoError(t, err)

	_, err = client.Diocese.Get(skipCtx, d.ID)
	assert.True(t, ent.IsNotFound(err))

	// 7. Test Parish soft delete and edge traversal
	d2, err := client.Diocese.Create().
		SetName("Keuskupan Bandung").
		Save(ctxUser)
	require.NoError(t, err)

	p, err := client.Parish.Create().
		SetName("Paroki Santo Yusuf").
		SetDioceseID(d2.ID).
		Save(context.Background())
	require.NoError(t, err)
	assert.Nil(t, p.DeletedAt)

	// Soft delete parish
	err = client.Parish.DeleteOneID(p.ID).Exec(context.Background())
	require.NoError(t, err)

	// Query parish normally -> not found
	_, err = client.Parish.Get(context.Background(), p.ID)
	assert.True(t, ent.IsNotFound(err))

	// Query parishes through diocese edge -> excluded
	parishes, err := d2.QueryParishes().All(context.Background())
	require.NoError(t, err)
	assert.Empty(t, parishes)

	// Query parishes with skipCtx -> found
	parishesWithDeleted, err := d2.QueryParishes().All(skipCtx)
	require.NoError(t, err)
	assert.Len(t, parishesWithDeleted, 1)
	assert.NotNil(t, parishesWithDeleted[0].DeletedAt)
}
