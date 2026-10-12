package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/user"
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

func userToResponse(u *ent.User) *model.UserResponse {
	if u == nil {
		return nil
	}
	return &model.UserResponse{
		ID:         u.ID,
		AuthUserID: u.AuthUserID,
		Name:       u.Name,
		Email:      u.Email,
		Avatar:     u.Avatar,
		Status:     string(u.Status),
	}
}

func (r UserRepository) GetByID(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
	u, err := r.entcli.User.Get(c, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	return userToResponse(u), nil
}

func (r UserRepository) GetByAuthID(c context.Context, authID uuid.UUID) (*model.UserResponse, error) {
	u, err := r.entcli.User.Query().
		Where(user.AuthUserIDEQ(authID)).
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	return userToResponse(u), nil
}

func (r UserRepository) GetByEmail(c context.Context, email string) (*model.UserResponse, error) {
	u, err := r.entcli.User.Query().
		Where(user.EmailEqualFold(email)).
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	return userToResponse(u), nil
}

func (r UserRepository) Activate(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
	var u *ent.User
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		var err error
		u, err = tx.User.UpdateOneID(id).
			SetStatus(user.StatusActive).
			Save(c)
		return err
	})
	if err != nil {
		return nil, err
	}
	return userToResponse(u), nil
}

func (r UserRepository) BindAuthIDAndActivate(c context.Context, id uuid.UUID, authID uuid.UUID) (*model.UserResponse, error) {
	var u *ent.User
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		var err error
		u, err = tx.User.UpdateOneID(id).
			SetAuthUserID(authID).
			SetStatus(user.StatusActive).
			Save(c)
		return err
	})
	if err != nil {
		return nil, err
	}
	return userToResponse(u), nil
}

func (r UserRepository) Create(c context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
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

	return userToResponse(user), nil
}

func (r UserRepository) Update(c context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error) {
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

	return userToResponse(user), nil
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
