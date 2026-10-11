package repository

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/posting"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type AccountRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewAccountRepository(entcli *ent.Client, log *slog.Logger) AccountRepository {
	return AccountRepository{entcli: entcli, log: log}
}

func toAccountResponse(acc *ent.Account) model.AccountResponse {
	resp := model.AccountResponse{
		ID:            acc.ID,
		Code:          acc.Code,
		Name:          acc.Name,
		Type:          string(acc.Type),
		NormalBalance: string(acc.NormalBalance),
		Postable:      acc.Postable,
		Status:        string(acc.Status),
		Description:   acc.Description,
		CreatedAt:     acc.CreatedAt,
		UpdatedAt:     acc.UpdatedAt,
	}
	if acc.Edges.Parish != nil {
		resp.ParishID = acc.Edges.Parish.ID
	}
	if acc.Edges.Parent != nil {
		resp.ParentID = &acc.Edges.Parent.ID
	}
	if acc.Edges.Unit != nil {
		resp.UnitID = acc.Edges.Unit.ID
	}
	return resp
}

func (r AccountRepository) List(ctx context.Context, filter model.ListAccountsQuery) ([]model.AccountResponse, error) {
	fc, _ := model.FinanceContextFromContext(ctx)
	q := r.entcli.Account.Query().
		Where(
			account.HasParishWith(parish.IDEQ(fc.ParishID)),
			account.HasUnitWith(entunit.IDEQ(fc.UnitID)),
		).
		WithParish().
		WithParent().
		WithUnit()

	if filter.Type != nil && *filter.Type != "" {
		q = q.Where(account.TypeEQ(account.Type(*filter.Type)))
	}
	if filter.Status != nil && *filter.Status != "" {
		q = q.Where(account.StatusEQ(account.Status(*filter.Status)))
	}
	if filter.Postable != nil {
		q = q.Where(account.PostableEQ(*filter.Postable))
	}

	accounts, err := q.Order(ent.Asc(account.FieldCode)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.AccountResponse, len(accounts))
	for i, acc := range accounts {
		result[i] = toAccountResponse(acc)
		if result[i].ParishID == 0 {
			result[i].ParishID = fc.ParishID
		}
	}
	return result, nil
}

func (r AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AccountResponse, error) {
	fc, _ := model.FinanceContextFromContext(ctx)
	acc, err := r.entcli.Account.Query().
		Where(
			account.IDEQ(id),
			account.HasParishWith(parish.IDEQ(fc.ParishID)),
			account.HasUnitWith(entunit.IDEQ(fc.UnitID)),
		).
		WithParish().
		WithParent().
		WithUnit().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	resp := toAccountResponse(acc)
	if resp.ParishID == 0 {
		resp.ParishID = fc.ParishID
	}
	return &resp, nil
}

func (r AccountRepository) GetEntByID(ctx context.Context, id uuid.UUID) (*ent.Account, error) {
	fc, _ := model.FinanceContextFromContext(ctx)
	acc, err := r.entcli.Account.Query().
		Where(
			account.IDEQ(id),
			account.HasParishWith(parish.IDEQ(fc.ParishID)),
			account.HasUnitWith(entunit.IDEQ(fc.UnitID)),
		).
		WithParish().
		WithParent().
		WithUnit().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	return acc, nil
}

func (r AccountRepository) GetEntsByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.Account, error) {
	fc, _ := model.FinanceContextFromContext(ctx)
	return r.entcli.Account.Query().
		Where(
			account.IDIn(ids...),
			account.HasParishWith(parish.IDEQ(fc.ParishID)),
			account.HasUnitWith(entunit.IDEQ(fc.UnitID)),
		).
		WithParish().
		WithParent().
		WithUnit().
		All(ctx)
}

func (r AccountRepository) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.entcli.Account.Query().
		Where(account.HasParentWith(account.IDEQ(id))).
		Exist(ctx)
}

func (r AccountRepository) HasPostings(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.entcli.Posting.Query().
		Where(posting.HasAccountWith(account.IDEQ(id))).
		Exist(ctx)
}

func (r AccountRepository) HasDescendantPostings(ctx context.Context, id uuid.UUID) (bool, error) {
	var descendants []uuid.UUID
	queue := []uuid.UUID{id}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		children, err := r.entcli.Account.Query().
			Where(account.HasParentWith(account.IDEQ(curr))).
			Select(account.FieldID).
			All(ctx)
		if err != nil {
			return false, err
		}
		for _, child := range children {
			descendants = append(descendants, child.ID)
			queue = append(queue, child.ID)
		}
	}
	if len(descendants) == 0 {
		return false, nil
	}
	return r.entcli.Posting.Query().
		Where(posting.HasAccountWith(account.IDIn(descendants...))).
		Exist(ctx)
}

func (r AccountRepository) updateDescendantsType(ctx context.Context, tx *ent.Tx, parentID uuid.UUID, accType account.Type, normalBal account.NormalBalance) error {
	children, err := tx.Account.Query().
		Where(account.HasParentWith(account.IDEQ(parentID))).
		All(ctx)
	if err != nil {
		return err
	}
	for _, child := range children {
		_, err := tx.Account.UpdateOneID(child.ID).
			SetType(accType).
			SetNormalBalance(normalBal).
			Save(ctx)
		if err != nil {
			return err
		}
		if err := r.updateDescendantsType(ctx, tx, child.ID, accType, normalBal); err != nil {
			return err
		}
	}
	return nil
}

func (r AccountRepository) Create(ctx context.Context, req model.CreateAccountRequest, normalBalance string) (*model.AccountResponse, error) {
	fc, _ := model.FinanceContextFromContext(ctx)
	var created *ent.Account
	err := database.WithTx(ctx, r.entcli, func(tx *ent.Tx) error {
		builder := tx.Account.Create().
			SetCode(req.Code).
			SetName(req.Name).
			SetType(account.Type(req.Type)).
			SetNormalBalance(account.NormalBalance(normalBalance))

		if req.Postable != nil {
			builder.SetPostable(*req.Postable)
		} else {
			builder.SetPostable(true)
		}

		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(account.Status(*req.Status))
		} else {
			builder.SetStatus(account.StatusActive)
		}

		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		}

		if req.ParentID != nil {
			builder.SetParentID(*req.ParentID)
		}

		// Set explicitly as well if in context, otherwise Ent hook sets it
		if fc.ParishID > 0 {
			builder.SetParishID(fc.ParishID)
		}
		if fc.UnitID > 0 {
			builder.SetUnitID(fc.UnitID)
		}

		acc, err := builder.Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				return fiber.NewError(fiber.StatusConflict, "account code already exists in this parish")
			}
			return err
		}

		created, err = tx.Account.Query().
			Where(account.IDEQ(acc.ID)).
			WithParish().
			WithParent().
			WithUnit().
			Only(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	resp := toAccountResponse(created)
	if resp.ParishID == 0 {
		resp.ParishID = fc.ParishID
	}
	return &resp, nil
}

func (r AccountRepository) Update(ctx context.Context, id uuid.UUID, req model.UpdateAccountRequest, normalBalance string, updateChildrenType bool) (*model.AccountResponse, error) {
	fc, _ := model.FinanceContextFromContext(ctx)
	var updated *ent.Account
	err := database.WithTx(ctx, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Account.Query().
			Where(
				account.IDEQ(id),
				account.HasParishWith(parish.IDEQ(fc.ParishID)),
				account.HasUnitWith(entunit.IDEQ(fc.UnitID)),
			).
			Exist(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}

		builder := tx.Account.UpdateOneID(id).
			SetCode(req.Code).
			SetName(req.Name).
			SetType(account.Type(req.Type)).
			SetNormalBalance(account.NormalBalance(normalBalance))

		if req.Postable != nil {
			builder.SetPostable(*req.Postable)
		}
		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(account.Status(*req.Status))
		}
		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		} else {
			builder.ClearDescription()
		}
		if req.ParentID != nil {
			builder.SetParentID(*req.ParentID)
		} else {
			builder.ClearParent()
		}

		_, err = builder.Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				return fiber.NewError(fiber.StatusConflict, "account code already exists in this parish")
			}
			return err
		}

		if updateChildrenType {
			if err := r.updateDescendantsType(ctx, tx, id, account.Type(req.Type), account.NormalBalance(normalBalance)); err != nil {
				return err
			}
		}

		updated, err = tx.Account.Query().
			Where(account.IDEQ(id)).
			WithParish().
			WithParent().
			WithUnit().
			Only(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	resp := toAccountResponse(updated)
	if resp.ParishID == 0 {
		resp.ParishID = fc.ParishID
	}
	return &resp, nil
}

func (r AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	fc, _ := model.FinanceContextFromContext(ctx)
	return database.WithTx(ctx, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Account.Query().
			Where(
				account.IDEQ(id),
				account.HasParishWith(parish.IDEQ(fc.ParishID)),
				account.HasUnitWith(entunit.IDEQ(fc.UnitID)),
			).
			Exist(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}
		return tx.Account.DeleteOneID(id).Exec(ctx)
	})
}
