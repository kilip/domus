package service

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type AccountRepository interface {
	List(ctx context.Context, parishID int, filter model.ListAccountsQuery) ([]model.AccountResponse, error)
	GetByID(ctx context.Context, parishID int, id uuid.UUID) (*model.AccountResponse, error)
	GetEntByID(ctx context.Context, parishID int, id uuid.UUID) (*ent.Account, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	HasPostings(ctx context.Context, id uuid.UUID) (bool, error)
	HasDescendantPostings(ctx context.Context, id uuid.UUID) (bool, error)
	Create(ctx context.Context, parishID int, req model.CreateAccountRequest, normalBalance string) (*model.AccountResponse, error)
	Update(ctx context.Context, parishID int, id uuid.UUID, req model.UpdateAccountRequest, normalBalance string, updateChildrenType bool) (*model.AccountResponse, error)
	Delete(ctx context.Context, parishID int, id uuid.UUID) error
}

type AccountService struct {
	repo   AccountRepository
	entcli *ent.Client
	log    *slog.Logger
}

func NewAccountService(repo AccountRepository, entcli *ent.Client, log *slog.Logger) AccountService {
	return AccountService{repo: repo, entcli: entcli, log: log}
}

func (s AccountService) List(ctx context.Context, parishID int, filter model.ListAccountsQuery) ([]model.AccountResponse, error) {
	return s.repo.List(ctx, parishID, filter)
}

func (s AccountService) GetByID(ctx context.Context, parishID int, id uuid.UUID) (*model.AccountResponse, error) {
	return s.repo.GetByID(ctx, parishID, id)
}

func (s AccountService) Create(ctx context.Context, parishID int, req model.CreateAccountRequest) (*model.AccountResponse, error) {
	normalBalance := model.DeriveNormalBalance(req.Type)

	if req.ParentID != nil {
		parent, err := s.repo.GetEntByID(ctx, parishID, *req.ParentID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "parent account does not exist in this parish")
		}
		if parent.Postable {
			return nil, fiber.NewError(fiber.StatusBadRequest, "parent account must not be postable")
		}
		if string(parent.Type) != req.Type {
			return nil, fiber.NewError(fiber.StatusBadRequest, "child account type must match parent account type")
		}
	}

	exists, err := s.entcli.Unit.Query().
		Where(entunit.IDEQ(req.UnitID), entunit.HasParishWith(parish.IDEQ(parishID))).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fiber.NewError(fiber.StatusBadRequest, "unit belongs to a different parish or does not exist")
	}

	return s.repo.Create(ctx, parishID, req, normalBalance)
}

func (s AccountService) Update(ctx context.Context, parishID int, id uuid.UUID, req model.UpdateAccountRequest) (*model.AccountResponse, error) {
	existing, err := s.repo.GetEntByID(ctx, parishID, id)
	if err != nil {
		return nil, err
	}

	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, fiber.NewError(fiber.StatusBadRequest, "account cannot be its own parent")
		}

		parent, err := s.repo.GetEntByID(ctx, parishID, *req.ParentID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "parent account does not exist in this parish")
		}
		if parent.Postable {
			return nil, fiber.NewError(fiber.StatusBadRequest, "parent account must not be postable")
		}
		if string(parent.Type) != req.Type {
			return nil, fiber.NewError(fiber.StatusBadRequest, "child account type must match parent account type")
		}

		// Detect cycles: walk up parent ancestry
		currID := *req.ParentID
		visited := map[uuid.UUID]bool{id: true}
		for {
			if visited[currID] {
				return nil, fiber.NewError(fiber.StatusBadRequest, "cycle detected: account cannot have its descendant as parent")
			}
			visited[currID] = true

			curr, err := s.repo.GetEntByID(ctx, parishID, currID)
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
			return nil, fiber.NewError(fiber.StatusBadRequest, "cannot change type of account with postings")
		}
		if req.Postable != nil && *req.Postable != existing.Postable {
			return nil, fiber.NewError(fiber.StatusBadRequest, "cannot change postable of account with postings")
		}
	}

	updateChildrenType := false
	if string(existing.Type) != req.Type {
		hasDescPostings, err := s.repo.HasDescendantPostings(ctx, id)
		if err != nil {
			return nil, err
		}
		if hasDescPostings {
			return nil, fiber.NewError(fiber.StatusBadRequest, "cannot change type of account when descendant has postings")
		}
		updateChildrenType = true
	}

	if req.Postable != nil && *req.Postable && !existing.Postable {
		hasChildren, err := s.repo.HasChildren(ctx, id)
		if err != nil {
			return nil, err
		}
		if hasChildren {
			return nil, fiber.NewError(fiber.StatusBadRequest, "cannot make account postable when it has children")
		}
	}

	normalBalance := model.DeriveNormalBalance(req.Type)
	return s.repo.Update(ctx, parishID, id, req, normalBalance, updateChildrenType)
}

func (s AccountService) Delete(ctx context.Context, parishID int, id uuid.UUID) error {
	_, err := s.repo.GetEntByID(ctx, parishID, id)
	if err != nil {
		return err
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return fiber.NewError(fiber.StatusBadRequest, "cannot delete account with children; deactivate it instead")
	}

	hasPostings, err := s.repo.HasPostings(ctx, id)
	if err != nil {
		return err
	}
	if hasPostings {
		return fiber.NewError(fiber.StatusBadRequest, "cannot delete account with postings; deactivate it instead")
	}

	return s.repo.Delete(ctx, parishID, id)
}
