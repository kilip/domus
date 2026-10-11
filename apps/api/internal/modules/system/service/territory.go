package service

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/system/model"
)

type TerritoryRepository interface {
	Search(c context.Context, req model.SearchTerritoryRequest) (*model.PaginatedResult[model.TerritoryResponse], error)
	GetByID(c context.Context, id int) (*model.TerritoryResponse, error)
	GetEntByID(c context.Context, id int) (*ent.Territory, error)
	HasChildren(c context.Context, id int) (bool, error)
	Create(c context.Context, req model.CreateTerritoryRequest) (*model.TerritoryResponse, error)
	Update(c context.Context, id int, req model.UpdateTerritoryRequest) (*model.TerritoryResponse, error)
	Delete(c context.Context, id int) error
}

type TerritoryService struct {
	repo TerritoryRepository
	log  *slog.Logger
}

func NewTerritoryService(repo TerritoryRepository, log *slog.Logger) TerritoryService {
	return TerritoryService{repo, log}
}

func (s TerritoryService) Search(c context.Context, req model.SearchTerritoryRequest) (*model.PaginatedResult[model.TerritoryResponse], error) {
	return s.repo.Search(c, req)
}

func (s TerritoryService) GetByID(c context.Context, id int) (*model.TerritoryResponse, error) {
	return s.repo.GetByID(c, id)
}

func (s TerritoryService) Create(c context.Context, req model.CreateTerritoryRequest) (*model.TerritoryResponse, error) {
	if req.ParentID != nil {
		parent, err := s.repo.GetEntByID(c, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Edges.Parish == nil || parent.Edges.Parish.ID != req.ParishID {
			return nil, model.ErrTerritoryDifferentParish
		}
	}

	return s.repo.Create(c, req)
}

func (s TerritoryService) Update(c context.Context, id int, req model.UpdateTerritoryRequest) (*model.TerritoryResponse, error) {
	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, model.ErrTerritorySelfParent
		}

		parent, err := s.repo.GetEntByID(c, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Edges.Parish == nil || parent.Edges.Parish.ID != req.ParishID {
			return nil, model.ErrTerritoryDifferentParish
		}

		// Detect cycles: walk up parent ancestry
		currID := *req.ParentID
		visited := map[int]bool{id: true}
		for {
			if visited[currID] {
				return nil, model.ErrTerritoryCycleDetected
			}
			visited[currID] = true

			curr, err := s.repo.GetEntByID(c, currID)
			if err != nil {
				break
			}
			if curr.Edges.Parent == nil {
				break
			}
			currID = curr.Edges.Parent.ID
		}
	}

	return s.repo.Update(c, id, req)
}

func (s TerritoryService) Delete(c context.Context, id int) error {
	hasChildren, err := s.repo.HasChildren(c, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return model.ErrTerritoryDeleteWithChildren
	}

	return s.repo.Delete(c, id)
}
