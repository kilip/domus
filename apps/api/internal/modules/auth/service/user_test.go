package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
	"github.com/paroki/domus/api/internal/modules/auth/service"
	"github.com/stretchr/testify/assert"
)

type fakeUserRepo struct {
	getByAuthIDFn            func(c context.Context, authID uuid.UUID) (*model.UserResponse, error)
	getByEmailFn             func(c context.Context, email string) (*model.UserResponse, error)
	activateFn               func(c context.Context, id uuid.UUID) (*model.UserResponse, error)
	bindAuthIDAndActivateFn  func(c context.Context, id uuid.UUID, authID uuid.UUID) (*model.UserResponse, error)
	createFn                 func(c context.Context, req model.CreateUserRequest) (*model.UserResponse, error)
	getByIDFn                func(c context.Context, id uuid.UUID) (*model.UserResponse, error)
	updateFn                 func(c context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error)
	deleteFn                 func(c context.Context, id uuid.UUID) error
}

func (f *fakeUserRepo) GetByID(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(c, id)
	}
	return nil, nil
}

func (f *fakeUserRepo) GetByAuthID(c context.Context, authID uuid.UUID) (*model.UserResponse, error) {
	if f.getByAuthIDFn != nil {
		return f.getByAuthIDFn(c, authID)
	}
	return nil, core.ErrItemNotFound
}

func (f *fakeUserRepo) GetByEmail(c context.Context, email string) (*model.UserResponse, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(c, email)
	}
	return nil, core.ErrItemNotFound
}

func (f *fakeUserRepo) Activate(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
	if f.activateFn != nil {
		return f.activateFn(c, id)
	}
	return nil, nil
}

func (f *fakeUserRepo) BindAuthIDAndActivate(c context.Context, id uuid.UUID, authID uuid.UUID) (*model.UserResponse, error) {
	if f.bindAuthIDAndActivateFn != nil {
		return f.bindAuthIDAndActivateFn(c, id, authID)
	}
	return nil, nil
}

func (f *fakeUserRepo) Create(c context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
	if f.createFn != nil {
		return f.createFn(c, req)
	}
	return nil, nil
}

func (f *fakeUserRepo) Update(c context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error) {
	if f.updateFn != nil {
		return f.updateFn(c, id, req)
	}
	return nil, nil
}

func (f *fakeUserRepo) Delete(c context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(c, id)
	}
	return nil
}

func TestUserService_Validate_UnknownUser(t *testing.T) {
	repo := &fakeUserRepo{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewUserService(repo, logger)

	authID := uuid.New()
	user := core.AuthenticatedUser{
		ID:            authID,
		AuthID:        &authID,
		Email:         "unknown@example.com",
		EmailVerified: true,
	}

	res, err := svc.Validate(context.Background(), user)
	assert.ErrorIs(t, err, core.ErrUserNotRegistered)
	assert.Nil(t, res)
}

func TestUserService_Validate_ActiveUser_NoWrite(t *testing.T) {
	internalID := uuid.New()
	authID := uuid.New()
	user := core.AuthenticatedUser{
		ID:            authID,
		AuthID:        &authID,
		Email:         "active@example.com",
		EmailVerified: true,
	}

	existing := &model.UserResponse{
		ID:         internalID,
		AuthUserID: &authID,
		Email:      "active@example.com",
		Status:     "active",
	}

	activateCalled := false
	repo := &fakeUserRepo{
		getByAuthIDFn: func(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
			assert.Equal(t, authID, id)
			return existing, nil
		},
		activateFn: func(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
			activateCalled = true
			return nil, nil
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewUserService(repo, logger)

	res, err := svc.Validate(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, internalID, res.ID)
	assert.False(t, activateCalled, "Activate must not be called when user is already active")
}

func TestUserService_Validate_InvitedUser_Bound_Activates(t *testing.T) {
	internalID := uuid.New()
	authID := uuid.New()
	user := core.AuthenticatedUser{
		ID:            authID,
		AuthID:        &authID,
		Email:         "invited@example.com",
		EmailVerified: true,
	}

	existing := &model.UserResponse{
		ID:         internalID,
		AuthUserID: &authID,
		Email:      "invited@example.com",
		Status:     "invited",
	}
	activated := &model.UserResponse{
		ID:         internalID,
		AuthUserID: &authID,
		Email:      "invited@example.com",
		Status:     "active",
	}

	activateCalled := false
	repo := &fakeUserRepo{
		getByAuthIDFn: func(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
			return existing, nil
		},
		activateFn: func(c context.Context, id uuid.UUID) (*model.UserResponse, error) {
			activateCalled = true
			assert.Equal(t, internalID, id)
			return activated, nil
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewUserService(repo, logger)

	res, err := svc.Validate(context.Background(), user)
	assert.NoError(t, err)
	assert.True(t, activateCalled)
	assert.Equal(t, "active", res.Status)
}

func TestUserService_Validate_InvitedUser_UnverifiedEmail_Rejected(t *testing.T) {
	internalID := uuid.New()
	authID := uuid.New()
	user := core.AuthenticatedUser{
		ID:            authID,
		AuthID:        &authID,
		Email:         "invited@example.com",
		EmailVerified: false,
	}

	existing := &model.UserResponse{
		ID:     internalID,
		Email:  "invited@example.com",
		Status: "invited",
	}

	bindCalled := false
	repo := &fakeUserRepo{
		getByEmailFn: func(c context.Context, email string) (*model.UserResponse, error) {
			return existing, nil
		},
		bindAuthIDAndActivateFn: func(c context.Context, id uuid.UUID, aID uuid.UUID) (*model.UserResponse, error) {
			bindCalled = true
			return nil, nil
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewUserService(repo, logger)

	res, err := svc.Validate(context.Background(), user)
	assert.ErrorIs(t, err, core.ErrUserNotRegistered)
	assert.Nil(t, res)
	assert.False(t, bindCalled)
}

func TestUserService_Validate_InvitedUser_AlreadyBoundDifferentAuthID_Rejected(t *testing.T) {
	internalID := uuid.New()
	authID := uuid.New()
	otherAuthID := uuid.New()
	user := core.AuthenticatedUser{
		ID:            authID,
		AuthID:        &authID,
		Email:         "invited@example.com",
		EmailVerified: true,
	}

	existing := &model.UserResponse{
		ID:         internalID,
		AuthUserID: &otherAuthID,
		Email:      "invited@example.com",
		Status:     "active",
	}

	bindCalled := false
	repo := &fakeUserRepo{
		getByEmailFn: func(c context.Context, email string) (*model.UserResponse, error) {
			return existing, nil
		},
		bindAuthIDAndActivateFn: func(c context.Context, id uuid.UUID, aID uuid.UUID) (*model.UserResponse, error) {
			bindCalled = true
			return nil, nil
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewUserService(repo, logger)

	res, err := svc.Validate(context.Background(), user)
	assert.ErrorIs(t, err, core.ErrUserNotRegistered)
	assert.Nil(t, res)
	assert.False(t, bindCalled)
}

func TestUserService_Validate_InvitedUser_FirstLogin_BindsAndActivates(t *testing.T) {
	internalID := uuid.New()
	authID := uuid.New()
	user := core.AuthenticatedUser{
		ID:            authID,
		AuthID:        &authID,
		Email:         "invited@example.com",
		EmailVerified: true,
	}

	existing := &model.UserResponse{
		ID:         internalID,
		AuthUserID: nil,
		Email:      "invited@example.com",
		Status:     "invited",
	}
	bound := &model.UserResponse{
		ID:         internalID,
		AuthUserID: &authID,
		Email:      "invited@example.com",
		Status:     "active",
	}

	bindCalled := false
	repo := &fakeUserRepo{
		getByEmailFn: func(c context.Context, email string) (*model.UserResponse, error) {
			return existing, nil
		},
		bindAuthIDAndActivateFn: func(c context.Context, id uuid.UUID, aID uuid.UUID) (*model.UserResponse, error) {
			bindCalled = true
			assert.Equal(t, internalID, id)
			assert.Equal(t, authID, aID)
			return bound, nil
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewUserService(repo, logger)

	res, err := svc.Validate(context.Background(), user)
	assert.NoError(t, err)
	assert.True(t, bindCalled)
	assert.Equal(t, internalID, res.ID)
	assert.Equal(t, "active", res.Status)
}

