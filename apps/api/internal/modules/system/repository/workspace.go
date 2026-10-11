package repository

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/workspace"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/system/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type WorkspaceRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewWorkspaceRepository(entcli *ent.Client, log *slog.Logger) WorkspaceRepository {
	return WorkspaceRepository{entcli, log}
}

func toWorkspaceResponse(w *ent.Workspace) model.WorkspaceResponse {
	resp := model.WorkspaceResponse{
		ID:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		Status:      string(w.Status),
	}
	if w.Edges.Parish != nil {
		resp.ParishID = w.Edges.Parish.ID
		pResp := toParishResponse(w.Edges.Parish)
		resp.Parish = &pResp
	}
	return resp
}

func (r WorkspaceRepository) Search(c context.Context, req model.SearchWorkspaceRequest) (*model.PaginatedResult[model.WorkspaceResponse], error) {
	query := r.entcli.Workspace.Query().
		WithParish()

	if req.ParishID != nil {
		query = query.Where(workspace.HasParishWith(parish.IDEQ(*req.ParishID)))
	}
	if req.Status != nil && *req.Status != "" {
		query = query.Where(workspace.StatusEQ(workspace.Status(*req.Status)))
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
			query = query.Where(workspace.IDGT(cursorID))
		}
	}

	items, err := query.Order(ent.Asc(workspace.FieldID)).Limit(limit + 1).All(c)
	if err != nil {
		return nil, err
	}

	var nextCursor string
	if len(items) > limit {
		nextCursor = strconv.Itoa(items[limit-1].ID)
		items = items[:limit]
	}

	responses := make([]model.WorkspaceResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toWorkspaceResponse(item))
	}

	return &model.PaginatedResult[model.WorkspaceResponse]{
		Data:       responses,
		NextCursor: nextCursor,
	}, nil
}

func (r WorkspaceRepository) GetByID(c context.Context, id int) (*model.WorkspaceResponse, error) {
	item, err := r.entcli.Workspace.Query().
		Where(workspace.IDEQ(id)).
		WithParish().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	resp := toWorkspaceResponse(item)
	return &resp, nil
}

func (r WorkspaceRepository) GetEntByID(c context.Context, id int) (*ent.Workspace, error) {
	item, err := r.entcli.Workspace.Query().
		Where(workspace.IDEQ(id)).
		WithParish().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r WorkspaceRepository) Create(c context.Context, req model.CreateWorkspaceRequest) (*model.WorkspaceResponse, error) {
	var created *ent.Workspace
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		builder := tx.Workspace.Create().
			SetName(req.Name).
			SetParishID(req.ParishID)

		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		}
		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(workspace.Status(*req.Status))
		}

		w, err := builder.Save(c)
		if err != nil {
			return err
		}

		created, err = tx.Workspace.Query().
			Where(workspace.IDEQ(w.ID)).
			WithParish().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toWorkspaceResponse(created)
	return &resp, nil
}

func (r WorkspaceRepository) Update(c context.Context, id int, req model.UpdateWorkspaceRequest) (*model.WorkspaceResponse, error) {
	var updated *ent.Workspace
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Workspace.Query().Where(workspace.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}

		builder := tx.Workspace.UpdateOneID(id).
			SetName(req.Name).
			SetParishID(req.ParishID)

		if req.Description != nil && *req.Description != "" {
			builder.SetDescription(*req.Description)
		} else {
			builder.ClearDescription()
		}
		if req.Status != nil && *req.Status != "" {
			builder.SetStatus(workspace.Status(*req.Status))
		}

		_, err = builder.Save(c)
		if err != nil {
			return err
		}

		updated, err = tx.Workspace.Query().
			Where(workspace.IDEQ(id)).
			WithParish().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toWorkspaceResponse(updated)
	return &resp, nil
}

func (r WorkspaceRepository) Delete(c context.Context, id int) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Workspace.Query().Where(workspace.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}
		return tx.Workspace.DeleteOneID(id).Exec(c)
	})
}
