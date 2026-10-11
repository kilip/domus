package test

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type UnitSuite struct {
	testutil.ApiTestSuite[model.UnitResponse]
}

func (s *UnitSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.GrantRole(authz.RoleSuperadmin, authz.System)
}

func (s *UnitSuite) createParish(name string) model.ParishResponse {
	req := model.CreateParishRequest{Name: name}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	pResp := testutil.DecodeOther[model.ParishResponse](s)
	return pResp.Data
}

func (s *UnitSuite) createTerritory(parishID int, name string) model.TerritoryResponse {
	req := model.CreateTerritoryRequest{
		Name:     name,
		Kind:     "neighborhood",
		ParishID: parishID,
	}
	s.Request("/territories", fiber.MethodPost, req)
	s.Created()
	tResp := testutil.DecodeOther[model.TerritoryResponse](s)
	return tResp.Data
}

func (s *UnitSuite) TestCreate_Success() {
	parish := s.createParish("Paroki Dewan Paroki")
	code := "DPP"
	desc := "Dewan Pastoral Paroki"
	req := model.CreateUnitRequest{
		Name:        "Dewan Paroki Pleno",
		Code:        &code,
		Kind:        "council",
		Description: &desc,
		ParishID:    parish.ID,
	}

	s.Request("/units", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Dewan Paroki Pleno", res.Data.Name)
	s.Equal("council", res.Data.Kind)
	s.Equal("active", res.Data.Status)
	s.Equal(parish.ID, res.Data.ParishID)
	s.Nil(res.Data.TerritoryID)
}

func (s *UnitSuite) TestCreate_TerritorialKind_Success() {
	parish := s.createParish("Paroki Unit Teritorial")
	territory := s.createTerritory(parish.ID, "Lingkungan Santo Yosef")

	req := model.CreateUnitRequest{
		Name:        "Pengurus Lingkungan Santo Yosef",
		Kind:        "territorial",
		ParishID:    parish.ID,
		TerritoryID: &territory.ID,
	}

	s.Request("/units", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("territorial", res.Data.Kind)
	s.NotNil(res.Data.TerritoryID)
	s.Equal(territory.ID, *res.Data.TerritoryID)
	s.NotNil(res.Data.Territory)
	s.Equal(territory.Name, res.Data.Territory.Name)
}

func (s *UnitSuite) TestCreate_TerritorialWithoutTerritory_BadRequest() {
	parish := s.createParish("Paroki Unit Kurang Territory")

	req := model.CreateUnitRequest{
		Name:     "Pengurus Lingkungan",
		Kind:     "territorial",
		ParishID: parish.ID,
	}

	s.Request("/units", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestCreate_NonTerritorialWithTerritory_BadRequest() {
	parish := s.createParish("Paroki Unit Salah Territory")
	territory := s.createTerritory(parish.ID, "Lingkungan A")

	req := model.CreateUnitRequest{
		Name:        "Seksi Liturgi",
		Kind:        "section",
		ParishID:    parish.ID,
		TerritoryID: &territory.ID,
	}

	s.Request("/units", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestCreate_CrossParishTerritory_BadRequest() {
	parish1 := s.createParish("Paroki 1")
	parish2 := s.createParish("Paroki 2")
	territoryParish2 := s.createTerritory(parish2.ID, "Lingkungan P2")

	req := model.CreateUnitRequest{
		Name:        "Pengurus Lingkungan",
		Kind:        "territorial",
		ParishID:    parish1.ID,
		TerritoryID: &territoryParish2.ID,
	}

	s.Request("/units", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestCreate_CrossParishParent_BadRequest() {
	parish1 := s.createParish("Paroki A")
	parish2 := s.createParish("Paroki B")

	s.Request("/units", fiber.MethodPost, model.CreateUnitRequest{
		Name:     "Parent Unit",
		Kind:     "council",
		ParishID: parish1.ID,
	})
	s.Created()
	parent := s.GetResponse().Data

	req := model.CreateUnitRequest{
		Name:     "Child Unit",
		Kind:     "section",
		ParishID: parish2.ID,
		ParentID: &parent.ID,
	}

	s.Request("/units", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestCreate_ValidationError() {
	req := model.CreateUnitRequest{
		Name: "",
		Kind: "invalid-kind",
	}

	s.Request("/units", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *UnitSuite) TestSearch_Success_And_CursorPagination() {
	parish := s.createParish("Paroki Unit Search")

	for i := 1; i <= 5; i++ {
		req := model.CreateUnitRequest{
			Name:     fmt.Sprintf("Seksi %02d", i),
			Kind:     "section",
			ParishID: parish.ID,
		}
		s.Request("/units", fiber.MethodPost, req)
		s.Created()
	}

	limit := 2
	searchReq := model.SearchUnitRequest{
		ParishID: &parish.ID,
		Limit:    &limit,
	}
	s.Request("/units/search", fiber.MethodPost, searchReq)
	s.OK()

	res1 := s.PagedResponse()
	s.Len(res1.Data, 2)
	s.NotEmpty(res1.Meta.Cursor)

	cursor := res1.Meta.Cursor
	searchReq2 := model.SearchUnitRequest{
		ParishID: &parish.ID,
		Limit:    &limit,
		Cursor:   &cursor,
	}
	s.Request("/units/search", fiber.MethodPost, searchReq2)
	s.OK()

	res2 := s.PagedResponse()
	s.Len(res2.Data, 2)
	s.NotEmpty(res2.Meta.Cursor)
	s.NotEqual(res1.Data[0].ID, res2.Data[0].ID)
}

func (s *UnitSuite) TestGetByID_Success() {
	parish := s.createParish("Paroki Unit GetByID")
	req := model.CreateUnitRequest{
		Name:     "OMK Paroki",
		Kind:     "category",
		ParishID: parish.ID,
	}
	s.Request("/units", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/units/%d", created.ID), fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("OMK Paroki", res.Data.Name)
}

func (s *UnitSuite) TestGetByID_NotFound() {
	s.Request("/units/999999", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *UnitSuite) TestGetByID_InvalidID() {
	s.Request("/units/invalid-id", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestUpdate_Success() {
	parish := s.createParish("Paroki Unit Update")
	req := model.CreateUnitRequest{
		Name:     "Seksi Liturgi",
		Kind:     "section",
		ParishID: parish.ID,
	}
	s.Request("/units", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	updateReq := model.UpdateUnitRequest{
		Name:     "Seksi Liturgi & Musik",
		Kind:     "section",
		ParishID: parish.ID,
	}
	s.Request(fmt.Sprintf("/units/%d", created.ID), fiber.MethodPut, updateReq)
	s.OK()

	res := s.GetResponse()
	s.Equal("Seksi Liturgi & Musik", res.Data.Name)
}

func (s *UnitSuite) TestUpdate_SelfParent_BadRequest() {
	parish := s.createParish("Paroki Unit Self Parent")
	req := model.CreateUnitRequest{
		Name:     "Unit Utama",
		Kind:     "council",
		ParishID: parish.ID,
	}
	s.Request("/units", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	updateReq := model.UpdateUnitRequest{
		Name:     "Unit Utama",
		Kind:     "council",
		ParishID: parish.ID,
		ParentID: &created.ID,
	}
	s.Request(fmt.Sprintf("/units/%d", created.ID), fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestUpdate_CycleDetection_BadRequest() {
	parish := s.createParish("Paroki Unit Cycle")

	// Parent node
	s.Request("/units", fiber.MethodPost, model.CreateUnitRequest{
		Name:     "Council",
		Kind:     "council",
		ParishID: parish.ID,
	})
	s.Created()
	node1 := s.GetResponse().Data

	// Child node
	s.Request("/units", fiber.MethodPost, model.CreateUnitRequest{
		Name:     "Committee",
		Kind:     "committee",
		ParishID: parish.ID,
		ParentID: &node1.ID,
	})
	s.Created()
	node2 := s.GetResponse().Data

	// Set child as parent of parent
	updateReq := model.UpdateUnitRequest{
		Name:     node1.Name,
		Kind:     node1.Kind,
		ParishID: parish.ID,
		ParentID: &node2.ID,
	}
	s.Request(fmt.Sprintf("/units/%d", node1.ID), fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestDelete_Success() {
	parish := s.createParish("Paroki Unit Delete")
	req := model.CreateUnitRequest{
		Name:     "Panitia Paskah 2026",
		Kind:     "committee",
		ParishID: parish.ID,
	}
	s.Request("/units", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/units/%d", created.ID), fiber.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/units/%d", created.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *UnitSuite) TestDelete_HasChildren_BadRequest() {
	parish := s.createParish("Paroki Unit Delete Children")
	parentReq := model.CreateUnitRequest{
		Name:     "Dewan Paroki",
		Kind:     "council",
		ParishID: parish.ID,
	}
	s.Request("/units", fiber.MethodPost, parentReq)
	s.Created()
	parent := s.GetResponse().Data

	childReq := model.CreateUnitRequest{
		Name:     "Seksi Katekese",
		Kind:     "section",
		ParishID: parish.ID,
		ParentID: &parent.ID,
	}
	s.Request("/units", fiber.MethodPost, childReq)
	s.Created()

	// Deleting parent unit with child should fail
	s.Request(fmt.Sprintf("/units/%d", parent.ID), fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *UnitSuite) TestForbidden_NoPermission() {
	s.User.ID = core.GenerateID()

	s.Request("/units/search", fiber.MethodPost, model.SearchUnitRequest{})
	s.AssertStatus(fiber.StatusForbidden)
}

func TestUnitSuite(t *testing.T) {
	suite.Run(t, new(UnitSuite))
}
