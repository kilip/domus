package controller

import (
	"context"
	"mime/multipart"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type AttachmentService interface {
	Upload(ctx context.Context, parishID int, transactionID uuid.UUID, fileHeader *multipart.FileHeader) (*model.AttachmentResponse, error)
	List(ctx context.Context, parishID int, transactionID uuid.UUID) ([]model.AttachmentResponse, error)
	GetDownloadURL(ctx context.Context, parishID int, transactionID uuid.UUID, attachmentID uuid.UUID) (*model.DownloadResponse, error)
}

type AttachmentController struct {
	svc AttachmentService
}

func NewAttachmentController(svc AttachmentService) AttachmentController {
	return AttachmentController{svc: svc}
}

// Upload godoc
//
//	@Summary		Upload transaction attachment
//	@Description	Upload a file attachment for a transaction (PDF, JPEG, PNG, max 10 MiB)
//	@Tags			Finance Attachment
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId	path		int		true	"Parish ID"
//	@Param			id			path		string	true	"Transaction ID"
//	@Param			file		formData	file	true	"Attachment file"
//	@Success		201			{object}	httpx.WebResponse[model.AttachmentResponse]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Failure		413			{object}	httpx.ErrorResponse
//	@Failure		422			{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/transactions/{id}/attachments [post]
func (ac AttachmentController) Upload(c fiber.Ctx) error {
	parishID, err := httpx.GetIntParam(c, "parishId")
	if err != nil {
		return err
	}

	txID, err := httpx.GetID(c)
	if err != nil {
		return err
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file field is required")
	}

	resp, err := ac.svc.Upload(c, parishID, *txID, fileHeader)
	if err != nil {
		return err
	}

	return httpx.Created(c, resp)
}

// List godoc
//
//	@Summary		List transaction attachments
//	@Description	List all attachments for a transaction
//	@Tags			Finance Attachment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId	path		int		true	"Parish ID"
//	@Param			id			path		string	true	"Transaction ID"
//	@Success		200			{object}	httpx.WebResponse[[]model.AttachmentResponse]
//	@Failure		400			{object}	httpx.ErrorResponse
//	@Failure		401			{object}	httpx.ErrorResponse
//	@Failure		403			{object}	httpx.ErrorResponse
//	@Failure		404			{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/transactions/{id}/attachments [get]
func (ac AttachmentController) List(c fiber.Ctx) error {
	parishID, err := httpx.GetIntParam(c, "parishId")
	if err != nil {
		return err
	}

	txID, err := httpx.GetID(c)
	if err != nil {
		return err
	}

	resp, err := ac.svc.List(c, parishID, *txID)
	if err != nil {
		return err
	}

	return httpx.OK(c, resp)
}

// Download godoc
//
//	@Summary		Get presigned download URL for attachment
//	@Description	Generate a presigned GET URL valid for 5 minutes to download an attachment
//	@Tags			Finance Attachment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parishId		path		int		true	"Parish ID"
//	@Param			id				path		string	true	"Transaction ID"
//	@Param			attachmentId	path		string	true	"Attachment ID"
//	@Success		200				{object}	httpx.WebResponse[model.DownloadResponse]
//	@Failure		400				{object}	httpx.ErrorResponse
//	@Failure		401				{object}	httpx.ErrorResponse
//	@Failure		403				{object}	httpx.ErrorResponse
//	@Failure		404				{object}	httpx.ErrorResponse
//	@Router			/parishes/{parishId}/finance/transactions/{id}/attachments/{attachmentId}/download [get]
func (ac AttachmentController) Download(c fiber.Ctx) error {
	parishID, err := httpx.GetIntParam(c, "parishId")
	if err != nil {
		return err
	}

	txID, err := httpx.GetID(c)
	if err != nil {
		return err
	}

	attIDStr := c.Params("attachmentId", "undefined")
	if attIDStr == "undefined" {
		return core.ErrInvalidID
	}
	attID, err := uuid.Parse(attIDStr)
	if err != nil {
		return core.ErrInvalidID
	}

	resp, err := ac.svc.GetDownloadURL(c, parishID, *txID, attID)
	if err != nil {
		return err
	}

	return httpx.OK(c, resp)
}
