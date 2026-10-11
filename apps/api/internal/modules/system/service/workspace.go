package service

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/system/model"
)

type WorkspaceRepository interface {
	Search(c context.Context, req model.SearchWorkspaceRequest) (*model.PaginatedResult[model.WorkspaceResponse], error)
	GetByID(c context.Context, id int) (*model.WorkspaceResponse, error)
	GetEntByID(c context.Context, id int) (*ent.Workspace, error)
	Create(c context.Context, req model.CreateWorkspaceRequest) (*model.WorkspaceResponse, error)
	Update(c context.Context, id int, req model.UpdateWorkspaceRequest) (*model.WorkspaceResponse, error)
	Delete(c context.Context, id int) error
}

type WorkspaceService struct {
	repo WorkspaceRepository
	log  *slog.Logger
}

func NewWorkspaceService(repo WorkspaceRepository, log *slog.Logger) WorkspaceService {
	return WorkspaceService{repo, log}
}

func (s WorkspaceService) Search(c context.Context, req model.SearchWorkspaceRequest) (*model.PaginatedResult[model.WorkspaceResponse], error) {
	return s.repo.Search(c, req)
}

func (s WorkspaceService) GetByID(c context.Context, id int) (*model.WorkspaceResponse, error) {
	return s.repo.GetByID(c, id)
}

func (s WorkspaceService) Create(c context.Context, req model.CreateWorkspaceRequest) (*model.WorkspaceResponse, error) {
	return s.repo.Create(c, req)
}

func (s WorkspaceService) Update(c context.Context, id int, req model.UpdateWorkspaceRequest) (*model.WorkspaceResponse, error) {
	return s.repo.Update(c, id, req)
}

func (s WorkspaceService) Delete(c context.Context, id int) error {
	return s.repo.Delete(c, id)
}
