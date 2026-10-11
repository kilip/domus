package service

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
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
		return nil, fiber.NewError(fiber.StatusBadRequest, "territorial unit requires a territory")
	}
	if req.Kind != "territorial" && req.TerritoryID != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "non-territorial unit cannot have a territory")
	}

	if req.TerritoryID != nil {
		t, err := s.territoryRepo.GetEntByID(c, *req.TerritoryID)
		if err != nil {
			return nil, err
		}
		if t.Edges.Parish == nil || t.Edges.Parish.ID != req.ParishID {
			return nil, fiber.NewError(fiber.StatusBadRequest, "territory belongs to a different parish")
		}
	}

	if req.ParentID != nil {
		parent, err := s.repo.GetEntByID(c, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Edges.Parish == nil || parent.Edges.Parish.ID != req.ParishID {
			return nil, fiber.NewError(fiber.StatusBadRequest, "parent unit belongs to a different parish")
		}
	}

	return s.repo.Create(c, req)
}

func (s UnitService) Update(c context.Context, id int, req model.UpdateUnitRequest) (*model.UnitResponse, error) {
	if req.Kind == "territorial" && req.TerritoryID == nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "territorial unit requires a territory")
	}
	if req.Kind != "territorial" && req.TerritoryID != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "non-territorial unit cannot have a territory")
	}

	if req.TerritoryID != nil {
		t, err := s.territoryRepo.GetEntByID(c, *req.TerritoryID)
		if err != nil {
			return nil, err
		}
		if t.Edges.Parish == nil || t.Edges.Parish.ID != req.ParishID {
			return nil, fiber.NewError(fiber.StatusBadRequest, "territory belongs to a different parish")
		}
	}

	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, fiber.NewError(fiber.StatusBadRequest, "unit cannot be its own parent")
		}

		parent, err := s.repo.GetEntByID(c, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Edges.Parish == nil || parent.Edges.Parish.ID != req.ParishID {
			return nil, fiber.NewError(fiber.StatusBadRequest, "parent unit belongs to a different parish")
		}

		// Detect cycles: walk up parent ancestry
		currID := *req.ParentID
		visited := map[int]bool{id: true}
		for {
			if visited[currID] {
				return nil, fiber.NewError(fiber.StatusBadRequest, "cycle detected: unit cannot have its descendant as parent")
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
		return fiber.NewError(fiber.StatusBadRequest, "cannot delete unit with children; deactivate it instead")
	}

	return s.repo.Delete(c, id)
}

