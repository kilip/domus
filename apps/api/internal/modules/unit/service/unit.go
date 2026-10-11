package service

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/unit/model"
)

type UnitRepository interface {
	Search(c context.Context, req model.SearchUnitRequest) (*model.PaginatedResult[model.UnitResponse], error)
	GetByID(c context.Context, id int) (*model.UnitResponse, error)
	GetEntByID(c context.Context, id int) (*ent.Unit, error)
	HasChildren(c context.Context, id int) (bool, error)
	Create(c context.Context, req model.CreateUnitRequest) (*model.UnitResponse, error)
	Update(c context.Context, id int, req model.UpdateUnitRequest) (*model.UnitResponse, error)
	Delete(c context.Context, id int) error
}

type TerritoryFinder interface {
	GetEntByID(c context.Context, id int) (*ent.Territory, error)
}

type UnitService struct {
	repo          UnitRepository
	territoryRepo TerritoryFinder
	log           *slog.Logger
}

func NewUnitService(repo UnitRepository, territoryRepo TerritoryFinder, log *slog.Logger) UnitService {
	return UnitService{repo, territoryRepo, log}
}

func (s UnitService) Search(c context.Context, req model.SearchUnitRequest) (*model.PaginatedResult[model.UnitResponse], error) {
	return s.repo.Search(c, req)
}

func (s UnitService) GetByID(c context.Context, id int) (*model.UnitResponse, error) {
	return s.repo.GetByID(c, id)
}

func (s UnitService) Create(c context.Context, req model.CreateUnitRequest) (*model.UnitResponse, error) {
	if req.Kind == "territorial" && req.TerritoryID == nil {
		return nil, model.ErrUnitRequiresTerritory
	}
	if req.Kind != "territorial" && req.TerritoryID != nil {
		return nil, model.ErrUnitNonTerritorialHasTerritory
	}

	if req.TerritoryID != nil {
		t, err := s.territoryRepo.GetEntByID(c, *req.TerritoryID)
		if err != nil {
			return nil, err
		}
		if t.Edges.Parish == nil || t.Edges.Parish.ID != req.ParishID {
			return nil, model.ErrUnitTerritoryDifferentParish
		}
	}

	if req.ParentID != nil {
		parent, err := s.repo.GetEntByID(c, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Edges.Parish == nil || parent.Edges.Parish.ID != req.ParishID {
			return nil, model.ErrUnitParentDifferentParish
		}
	}

	return s.repo.Create(c, req)
}

func (s UnitService) Update(c context.Context, id int, req model.UpdateUnitRequest) (*model.UnitResponse, error) {
	if req.Kind == "territorial" && req.TerritoryID == nil {
		return nil, model.ErrUnitRequiresTerritory
	}
	if req.Kind != "territorial" && req.TerritoryID != nil {
		return nil, model.ErrUnitNonTerritorialHasTerritory
	}

	if req.TerritoryID != nil {
		t, err := s.territoryRepo.GetEntByID(c, *req.TerritoryID)
		if err != nil {
			return nil, err
		}
		if t.Edges.Parish == nil || t.Edges.Parish.ID != req.ParishID {
			return nil, model.ErrUnitTerritoryDifferentParish
		}
	}

	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, model.ErrUnitSelfParent
		}

		parent, err := s.repo.GetEntByID(c, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Edges.Parish == nil || parent.Edges.Parish.ID != req.ParishID {
			return nil, model.ErrUnitParentDifferentParish
		}

		// Detect cycles: walk up parent ancestry
		currID := *req.ParentID
		visited := map[int]bool{id: true}
		for {
			if visited[currID] {
				return nil, model.ErrUnitCycleDetected
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

func (s UnitService) Delete(c context.Context, id int) error {
	hasChildren, err := s.repo.HasChildren(c, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return model.ErrUnitDeleteWithChildren
	}

	return s.repo.Delete(c, id)
}
