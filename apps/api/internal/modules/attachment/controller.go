package attachment

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type Controller struct {
	svc Service
}

func NewController(svc Service) Controller {
	return Controller{svc: svc}
}

// Upload godoc
//
//	@Summary		Upload attachment
//	@Description	Upload a file (PDF, JPEG, PNG, max 10 MiB) attached to an owner record, e.g. a finance transaction. Permission is checked at the owner's unit scope.
//	@Tags			Attachment
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId	path		int		true	"Parish ID"
//	@Param			id			path		string	true	"Owner (transaction) ID"
//	@Param			file		formData	file	true	"Attachment file"
//	@Success		201			{object}	httpx.WebResponse[attachment.Response]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Failure		413			{object}	httpx.ErrorResponse
//	@Failure		422			{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/transactions/{id}/attachments [post]
func (ctl Controller) Upload(ownerType string) fiber.Handler {
	return func(c fiber.Ctx) error {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return ErrFileRequired
		}
		resp, err := ctl.svc.Upload(c, ownerType, ownerFrom(c), fileHeader)
		if err != nil {
			return err
		}
		return httpx.Created(c, resp)
	}
}

// List godoc
//
//	@Summary		List attachments
//	@Description	List all attachments of an owner record
//	@Tags			Attachment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId	path		int		true	"Parish ID"
//	@Param			id			path		string	true	"Owner (transaction) ID"
//	@Success		200			{object}	httpx.WebResponse[[]attachment.Response]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/transactions/{id}/attachments [get]
func (ctl Controller) List(ownerType string) fiber.Handler {
	return func(c fiber.Ctx) error {
		resp, err := ctl.svc.List(c, ownerType, ownerFrom(c))
		if err != nil {
			return err
		}
		return httpx.OK(c, resp)
	}
}

// Download godoc
//
//	@Summary		Get presigned download URL for attachment
//	@Description	Generate a presigned GET URL valid for 5 minutes to download an attachment
//	@Tags			Attachment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId		path		int		true	"Parish ID"
//	@Param			id				path		string	true	"Owner (transaction) ID"
//	@Param			attachmentId	path		string	true	"Attachment ID"
//	@Success		200				{object}	httpx.WebResponse[attachment.DownloadResponse]
//	@Failure		400				{object}	httpx.ErrorResponse
//	@Failure		401				{object}	httpx.ErrorResponse
//	@Failure		403				{object}	httpx.ErrorResponse
//	@Failure		404				{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/transactions/{id}/attachments/{attachmentId}/download [get]
func (ctl Controller) Download(ownerType string) fiber.Handler {
	return func(c fiber.Ctx) error {
		attID, err := uuid.Parse(c.Params("attachmentId"))
		if err != nil {
			return core.ErrInvalidID
		}
		resp, err := ctl.svc.GetDownloadURL(c, ownerType, ownerFrom(c), attID)
		if err != nil {
			return err
		}
		return httpx.OK(c, resp)
	}
}

func parseParish(c fiber.Ctx, param string) (int, error) {
	return httpx.GetIntParam(c, param)
}
