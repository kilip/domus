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

type WorkspaceSuite struct {
	testutil.ApiTestSuite[model.WorkspaceResponse]
}

func (s *WorkspaceSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.GrantRole(authz.RoleSuperadmin, authz.System)
}

func (s *WorkspaceSuite) createParish(name string) model.ParishResponse {
	req := model.CreateParishRequest{Name: name}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	pResp := testutil.DecodeOther[model.ParishResponse](s)
	return pResp.Data
}

func (s *WorkspaceSuite) TestCreate_Success() {
	parish := s.createParish("Paroki Dewan Paroki")
	desc := "Dewan Pastoral Paroki"
	req := model.CreateWorkspaceRequest{
		Name:        "Dewan Paroki Pleno",
		Description: &desc,
		ParishID:    parish.ID,
	}

	s.Request("/workspaces", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Dewan Paroki Pleno", res.Data.Name)
	s.Equal("active", res.Data.Status)
	s.Equal(parish.ID, res.Data.ParishID)
	s.Equal(&desc, res.Data.Description)
}

func (s *WorkspaceSuite) TestCreate_ValidationError() {
	req := model.CreateWorkspaceRequest{
		Name: "", // missing required name
	}

	s.Request("/workspaces", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *WorkspaceSuite) TestSearch_Success_And_CursorPagination() {
	parish := s.createParish("Paroki Workspace Search")

	for i := 1; i <= 5; i++ {
		req := model.CreateWorkspaceRequest{
			Name:     fmt.Sprintf("Workspace %02d", i),
			ParishID: parish.ID,
		}
		s.Request("/workspaces", fiber.MethodPost, req)
		s.Created()
	}

	limit := 2
	searchReq := model.SearchWorkspaceRequest{
		ParishID: &parish.ID,
		Limit:    &limit,
	}
	s.Request("/workspaces/search", fiber.MethodPost, searchReq)
	s.OK()

	listResp := testutil.DecodeOther[[]model.WorkspaceResponse](s)
	s.Len(listResp.Data, 2)
	s.NotEmpty(listResp.Meta.Cursor)

	// Fetch next page
	searchReq2 := model.SearchWorkspaceRequest{
		ParishID: &parish.ID,
		Limit:    &limit,
		Cursor:   &listResp.Meta.Cursor,
	}
	s.Request("/workspaces/search", fiber.MethodPost, searchReq2)
	s.OK()

	listResp2 := testutil.DecodeOther[[]model.WorkspaceResponse](s)
	s.Len(listResp2.Data, 2)
}

func (s *WorkspaceSuite) TestGetByID_Success() {
	parish := s.createParish("Paroki Workspace GetByID")
	req := model.CreateWorkspaceRequest{
		Name:     "Workspace Get Test",
		ParishID: parish.ID,
	}
	s.Request("/workspaces", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/workspaces/%d", created.ID), fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Workspace Get Test", res.Data.Name)
	s.Equal(parish.ID, res.Data.ParishID)
	s.NotNil(res.Data.Parish)
	s.Equal(parish.Name, res.Data.Parish.Name)
}

func (s *WorkspaceSuite) TestGetByID_NotFound() {
	s.Request("/workspaces/999999", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *WorkspaceSuite) TestGetByID_InvalidID() {
	s.Request("/workspaces/invalid-id", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *WorkspaceSuite) TestUpdate_Success() {
	parish := s.createParish("Paroki Workspace Update")
	req := model.CreateWorkspaceRequest{
		Name:     "Initial Name",
		ParishID: parish.ID,
	}
	s.Request("/workspaces", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	newDesc := "Updated Description"
	status := "inactive"
	updateReq := model.UpdateWorkspaceRequest{
		Name:        "Updated Name",
		Description: &newDesc,
		Status:      &status,
		ParishID:    parish.ID,
	}

	s.Request(fmt.Sprintf("/workspaces/%d", created.ID), fiber.MethodPut, updateReq)
	s.OK()

	res := s.GetResponse()
	s.Equal("Updated Name", res.Data.Name)
	s.Equal("inactive", res.Data.Status)
	s.Equal(&newDesc, res.Data.Description)
}

func (s *WorkspaceSuite) TestDelete_Success() {
	parish := s.createParish("Paroki Workspace Delete")
	req := model.CreateWorkspaceRequest{
		Name:     "To Delete",
		ParishID: parish.ID,
	}
	s.Request("/workspaces", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/workspaces/%d", created.ID), fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusNoContent)

	s.Request(fmt.Sprintf("/workspaces/%d", created.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *WorkspaceSuite) TestForbidden_NoPermission() {
	// new user without any membership
	s.User.ID = core.GenerateID()

	s.Request("/workspaces/search", fiber.MethodPost, model.SearchWorkspaceRequest{})
	s.AssertStatus(fiber.StatusForbidden)
}

func TestWorkspaceSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceSuite))
}
