package repository

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/territory"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type TerritoryRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewTerritoryRepository(entcli *ent.Client, log *slog.Logger) TerritoryRepository {
	return TerritoryRepository{entcli, log}
}

func toTerritoryResponse(t *ent.Territory) model.TerritoryResponse {
	resp := model.TerritoryResponse{
		ID:          t.ID,
		Name:        t.Name,
		Code:        t.Code,
		Kind:        t.Kind,
		Description: t.Description,
		Status:      string(t.Status),
	}
	if t.Edges.Parish != nil {
		resp.ParishID = t.Edges.Parish.ID
		pResp := toParishResponse(t.Edges.Parish)
		resp.Parish = &pResp
	}
	if t.Edges.Parent != nil {
		resp.ParentID = &t.Edges.Parent.ID
		parentResp := model.TerritoryResponse{
			ID:          t.Edges.Parent.ID,
			Name:        t.Edges.Parent.Name,
			Code:        t.Edges.Parent.Code,
			Kind:        t.Edges.Parent.Kind,
			Description: t.Edges.Parent.Description,
			Status:      string(t.Edges.Parent.Status),
		}
		resp.Parent = &parentResp
	}
	if t.Edges.Unit != nil {
		resp.UnitID = &t.Edges.Unit.ID
	}
	return resp
}

func (r TerritoryRepository) Search(c context.Context, req model.SearchTerritoryRequest) (*model.PaginatedResult[model.TerritoryResponse], error) {
	query := r.entcli.Territory.Query().
		WithParish().
		WithParent().
		WithUnit()

	if req.ParishID != nil {
		query = query.Where(territory.HasParishWith(parish.IDEQ(*req.ParishID)))
	}
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			query = query.Where(territory.Not(territory.HasParent()))
		} else {
			query = query.Where(territory.HasParentWith(territory.IDEQ(*req.ParentID)))
		}
	}
	if req.Kind != nil && *req.Kind != "" {
		query = query.Where(territory.KindEQ(*req.Kind))
	}
	if req.Status != nil && *req.Status != "" {
		query = query.Where(territory.StatusEQ(territory.Status(*req.Status)))
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
			query = query.Where(territory.IDGT(cursorID))
		}
	}

	items, err := query.Order(ent.Asc(territory.FieldID)).Limit(limit + 1).All(c)
	if err != nil {
		return nil, err
	}

	var nextCursor string
	if len(items) > limit {
		nextCursor = strconv.Itoa(items[limit-1].ID)
		items = items[:limit]
	}

	responses := make([]model.TerritoryResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toTerritoryResponse(item))
	}

	return &model.PaginatedResult[model.TerritoryResponse]{
		Data:       responses,
		NextCursor: nextCursor,
	}, nil
}

func (r TerritoryRepository) GetByID(c context.Context, id int) (*model.TerritoryResponse, error) {
	item, err := r.entcli.Territory.Query().
		Where(territory.IDEQ(id)).
		WithParish().
		WithParent().
		WithUnit().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	resp := toTerritoryResponse(item)
	return &resp, nil
}

func (r TerritoryRepository) GetEntByID(c context.Context, id int) (*ent.Territory, error) {
	item, err := r.entcli.Territory.Query().
		Where(territory.IDEQ(id)).
		WithParish().
		WithParent().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r TerritoryRepository) HasChildren(c context.Context, id int) (bool, error) {
	return r.entcli.Territory.Query().
		Where(territory.HasParentWith(territory.IDEQ(id))).
		Exist(c)
}

func (r TerritoryRepository) Create(c context.Context, req model.CreateTerritoryRequest) (*model.TerritoryResponse, error) {
	var created *ent.Territory
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		builder := tx.Territory.Create().
			SetName(req.Name).
			SetKind(req.Kind).
			SetParishID(req.ParishID)

		if req.Code != nil && *req.Code != "" {
			builder.SetCode(*req.Code)
		}
		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		}
		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(territory.Status(*req.Status))
		}
		if req.ParentID != nil {
			builder.SetParentID(*req.ParentID)
		}

		t, err := builder.Save(c)
		if err != nil {
			return err
		}

		created, err = tx.Territory.Query().
			Where(territory.IDEQ(t.ID)).
			WithParish().
			WithParent().
			WithUnit().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toTerritoryResponse(created)
	return &resp, nil
}

func (r TerritoryRepository) Update(c context.Context, id int, req model.UpdateTerritoryRequest) (*model.TerritoryResponse, error) {
	var updated *ent.Territory
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Territory.Query().Where(territory.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}

		builder := tx.Territory.UpdateOneID(id).
			SetName(req.Name).
			SetKind(req.Kind).
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
			builder.SetStatus(territory.Status(*req.Status))
		}
		if req.ParentID != nil {
			builder.SetParentID(*req.ParentID)
		} else {
			builder.ClearParent()
		}

		_, err = builder.Save(c)
		if err != nil {
			return err
		}

		updated, err = tx.Territory.Query().
			Where(territory.IDEQ(id)).
			WithParish().
			WithParent().
			WithUnit().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toTerritoryResponse(updated)
	return &resp, nil
}

func (r TerritoryRepository) Delete(c context.Context, id int) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Territory.Query().Where(territory.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}
		return tx.Territory.DeleteOneID(id).Exec(c)
	})
}

