package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
)

type UserRepository interface {
	GetByID(c context.Context, id uuid.UUID) (*model.UserResponse, error)
	GetByAuthID(c context.Context, authID uuid.UUID) (*model.UserResponse, error)
	GetByEmail(c context.Context, email string) (*model.UserResponse, error)
	Activate(c context.Context, id uuid.UUID) (*model.UserResponse, error)
	BindAuthIDAndActivate(c context.Context, id uuid.UUID, authID uuid.UUID) (*model.UserResponse, error)
	Create(c context.Context, req model.CreateUserRequest) (*model.UserResponse, error)
	Update(c context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error)
	Delete(c context.Context, id uuid.UUID) error
}

type UserService struct {
	users UserRepository
	log   *slog.Logger
}

func NewUserService(users UserRepository, log *slog.Logger) UserService {
	return UserService{users, log}
}

func (s UserService) Create(c context.Context, req model.CreateUserRequest) error {
	if _, err := s.users.Create(c, req); err != nil {
		return err
	}

	return nil
}

// Validate ensures user exists and activates invited users.
// If user is unknown or soft-deleted, returns core.ErrUserNotRegistered (403).
// If user was invited and email is verified in claims, binds auth_user_id (if not bound) and sets status = active.
// If user is active, performs no database write.
func (s UserService) Validate(c context.Context, user core.AuthenticatedUser) (*model.UserResponse, error) {
	var authID uuid.UUID
	if user.AuthID != nil && *user.AuthID != uuid.Nil {
		authID = *user.AuthID
	} else if user.ID != uuid.Nil {
		authID = user.ID
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(user.Email))

	// 1. Try finding by auth_user_id
	if authID != uuid.Nil {
		existing, err := s.users.GetByAuthID(c, authID)
		if err == nil {
			if existing.Status == "invited" {
				return s.users.Activate(c, existing.ID)
			}
			return existing, nil
		}
		if !core.IsNotFound(err) {
			return nil, err
		}
	}

	// 2. Not found by auth_user_id: try binding once by email if invited
	if cleanEmail == "" {
		return nil, core.ErrUserNotRegistered
	}

	existing, err := s.users.GetByEmail(c, cleanEmail)
	if err != nil {
		if core.IsNotFound(err) {
			return nil, core.ErrUserNotRegistered
		}
		return nil, err
	}

	// If user is already bound to a different auth_user_id, do not allow rebinding
	if existing.AuthUserID != nil && *existing.AuthUserID != authID {
		return nil, core.ErrUserNotRegistered
	}

	// Email must be verified in claims before binding
	if !user.EmailVerified {
		return nil, core.ErrUserNotRegistered
	}

	if authID == uuid.Nil {
		return nil, core.ErrUserNotRegistered
	}

	return s.users.BindAuthIDAndActivate(c, existing.ID, authID)
}
