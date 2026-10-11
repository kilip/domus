package repository

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/territory"
	"github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type UnitRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewUnitRepository(entcli *ent.Client, log *slog.Logger) UnitRepository {
	return UnitRepository{entcli, log}
}

func toUnitResponse(u *ent.Unit) model.UnitResponse {
	resp := model.UnitResponse{
		ID:          u.ID,
		Name:        u.Name,
		Code:        u.Code,
		Kind:        string(u.Kind),
		Description: u.Description,
		Status:      string(u.Status),
	}
	if u.Edges.Parish != nil {
		resp.ParishID = u.Edges.Parish.ID
		pResp := toParishResponse(u.Edges.Parish)
		resp.Parish = &pResp
	}
	if u.Edges.Parent != nil {
		resp.ParentID = &u.Edges.Parent.ID
		parentResp := model.UnitResponse{
			ID:          u.Edges.Parent.ID,
			Name:        u.Edges.Parent.Name,
			Code:        u.Edges.Parent.Code,
			Kind:        string(u.Edges.Parent.Kind),
			Description: u.Edges.Parent.Description,
			Status:      string(u.Edges.Parent.Status),
		}
		resp.Parent = &parentResp
	}
	if u.Edges.Territory != nil {
		resp.TerritoryID = &u.Edges.Territory.ID
		tResp := toTerritoryResponse(u.Edges.Territory)
		resp.Territory = &tResp
	}
	return resp
}

func (r UnitRepository) Search(c context.Context, req model.SearchUnitRequest) (*model.PaginatedResult[model.UnitResponse], error) {
	query := r.entcli.Unit.Query().
		WithParish().
		WithParent().
		WithTerritory()

	if req.ParishID != nil {
		query = query.Where(unit.HasParishWith(parish.IDEQ(*req.ParishID)))
	}
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			query = query.Where(unit.Not(unit.HasParent()))
		} else {
			query = query.Where(unit.HasParentWith(unit.IDEQ(*req.ParentID)))
		}
	}
	if req.TerritoryID != nil {
		query = query.Where(unit.HasTerritoryWith(territory.IDEQ(*req.TerritoryID)))
	}
	if req.Kind != nil && *req.Kind != "" {
		query = query.Where(unit.KindEQ(unit.Kind(*req.Kind)))
	}
	if req.Status != nil && *req.Status != "" {
		query = query.Where(unit.StatusEQ(unit.Status(*req.Status)))
	}

	limit := 20
	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
		if limit > 100 {
			limit = 100
		}
	}

	if req.Cursor != nil && *req.Cursor != "" {
		cursorID, err := strconv.Atoi(*req.Cursor)
		if err == nil {
			query = query.Where(unit.IDGT(cursorID))
		}
	}

	items, err := query.Order(ent.Asc(unit.FieldID)).Limit(limit + 1).All(c)
	if err != nil {
		return nil, err
	}

	var nextCursor string
	if len(items) > limit {
		nextCursor = strconv.Itoa(items[limit-1].ID)
		items = items[:limit]
	}

	responses := make([]model.UnitResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toUnitResponse(item))
	}

	return &model.PaginatedResult[model.UnitResponse]{
		Data:       responses,
		NextCursor: nextCursor,
	}, nil
}

func (r UnitRepository) GetByID(c context.Context, id int) (*model.UnitResponse, error) {
	item, err := r.entcli.Unit.Query().
		Where(unit.IDEQ(id)).
		WithParish().
		WithParent().
		WithTerritory().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	resp := toUnitResponse(item)
	return &resp, nil
}

func (r UnitRepository) GetEntByID(c context.Context, id int) (*ent.Unit, error) {
	item, err := r.entcli.Unit.Query().
		Where(unit.IDEQ(id)).
		WithParish().
		WithParent().
		WithTerritory().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r UnitRepository) HasChildren(c context.Context, id int) (bool, error) {
	return r.entcli.Unit.Query().
		Where(unit.HasParentWith(unit.IDEQ(id))).
		Exist(c)
}

func (r UnitRepository) Create(c context.Context, req model.CreateUnitRequest) (*model.UnitResponse, error) {
	var created *ent.Unit
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		builder := tx.Unit.Create().
			SetName(req.Name).
			SetKind(unit.Kind(req.Kind)).
			SetParishID(req.ParishID)

		if req.Code != nil && *req.Code != "" {
			builder.SetCode(*req.Code)
		}
		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		}
		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(unit.Status(*req.Status))
		}
		if req.ParentID != nil {
			builder.SetParentID(*req.ParentID)
		}
		if req.TerritoryID != nil {
			builder.SetTerritoryID(*req.TerritoryID)
		}

		u, err := builder.Save(c)
		if err != nil {
			return err
		}

		created, err = tx.Unit.Query().
			Where(unit.IDEQ(u.ID)).
			WithParish().
			WithParent().
			WithTerritory().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toUnitResponse(created)
	return &resp, nil
}

func (r UnitRepository) Update(c context.Context, id int, req model.UpdateUnitRequest) (*model.UnitResponse, error) {
	var updated *ent.Unit
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Unit.Query().Where(unit.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}

		builder := tx.Unit.UpdateOneID(id).
			SetName(req.Name).
			SetKind(unit.Kind(req.Kind)).
			SetParishID(req.ParishID)

		if req.Code != nil && *req.Code != "" {
			builder.SetCode(*req.Code)
		} else {
			builder.ClearCode()
		}
		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		} else {
			builder.ClearDescription()
		}
		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(unit.Status(*req.Status))
		}
		if req.ParentID != nil {
			builder.SetParentID(*req.ParentID)
		} else {
			builder.ClearParent()
		}
		if req.TerritoryID != nil {
			builder.SetTerritoryID(*req.TerritoryID)
		} else {
			builder.ClearTerritory()
		}

		_, err = builder.Save(c)
		if err != nil {
			return err
		}

		updated, err = tx.Unit.Query().
			Where(unit.IDEQ(id)).
			WithParish().
			WithParent().
			WithTerritory().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toUnitResponse(updated)
	return &resp, nil
}

func (r UnitRepository) Delete(c context.Context, id int) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Unit.Query().Where(unit.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}
		return tx.Unit.DeleteOneID(id).Exec(c)
	})
}

