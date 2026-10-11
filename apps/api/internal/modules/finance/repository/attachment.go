package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/attachment"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/transaction"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type CreateAttachmentInput struct {
	ID        uuid.UUID
	Bucket    string
	ObjectKey string
	Filename  string
	MimeType  string
	Size      int64
	Checksum  string
}

type AttachmentRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewAttachmentRepository(entcli *ent.Client, log *slog.Logger) AttachmentRepository {
	return AttachmentRepository{
		entcli: entcli,
		log:    log,
	}
}

func (r AttachmentRepository) Create(ctx context.Context, parishID int, transactionID uuid.UUID, input CreateAttachmentInput) (*model.AttachmentResponse, error) {
	// Verify transaction belongs to parish
	txExists, err := r.entcli.Transaction.Query().
		Where(
			transaction.IDEQ(transactionID),
			transaction.HasParishWith(parish.IDEQ(parishID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !txExists {
		return nil, core.ErrItemNotFound
	}

	created, err := r.entcli.Attachment.Create().
		SetID(input.ID).
		SetTransactionID(transactionID).
		SetBucket(input.Bucket).
		SetObjectKey(input.ObjectKey).
		SetFilename(input.Filename).
		SetMimeType(input.MimeType).
		SetSize(input.Size).
		SetChecksum(input.Checksum).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	loaded, err := r.entcli.Attachment.Query().
		Where(attachment.IDEQ(created.ID)).
		WithCreator().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	resp := toAttachmentResponse(loaded, transactionID)
	return &resp, nil
}

func (r AttachmentRepository) ListByTransaction(ctx context.Context, parishID int, transactionID uuid.UUID) ([]model.AttachmentResponse, error) {
	// Verify transaction belongs to parish
	txExists, err := r.entcli.Transaction.Query().
		Where(
			transaction.IDEQ(transactionID),
			transaction.HasParishWith(parish.IDEQ(parishID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !txExists {
		return nil, core.ErrItemNotFound
	}

	attachments, err := r.entcli.Attachment.Query().
		Where(attachment.HasTransactionWith(
			transaction.IDEQ(transactionID),
			transaction.HasParishWith(parish.IDEQ(parishID)),
		)).
		WithCreator().
		Order(ent.Asc(attachment.FieldCreatedAt), ent.Asc(attachment.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]model.AttachmentResponse, len(attachments))
	for i, a := range attachments {
		res[i] = toAttachmentResponse(a, transactionID)
	}
	return res, nil
}

func (r AttachmentRepository) GetByID(ctx context.Context, parishID int, transactionID uuid.UUID, attachmentID uuid.UUID) (*ent.Attachment, error) {
	att, err := r.entcli.Attachment.Query().
		Where(
			attachment.IDEQ(attachmentID),
			attachment.HasTransactionWith(
				transaction.IDEQ(transactionID),
				transaction.HasParishWith(parish.IDEQ(parishID)),
			),
		).
		WithCreator().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	return att, nil
}

func toAttachmentResponse(a *ent.Attachment, txID uuid.UUID) model.AttachmentResponse {
	resp := model.AttachmentResponse{
		ID:            a.ID,
		TransactionID: txID,
		Filename:      a.Filename,
		MimeType:      a.MimeType,
		Size:          a.Size,
		Checksum:      a.Checksum,
		CreatedAt:     a.CreatedAt,
	}
	if a.Edges.Creator != nil {
		resp.CreatedBy = toCoreUser(a.Edges.Creator)
	}
	return resp
}
