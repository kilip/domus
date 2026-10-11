package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/stretchr/testify/assert"
)

func TestToCoreUser(t *testing.T) {
	// nil user
	assert.Empty(t, toCoreUser(nil).ID)

	// valid user
	id := uuid.New()
	avatar := "https://example.com/avatar.jpg"
	u := &ent.User{
		ID:     id,
		Name:   "Finance User",
		Email:  "finance@example.com",
		Avatar: avatar,
	}

	cu := toCoreUser(u)
	assert.Equal(t, id, cu.ID)
	assert.Equal(t, "Finance User", cu.Name)
	assert.Equal(t, "finance@example.com", cu.Email)
	assert.Equal(t, avatar, cu.Avatar)
}
