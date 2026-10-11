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

type DioceseSuite struct {
	testutil.ApiTestSuite[model.DioceseResponse]
}

func (s *DioceseSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.GrantRole(authz.RoleSuperadmin, authz.System)
}

func (s *DioceseSuite) TestCreate_Success() {
	req := model.CreateDioceseRequest{
		Name: "Keuskupan Agung Jakarta",
	}

	s.Request("/dioceses", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Keuskupan Agung Jakarta", res.Data.Name)
	s.True(res.Data.Active)
	s.Equal(s.User.ID, res.Data.CreatedBy.ID)
	s.Equal(s.User.Name, res.Data.CreatedBy.Name)
	s.Equal(s.User.ID, res.Data.UpdatedBy.ID)
	s.False(res.Data.CreatedAt.IsZero())
	s.False(res.Data.UpdatedAt.IsZero())
}

func (s *DioceseSuite) TestCreate_WithActiveFalse() {
	active := false
	req := model.CreateDioceseRequest{
		Name:   "Keuskupan Inaktif",
		Active: &active,
	}

	s.Request("/dioceses", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Keuskupan Inaktif", res.Data.Name)
	s.False(res.Data.Active)
}

func (s *DioceseSuite) TestCreate_ValidationError() {
	req := model.CreateDioceseRequest{
		Name: "",
	}

	s.Request("/dioceses", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *DioceseSuite) TestGetAll_Success() {
	req1 := model.CreateDioceseRequest{Name: "Keuskupan Bogor"}
	s.Request("/dioceses", fiber.MethodPost, req1)
	s.Created()

	req2 := model.CreateDioceseRequest{Name: "Keuskupan Bandung"}
	s.Request("/dioceses", fiber.MethodPost, req2)
	s.Created()

	s.Request("/dioceses", fiber.MethodGet, nil)
	s.OK()

	res := s.PagedResponse()
	s.GreaterOrEqual(len(res.Data), 2)
}

func (s *DioceseSuite) TestGetByID_Success() {
	req := model.CreateDioceseRequest{Name: "Keuskupan Surabaya"}
	s.Request("/dioceses", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/dioceses/%d", created.ID), fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Keuskupan Surabaya", res.Data.Name)
	s.Equal(s.User.ID, res.Data.CreatedBy.ID)
}

func (s *DioceseSuite) TestGetByID_NotFound() {
	s.Request("/dioceses/999999", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *DioceseSuite) TestGetByID_InvalidID() {
	s.Request("/dioceses/invalid-id", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *DioceseSuite) TestUpdate_Success() {
	req := model.CreateDioceseRequest{Name: "Keuskupan Malang"}
	s.Request("/dioceses", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	activeFalse := false
	updateReq := model.UpdateDioceseRequest{
		Name:   "Keuskupan Malang Updated",
		Active: &activeFalse,
	}
	s.Request(fmt.Sprintf("/dioceses/%d", created.ID), fiber.MethodPut, updateReq)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Keuskupan Malang Updated", res.Data.Name)
	s.False(res.Data.Active)
	s.Equal(s.User.ID, res.Data.UpdatedBy.ID)
}

func (s *DioceseSuite) TestUpdate_NotFound() {
	updateReq := model.UpdateDioceseRequest{Name: "Non Existent Diocese"}
	s.Request("/dioceses/999999", fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *DioceseSuite) TestDelete_Success() {
	req := model.CreateDioceseRequest{Name: "Keuskupan Denpasar"}
	s.Request("/dioceses", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/dioceses/%d", created.ID), fiber.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/dioceses/%d", created.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *DioceseSuite) TestDelete_NotFound() {
	s.Request("/dioceses/999999", fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *DioceseSuite) TestForbidden_NoPermission() {
	// new user without any membership
	s.User.ID = core.GenerateID()

	s.Request("/dioceses", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)
}

func TestDioceseSuite(t *testing.T) {
	suite.Run(t, new(DioceseSuite))
}
