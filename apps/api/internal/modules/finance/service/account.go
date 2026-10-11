package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type AccountRepository interface {
	List(ctx context.Context, filter model.ListAccountsQuery) ([]model.AccountResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.AccountResponse, error)
	GetEntByID(ctx context.Context, id uuid.UUID) (*ent.Account, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	HasPostings(ctx context.Context, id uuid.UUID) (bool, error)
	HasDescendantPostings(ctx context.Context, id uuid.UUID) (bool, error)
	Create(ctx context.Context, req model.CreateAccountRequest, normalBalance string) (*model.AccountResponse, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateAccountRequest, normalBalance string, updateChildrenType bool) (*model.AccountResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AccountService struct {
	repo AccountRepository
	log  *slog.Logger
}

func NewAccountService(repo AccountRepository, log *slog.Logger) AccountService {
	return AccountService{repo: repo, log: log}
}

func (s AccountService) List(ctx context.Context, filter model.ListAccountsQuery) ([]model.AccountResponse, error) {
	return s.repo.List(ctx, filter)
}

func (s AccountService) GetByID(ctx context.Context, id uuid.UUID) (*model.AccountResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s AccountService) Create(ctx context.Context, req model.CreateAccountRequest) (*model.AccountResponse, error) {
	normalBalance := model.DeriveNormalBalance(req.Type)

	if req.ParentID != nil {
		parent, err := s.repo.GetEntByID(ctx, *req.ParentID)
		if err != nil {
			return nil, model.ErrParentAccountNotExist
		}
		if parent.Postable {
			return nil, model.ErrParentAccountPostable
		}
		if string(parent.Type) != req.Type {
			return nil, model.ErrAccountTypeMismatch
		}
	}

	return s.repo.Create(ctx, req, normalBalance)
}

func (s AccountService) Update(ctx context.Context, id uuid.UUID, req model.UpdateAccountRequest) (*model.AccountResponse, error) {
	existing, err := s.repo.GetEntByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, model.ErrAccountSelfParent
		}

		parent, err := s.repo.GetEntByID(ctx, *req.ParentID)
		if err != nil {
			return nil, model.ErrParentAccountNotExist
		}
		if parent.Postable {
			return nil, model.ErrParentAccountPostable
		}
		if string(parent.Type) != req.Type {
			return nil, model.ErrAccountTypeMismatch
		}

		// Detect cycles: walk up parent ancestry
		currID := *req.ParentID
		visited := map[uuid.UUID]bool{id: true}
		for {
			if visited[currID] {
				return nil, model.ErrAccountCycleDetected
			}
			visited[currID] = true

			curr, err := s.repo.GetEntByID(ctx, currID)
			if err != nil {
				break
			}
			if curr.Edges.Parent == nil {
				break
			}
			currID = curr.Edges.Parent.ID
		}
	}

	hasPostings, err := s.repo.HasPostings(ctx, id)
	if err != nil {
		return nil, err
	}

	if hasPostings {
		if string(existing.Type) != req.Type {
			return nil, model.ErrAccountTypeWithPostings
		}
		if req.Postable != nil && *req.Postable != existing.Postable {
			return nil, model.ErrAccountPostableWithPostings
		}
	}

	updateChildrenType := false
	if string(existing.Type) != req.Type {
		hasDescPostings, err := s.repo.HasDescendantPostings(ctx, id)
		if err != nil {
			return nil, err
		}
		if hasDescPostings {
			return nil, model.ErrAccountDescendantHasPostings
		}
		updateChildrenType = true
	}

	if req.Postable != nil && *req.Postable && !existing.Postable {
		hasChildren, err := s.repo.HasChildren(ctx, id)
		if err != nil {
			return nil, err
		}
		if hasChildren {
			return nil, model.ErrAccountPostableWithChildren
		}
	}

	normalBalance := model.DeriveNormalBalance(req.Type)
	return s.repo.Update(ctx, id, req, normalBalance, updateChildrenType)
}

func (s AccountService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetEntByID(ctx, id)
	if err != nil {
		return err
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return model.ErrAccountDeleteWithChildren
	}

	hasPostings, err := s.repo.HasPostings(ctx, id)
	if err != nil {
		return err
	}
	if hasPostings {
		return model.ErrAccountDeleteWithPostings
	}

	return s.repo.Delete(ctx, id)
}
