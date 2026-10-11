package repository

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type ParishRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewParishRepository(entcli *ent.Client, log *slog.Logger) ParishRepository {
	return ParishRepository{entcli, log}
}

func toParishResponse(p *ent.Parish) model.ParishResponse {
	resp := model.ParishResponse{
		ID:      p.ID,
		Name:    p.Name,
		Active:  p.Active,
		Church:  p.Church,
		Address: p.Address,
		Phone:   p.Phone,
		Email:   p.Email,
	}
	if p.Edges.Diocese != nil {
		dResp := toDioceseResponse(p.Edges.Diocese)
		resp.Diocese = &dResp
	}
	return resp
}

func (r ParishRepository) GetAll(c context.Context) ([]model.ParishResponse, error) {
	items, err := r.entcli.Parish.Query().
		WithDiocese(func(q *ent.DioceseQuery) {
			q.WithCreator().WithUpdater()
		}).
		All(c)
	if err != nil {
		return nil, err
	}

	responses := make([]model.ParishResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toParishResponse(item))
	}
	return responses, nil
}

func (r ParishRepository) GetByID(c context.Context, id int) (*model.ParishResponse, error) {
	item, err := r.entcli.Parish.Query().
		Where(parish.IDEQ(id)).
		WithDiocese(func(q *ent.DioceseQuery) {
			q.WithCreator().WithUpdater()
		}).
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	resp := toParishResponse(item)
	return &resp, nil
}

func (r ParishRepository) Create(c context.Context, req model.CreateParishRequest) (*model.ParishResponse, error) {
	var created *ent.Parish
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		builder := tx.Parish.Create().
			SetName(req.Name)
		if req.Active != nil {
			builder.SetActive(*req.Active)
		} else {
			builder.SetActive(true)
		}
		if req.Church != nil && *req.Church != "" {
			builder.SetChurch(*req.Church)
		}
		if req.Address != nil && *req.Address != "" {
			builder.SetAddress(*req.Address)
		}
		if req.Phone != nil && *req.Phone != "" {
			builder.SetPhone(*req.Phone)
		}
		if req.Email != nil && *req.Email != "" {
			builder.SetEmail(*req.Email)
		}
		if req.DioceseID != nil {
			builder.SetDioceseID(*req.DioceseID)
		}
		p, err := builder.Save(c)
		if err != nil {
			return err
		}

		created, err = tx.Parish.Query().
			Where(parish.IDEQ(p.ID)).
			WithDiocese(func(q *ent.DioceseQuery) {
				q.WithCreator().WithUpdater()
			}).
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toParishResponse(created)
	return &resp, nil
}

func (r ParishRepository) Update(c context.Context, id int, req model.UpdateParishRequest) (*model.ParishResponse, error) {
	var updated *ent.Parish
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Parish.Query().Where(parish.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}

		builder := tx.Parish.UpdateOneID(id).
			SetName(req.Name)
		if req.Active != nil {
			builder.SetActive(*req.Active)
		}
		if req.Church != nil && *req.Church != "" {
			builder.SetChurch(*req.Church)
		} else {
			builder.ClearChurch()
		}
		if req.Address != nil && *req.Address != "" {
			builder.SetAddress(*req.Address)
		} else {
			builder.ClearAddress()
		}
		if req.Phone != nil && *req.Phone != "" {
			builder.SetPhone(*req.Phone)
		} else {
			builder.ClearPhone()
		}
		if req.Email != nil && *req.Email != "" {
			builder.SetEmail(*req.Email)
		} else {
			builder.ClearEmail()
		}
		if req.DioceseID != nil {
			builder.SetDioceseID(*req.DioceseID)
		} else {
			builder.ClearDiocese()
		}
		_, err = builder.Save(c)
		if err != nil {
			return err
		}

		updated, err = tx.Parish.Query().
			Where(parish.IDEQ(id)).
			WithDiocese(func(q *ent.DioceseQuery) {
				q.WithCreator().WithUpdater()
			}).
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toParishResponse(updated)
	return &resp, nil
}

func (r ParishRepository) Delete(c context.Context, id int) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Parish.Query().Where(parish.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}
		return tx.Parish.DeleteOneID(id).Exec(c)
	})
}

