package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/diocese"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/system/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type DioceseRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewDioceseRepository(entcli *ent.Client, log *slog.Logger) DioceseRepository {
	return DioceseRepository{entcli, log}
}

func toCoreUser(u *ent.User) core.User {
	if u == nil {
		return core.User{}
	}
	return core.User{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Avatar: u.Avatar,
	}
}

func toDioceseResponse(d *ent.Diocese) model.DioceseResponse {
	var creator, updater core.User
	if d.Edges.Creator != nil {
		creator = toCoreUser(d.Edges.Creator)
	}
	if d.Edges.Updater != nil {
		updater = toCoreUser(d.Edges.Updater)
	}

	return model.DioceseResponse{
		ID:        d.ID,
		Name:      d.Name,
		Active:    d.Active,
		CreatedBy: creator,
		CreatedAt: d.CreatedAt,
		UpdatedBy: updater,
		UpdatedAt: d.UpdatedAt,
	}
}

func (r DioceseRepository) GetAll(c context.Context) ([]model.DioceseResponse, error) {
	items, err := r.entcli.Diocese.Query().
		WithCreator().
		WithUpdater().
		All(c)
	if err != nil {
		return nil, err
	}

	responses := make([]model.DioceseResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toDioceseResponse(item))
	}
	return responses, nil
}

func (r DioceseRepository) GetByID(c context.Context, id int) (*model.DioceseResponse, error) {
	item, err := r.entcli.Diocese.Query().
		Where(diocese.IDEQ(id)).
		WithCreator().
		WithUpdater().
		Only(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	resp := toDioceseResponse(item)
	return &resp, nil
}

func (r DioceseRepository) Create(c context.Context, creatorID uuid.UUID, req model.CreateDioceseRequest) (*model.DioceseResponse, error) {
	var created *ent.Diocese
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		builder := tx.Diocese.Create().
			SetName(req.Name).
			SetCreatedBy(creatorID).
			SetUpdatedBy(creatorID)
		if req.Active != nil {
			builder.SetActive(*req.Active)
		} else {
			builder.SetActive(true)
		}
		d, err := builder.Save(c)
		if err != nil {
			return err
		}

		created, err = tx.Diocese.Query().
			Where(diocese.IDEQ(d.ID)).
			WithCreator().
			WithUpdater().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toDioceseResponse(created)
	return &resp, nil
}

func (r DioceseRepository) Update(c context.Context, updaterID uuid.UUID, id int, req model.UpdateDioceseRequest) (*model.DioceseResponse, error) {
	var updated *ent.Diocese
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Diocese.Query().Where(diocese.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}

		builder := tx.Diocese.UpdateOneID(id).
			SetName(req.Name).
			SetUpdatedBy(updaterID)
		if req.Active != nil {
			builder.SetActive(*req.Active)
		}
		_, err = builder.Save(c)
		if err != nil {
			return err
		}

		updated, err = tx.Diocese.Query().
			Where(diocese.IDEQ(id)).
			WithCreator().
			WithUpdater().
			Only(c)
		return err
	})

	if err != nil {
		return nil, err
	}

	resp := toDioceseResponse(updated)
	return &resp, nil
}

func (r DioceseRepository) Delete(c context.Context, id int) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		exists, err := tx.Diocese.Query().Where(diocese.IDEQ(id)).Exist(c)
		if err != nil {
			return err
		}
		if !exists {
			return core.ErrItemNotFound
		}
		return tx.Diocese.DeleteOneID(id).Exec(c)
	})
}
