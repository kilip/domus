package service

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/internal/modules/unit/model"
)

type ParishRepository interface {
	GetAll(c context.Context) ([]model.ParishResponse, error)
	GetByID(c context.Context, id int) (*model.ParishResponse, error)
	Create(c context.Context, req model.CreateParishRequest) (*model.ParishResponse, error)
	Update(c context.Context, id int, req model.UpdateParishRequest) (*model.ParishResponse, error)
	Delete(c context.Context, id int) error
}

type ParishService struct {
	repo ParishRepository
	log  *slog.Logger
}

func NewParishService(repo ParishRepository, log *slog.Logger) ParishService {
	return ParishService{repo, log}
}

func (s ParishService) GetAll(c context.Context) ([]model.ParishResponse, error) {
	return s.repo.GetAll(c)
}

func (s ParishService) GetByID(c context.Context, id int) (*model.ParishResponse, error) {
	return s.repo.GetByID(c, id)
}

func (s ParishService) Create(c context.Context, req model.CreateParishRequest) (*model.ParishResponse, error) {
	return s.repo.Create(c, req)
}

func (s ParishService) Update(c context.Context, id int, req model.UpdateParishRequest) (*model.ParishResponse, error) {
	return s.repo.Update(c, id, req)
}

func (s ParishService) Delete(c context.Context, id int) error {
	return s.repo.Delete(c, id)
}
