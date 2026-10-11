package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type ParishSuite struct {
	testutil.ApiTestSuite[model.ParishResponse]
}

func (s *ParishSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.GrantRole(authz.RoleSuperadmin, authz.System)
}

func (s *ParishSuite) createDiocese(name string) model.DioceseResponse {
	dReq := model.CreateDioceseRequest{Name: name}
	s.Request("/dioceses", fiber.MethodPost, dReq)
	s.Created()
	dResp := testutil.DecodeOther[model.DioceseResponse](s)
	return dResp.Data
}

func (s *ParishSuite) TestCreate_Success() {
	diocese := s.createDiocese("Keuskupan Agung Jakarta")
	church := "Santa Maria Pelindung Diangkat Ke Surga"
	addr := "Jl. Katedral No. 7, Jakarta Pusat"
	phone := "021-3819999"
	email := "katedral@kaj.or.id"

	req := model.CreateParishRequest{
		Name:      "Paroki Katedral Jakarta",
		Church:    &church,
		Address:   &addr,
		Phone:     &phone,
		Email:     &email,
		DioceseID: &diocese.ID,
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Paroki Katedral Jakarta", res.Data.Name)
	s.True(res.Data.Active)
	s.NotNil(res.Data.Church)
	s.Equal(church, *res.Data.Church)
	s.NotNil(res.Data.Address)
	s.Equal(addr, *res.Data.Address)
	s.NotNil(res.Data.Phone)
	s.Equal(phone, *res.Data.Phone)
	s.NotNil(res.Data.Email)
	s.Equal(email, *res.Data.Email)
	s.NotNil(res.Data.Diocese)
	s.Equal(diocese.ID, res.Data.Diocese.ID)
	s.Equal("Keuskupan Agung Jakarta", res.Data.Diocese.Name)
}

func (s *ParishSuite) TestCreate_WithActiveFalse() {
	active := false
	req := model.CreateParishRequest{
		Name:   "Paroki Inaktif",
		Active: &active,
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Paroki Inaktif", res.Data.Name)
	s.False(res.Data.Active)
}

func (s *ParishSuite) TestCreate_WithoutDiocese_Success() {
	req := model.CreateParishRequest{
		Name: "Paroki Tanpa Keuskupan",
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Paroki Tanpa Keuskupan", res.Data.Name)
	s.Nil(res.Data.Church)
	s.Nil(res.Data.Address)
	s.Nil(res.Data.Phone)
	s.Nil(res.Data.Email)
	s.Nil(res.Data.Diocese)
}

func (s *ParishSuite) TestCreate_ValidationError() {
	req := model.CreateParishRequest{
		Name: "",
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *ParishSuite) TestCreate_InvalidEmail_ValidationError() {
	invalidEmail := "not-an-email"
	req := model.CreateParishRequest{
		Name:  "Paroki Email Salah",
		Email: &invalidEmail,
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *ParishSuite) TestGetAll_Success() {
	req1 := model.CreateParishRequest{Name: "Paroki Theresia"}
	s.Request("/parishes", fiber.MethodPost, req1)
	s.Created()

	req2 := model.CreateParishRequest{Name: "Paroki Blok B"}
	s.Request("/parishes", fiber.MethodPost, req2)
	s.Created()

	s.Request("/parishes", fiber.MethodGet, nil)
	s.OK()

	res := s.PagedResponse()
	s.GreaterOrEqual(len(res.Data), 2)
}

func (s *ParishSuite) TestGetByID_Success() {
	diocese := s.createDiocese("Keuskupan Bandung")

	req := model.CreateParishRequest{
		Name:      "Paroki Santo Petrus Bandung",
		DioceseID: &diocese.ID,
	}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Paroki Santo Petrus Bandung", res.Data.Name)
	s.NotNil(res.Data.Diocese)
	s.Equal(diocese.ID, res.Data.Diocese.ID)
}

func (s *ParishSuite) TestGetByID_NotFound() {
	s.Request("/parishes/999999", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestGetByID_InvalidID() {
	s.Request("/parishes/invalid-id", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *ParishSuite) TestUpdate_Success() {
	diocese1 := s.createDiocese("Keuskupan Surabaya")
	diocese2 := s.createDiocese("Keuskupan Malang")

	church1 := "Hati Kudus Yesus Surabaya"
	addr1 := "Jl. Polisi Istimewa No. 15"
	phone1 := "031-5678901"
	email1 := "hkysurabaya@sby.org"

	req := model.CreateParishRequest{
		Name:      "Paroki Hati Kudus Surabaya",
		Church:    &church1,
		Address:   &addr1,
		Phone:     &phone1,
		Email:     &email1,
		DioceseID: &diocese1.ID,
	}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	church2 := "Katedral Hati Kudus Yesus"
	addr2 := "Jl. Polisi Istimewa No. 17"
	phone2 := "031-5678999"
	email2 := "katedral@sby.org"

	activeFalse := false
	updateReq := model.UpdateParishRequest{
		Name:      "Paroki Hati Kudus Yesus",
		Active:    &activeFalse,
		Church:    &church2,
		Address:   &addr2,
		Phone:     &phone2,
		Email:     &email2,
		DioceseID: &diocese2.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodPut, updateReq)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Paroki Hati Kudus Yesus", res.Data.Name)
	s.False(res.Data.Active)
	s.NotNil(res.Data.Church)
	s.Equal(church2, *res.Data.Church)
	s.NotNil(res.Data.Address)
	s.Equal(addr2, *res.Data.Address)
	s.NotNil(res.Data.Phone)
	s.Equal(phone2, *res.Data.Phone)
	s.NotNil(res.Data.Email)
	s.Equal(email2, *res.Data.Email)
	s.NotNil(res.Data.Diocese)
	s.Equal(diocese2.ID, res.Data.Diocese.ID)
}

func (s *ParishSuite) TestUpdate_NotFound() {
	updateReq := model.UpdateParishRequest{Name: "Non Existent Parish"}
	s.Request("/parishes/999999", fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestDelete_Success() {
	req := model.CreateParishRequest{Name: "Paroki Santo Yoseph Denpasar"}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Verify that the record is soft-deleted in DB (still exists with deleted_at set)
	p, err := testutil.GetState().Ent.Parish.Get(core.SkipSoftDelete(context.Background()), created.ID)
	s.Require().NoError(err)
	s.Assert().NotNil(p.DeletedAt)
}

func (s *ParishSuite) TestDelete_NotFound() {
	s.Request("/parishes/999999", fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestForbidden_NoPermission() {
	// new user without any membership
	s.User.ID = core.GenerateID()

	s.Request("/parishes", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)
}

func TestParishSuite(t *testing.T) {
	suite.Run(t, new(ParishSuite))
}

