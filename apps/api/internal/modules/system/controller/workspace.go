package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/modules/system/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type WorkspaceService interface {
	Search(c context.Context, req model.SearchWorkspaceRequest) (*model.PaginatedResult[model.WorkspaceResponse], error)
	GetByID(c context.Context, id int) (*model.WorkspaceResponse, error)
	Create(c context.Context, req model.CreateWorkspaceRequest) (*model.WorkspaceResponse, error)
	Update(c context.Context, id int, req model.UpdateWorkspaceRequest) (*model.WorkspaceResponse, error)
	Delete(c context.Context, id int) error
}

type WorkspaceController struct {
	svc WorkspaceService
}

func NewWorkspaceController(svc WorkspaceService) WorkspaceController {
	return WorkspaceController{svc}
}

// Search godoc
//
//	@Summary		Search workspaces
//	@Description	Search workspaces with filters and cursor pagination
//	@Tags			Workspace
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.SearchWorkspaceRequest	true	"Search workspace request"
//	@Success		200		{object}	httpx.WebResponse[[]model.WorkspaceResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/workspaces/search [post]
func (wc WorkspaceController) Search(c fiber.Ctx) error {
	var req model.SearchWorkspaceRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
	}
	res, err := wc.svc.Search(c, req)
	if err != nil {
		return err
	}
	return httpx.WithCursor(c, res.Data, res.NextCursor)
}

// GetByID godoc
//
//	@Summary		Get workspace by ID
//	@Description	Get workspace details by ID
//	@Tags			Workspace
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Workspace ID"
//	@Success		200	{object}	httpx.WebResponse[model.WorkspaceResponse]
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/workspaces/{id} [get]
func (wc WorkspaceController) GetByID(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	data, err := wc.svc.GetByID(c, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Create godoc
//
//	@Summary		Create workspace
//	@Description	Create a new workspace
//	@Tags			Workspace
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateWorkspaceRequest	true	"Create workspace request"
//	@Success		201		{object}	httpx.WebResponse[model.WorkspaceResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/workspaces [post]
func (wc WorkspaceController) Create(c fiber.Ctx) error {
	var req model.CreateWorkspaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := wc.svc.Create(c, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, data)
}

// Update godoc
//
//	@Summary		Update workspace
//	@Description	Update workspace by ID
//	@Tags			Workspace
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Workspace ID"
//	@Param			request	body		model.UpdateWorkspaceRequest	true	"Update workspace request"
//	@Success		200		{object}	httpx.WebResponse[model.WorkspaceResponse]
//	@Failure		400		{object}	httpx.ErrorResponse
//	@Failure		401		{object}	httpx.ErrorResponse
//	@Failure		403		{object}	httpx.ErrorResponse
//	@Failure		404		{object}	httpx.ErrorResponse
//	@Failure		422		{object}	httpx.ErrorResponse
//	@Router			/workspaces/{id} [put]
func (wc WorkspaceController) Update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req model.UpdateWorkspaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	data, err := wc.svc.Update(c, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, data)
}

// Delete godoc
//
//	@Summary		Delete workspace
//	@Description	Delete workspace by ID
//	@Tags			Workspace
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Workspace ID"
//	@Success		204
//	@Failure		400	{object}	httpx.ErrorResponse
//	@Failure		401	{object}	httpx.ErrorResponse
//	@Failure		403	{object}	httpx.ErrorResponse
//	@Failure		404	{object}	httpx.ErrorResponse
//	@Router			/workspaces/{id} [delete]
func (wc WorkspaceController) Delete(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := wc.svc.Delete(c, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
