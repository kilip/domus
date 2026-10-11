package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/modules/system/model"
)

type DioceseRepository interface {
	GetAll(c context.Context) ([]model.DioceseResponse, error)
	GetByID(c context.Context, id int) (*model.DioceseResponse, error)
	Create(c context.Context, creatorID uuid.UUID, req model.CreateDioceseRequest) (*model.DioceseResponse, error)
	Update(c context.Context, updaterID uuid.UUID, id int, req model.UpdateDioceseRequest) (*model.DioceseResponse, error)
	Delete(c context.Context, id int) error
}

type DioceseService struct {
	repo DioceseRepository
	log  *slog.Logger
}

func NewDioceseService(repo DioceseRepository, log *slog.Logger) DioceseService {
	return DioceseService{repo, log}
}

func (s DioceseService) GetAll(c context.Context) ([]model.DioceseResponse, error) {
	return s.repo.GetAll(c)
}

func (s DioceseService) GetByID(c context.Context, id int) (*model.DioceseResponse, error) {
	return s.repo.GetByID(c, id)
}

func (s DioceseService) Create(c context.Context, creatorID uuid.UUID, req model.CreateDioceseRequest) (*model.DioceseResponse, error) {
	return s.repo.Create(c, creatorID, req)
}

func (s DioceseService) Update(c context.Context, updaterID uuid.UUID, id int, req model.UpdateDioceseRequest) (*model.DioceseResponse, error) {
	return s.repo.Update(c, updaterID, id, req)
}

func (s DioceseService) Delete(c context.Context, id int) error {
	return s.repo.Delete(c, id)
}
