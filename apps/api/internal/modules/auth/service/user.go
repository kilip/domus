package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
)

type UserRepository interface {
	GetByID(c context.Context, id uuid.UUID) (*model.UserResponse, error)
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

// Validate ensure user exists
func (s UserService) Validate(c context.Context, user core.AuthenticatedUser) error {
	_, err := s.users.GetByID(c, user.ID)

	if err != nil {
		if core.IsNotFound(err) {
			return s.Create(c, model.CreateUserRequest{
				ID:     user.ID,
				Name:   user.Name,
				Email:  user.Email,
				Avatar: user.Avatar,
			})
		}
		return err
	}

	return nil
}
