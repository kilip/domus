package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type UserRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewUserRepository(entcli *ent.Client, log *slog.Logger) UserRepository {
	return UserRepository{entcli, log}
}

func (r UserRepository) GetByID(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
	var response model.UserResponse

	user, err := r.entcli.User.Get(c, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	core.ToValue(user, &response)

	return &response, nil
}

func (r UserRepository) Create(c context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
	var res model.UserResponse
	var user *ent.User
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		var err error
		user, err = tx.User.Create().
			SetID(req.ID).
			SetName(req.Name).
			SetEmail(req.Email).
			SetNillableAvatar(nilIfEmpty(req.Avatar)).
			Save(c)

		return err
	})

	if err != nil {
		return nil, err
	}

	core.ToValue(user, &res)

	return &res, nil
}

func (r UserRepository) Update(c context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error) {
	var res model.UserResponse
	var user *ent.User
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		var err error
		user, err = tx.User.UpdateOneID(id).
			SetName(req.Name).
			SetEmail(req.Email).
			SetNillableAvatar(nilIfEmpty(req.Avatar)).
			Save(c)

		return err
	})

	if err != nil {
		return nil, err
	}

	core.ToValue(user, &res)

	return &res, nil
}

// nilIfEmpty maps "" to nil so empty optional fields are stored as NULL.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r UserRepository) Delete(c context.Context, id uuid.UUID) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		return tx.User.DeleteOneID(id).Exec(c)
	})
}
