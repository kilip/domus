package test

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/system/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type TerritorySuite struct {
	testutil.ApiTestSuite[model.TerritoryResponse]
}

func (s *TerritorySuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.GrantRole(authz.RoleSuperadmin, authz.System)
}

func (s *TerritorySuite) createParish(name string) model.ParishResponse {
	req := model.CreateParishRequest{Name: name}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	pResp := testutil.DecodeOther[model.ParishResponse](s)
	return pResp.Data
}

func (s *TerritorySuite) TestCreate_Success() {
	parish := s.createParish("Paroki Katedral")
	code := "WIL-01"
	desc := "Wilayah Santo Petrus"
	req := model.CreateTerritoryRequest{
		Name:        "Wilayah 1",
		Code:        &code,
		Kind:        "region",
		Description: &desc,
		ParishID:    parish.ID,
	}

	s.Request("/territories", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Wilayah 1", res.Data.Name)
	s.NotNil(res.Data.Code)
	s.Equal(code, *res.Data.Code)
	s.Equal("region", res.Data.Kind)
	s.Equal("active", res.Data.Status)
	s.Equal(parish.ID, res.Data.ParishID)
	s.Nil(res.Data.ParentID)
}

func (s *TerritorySuite) TestCreate_WithParent_Success() {
	parish := s.createParish("Paroki Theresia")
	parentReq := model.CreateTerritoryRequest{
		Name:     "Wilayah A",
		Kind:     "region",
		ParishID: parish.ID,
	}
	s.Request("/territories", fiber.MethodPost, parentReq)
	s.Created()
	parent := s.GetResponse().Data

	childReq := model.CreateTerritoryRequest{
		Name:     "Lingkungan 1",
		Kind:     "neighborhood",
		ParishID: parish.ID,
		ParentID: &parent.ID,
	}
	s.Request("/territories", fiber.MethodPost, childReq)
	s.Created()

	child := s.GetResponse().Data
	s.NotNil(child.ParentID)
	s.Equal(parent.ID, *child.ParentID)
	s.NotNil(child.Parent)
	s.Equal("Wilayah A", child.Parent.Name)
}

func (s *TerritorySuite) TestCreate_CrossParishParent_BadRequest() {
	parish1 := s.createParish("Paroki Satu")
	parish2 := s.createParish("Paroki Dua")

	parentReq := model.CreateTerritoryRequest{
		Name:     "Wilayah Parish 1",
		Kind:     "region",
		ParishID: parish1.ID,
	}
	s.Request("/territories", fiber.MethodPost, parentReq)
	s.Created()
	parent := s.GetResponse().Data

	childReq := model.CreateTerritoryRequest{
		Name:     "Lingkungan Parish 2",
		Kind:     "neighborhood",
		ParishID: parish2.ID,
		ParentID: &parent.ID,
	}
	s.Request("/territories", fiber.MethodPost, childReq)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *TerritorySuite) TestCreate_ValidationError() {
	req := model.CreateTerritoryRequest{
		Name: "",
		Kind: "region",
	}
	s.Request("/territories", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *TerritorySuite) TestSearch_Success_And_CursorPagination() {
	parish := s.createParish("Paroki Paginated")

	for i := 1; i <= 5; i++ {
		req := model.CreateTerritoryRequest{
			Name:     fmt.Sprintf("Lingkungan %02d", i),
			Kind:     "neighborhood",
			ParishID: parish.ID,
		}
		s.Request("/territories", fiber.MethodPost, req)
		s.Created()
	}

	limit := 2
	searchReq := model.SearchTerritoryRequest{
		ParishID: &parish.ID,
		Limit:    &limit,
	}
	s.Request("/territories/search", fiber.MethodPost, searchReq)
	s.OK()

	res1 := s.PagedResponse()
	s.Len(res1.Data, 2)
	s.NotEmpty(res1.Meta.Cursor)

	// Fetch next page with cursor
	cursor := res1.Meta.Cursor
	searchReq2 := model.SearchTerritoryRequest{
		ParishID: &parish.ID,
		Limit:    &limit,
		Cursor:   &cursor,
	}
	s.Request("/territories/search", fiber.MethodPost, searchReq2)
	s.OK()

	res2 := s.PagedResponse()
	s.Len(res2.Data, 2)
	s.NotEmpty(res2.Meta.Cursor)
	s.NotEqual(res1.Data[0].ID, res2.Data[0].ID)
}

func (s *TerritorySuite) TestGetByID_Success() {
	parish := s.createParish("Paroki GetByID")
	req := model.CreateTerritoryRequest{
		Name:     "Stasi Yohanes",
		Kind:     "station",
		ParishID: parish.ID,
	}
	s.Request("/territories", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/territories/%d", created.ID), fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Stasi Yohanes", res.Data.Name)
}

func (s *TerritorySuite) TestGetByID_NotFound() {
	s.Request("/territories/999999", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *TerritorySuite) TestGetByID_InvalidID() {
	s.Request("/territories/invalid-id", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *TerritorySuite) TestUpdate_Success() {
	parish := s.createParish("Paroki Update")
	req := model.CreateTerritoryRequest{
		Name:     "Wilayah Awal",
		Kind:     "region",
		ParishID: parish.ID,
	}
	s.Request("/territories", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	updateReq := model.UpdateTerritoryRequest{
		Name:     "Wilayah Diperbarui",
		Kind:     "region",
		ParishID: parish.ID,
	}
	s.Request(fmt.Sprintf("/territories/%d", created.ID), fiber.MethodPut, updateReq)
	s.OK()

	res := s.GetResponse()
	s.Equal("Wilayah Diperbarui", res.Data.Name)
}

func (s *TerritorySuite) TestUpdate_SelfParent_BadRequest() {
	parish := s.createParish("Paroki Self Parent")
	req := model.CreateTerritoryRequest{
		Name:     "Wilayah A",
		Kind:     "region",
		ParishID: parish.ID,
	}
	s.Request("/territories", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	updateReq := model.UpdateTerritoryRequest{
		Name:     "Wilayah A",
		Kind:     "region",
		ParishID: parish.ID,
		ParentID: &created.ID,
	}
	s.Request(fmt.Sprintf("/territories/%d", created.ID), fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *TerritorySuite) TestUpdate_CycleDetection_BadRequest() {
	parish := s.createParish("Paroki Cycle")

	// Node 1
	s.Request("/territories", fiber.MethodPost, model.CreateTerritoryRequest{
		Name:     "Node 1",
		Kind:     "region",
		ParishID: parish.ID,
	})
	s.Created()
	node1 := s.GetResponse().Data

	// Node 2 (child of Node 1)
	s.Request("/territories", fiber.MethodPost, model.CreateTerritoryRequest{
		Name:     "Node 2",
		Kind:     "region",
		ParishID: parish.ID,
		ParentID: &node1.ID,
	})
	s.Created()
	node2 := s.GetResponse().Data

	// Set Node 2 as parent of Node 1 -> cycle!
	updateReq := model.UpdateTerritoryRequest{
		Name:     node1.Name,
		Kind:     node1.Kind,
		ParishID: parish.ID,
		ParentID: &node2.ID,
	}
	s.Request(fmt.Sprintf("/territories/%d", node1.ID), fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *TerritorySuite) TestDelete_Success() {
	parish := s.createParish("Paroki Delete")
	req := model.CreateTerritoryRequest{
		Name:     "Wilayah Hapus",
		Kind:     "region",
		ParishID: parish.ID,
	}
	s.Request("/territories", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/territories/%d", created.ID), fiber.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/territories/%d", created.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *TerritorySuite) TestDelete_HasChildren_BadRequest() {
	parish := s.createParish("Paroki Delete Children")
	parentReq := model.CreateTerritoryRequest{
		Name:     "Wilayah Induk",
		Kind:     "region",
		ParishID: parish.ID,
	}
	s.Request("/territories", fiber.MethodPost, parentReq)
	s.Created()
	parent := s.GetResponse().Data

	childReq := model.CreateTerritoryRequest{
		Name:     "Lingkungan Anak",
		Kind:     "neighborhood",
		ParishID: parish.ID,
		ParentID: &parent.ID,
	}
	s.Request("/territories", fiber.MethodPost, childReq)
	s.Created()

	// Deleting parent territory with child should fail
	s.Request(fmt.Sprintf("/territories/%d", parent.ID), fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *TerritorySuite) TestForbidden_NoPermission() {
	s.User.ID = core.GenerateID()

	s.Request("/territories/search", fiber.MethodPost, model.SearchTerritoryRequest{})
	s.AssertStatus(fiber.StatusForbidden)
}

func TestTerritorySuite(t *testing.T) {
	suite.Run(t, new(TerritorySuite))
}
