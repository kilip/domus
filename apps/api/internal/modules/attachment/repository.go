package attachment

import (
	"context"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	entattachment "github.com/paroki/domus/api/ent/attachment"
	"github.com/paroki/domus/api/internal/core"
)

type CreateInput struct {
	ID        uuid.UUID
	Bucket    string
	ObjectKey string
	Filename  string
	MimeType  string
	Size      int64
	Checksum  string
}

type Repository struct {
	entcli *ent.Client
}

func NewRepository(entcli *ent.Client) Repository {
	return Repository{entcli: entcli}
}

func (r Repository) Create(ctx context.Context, ownerType string, owner Owner, in CreateInput) (*Response, error) {
	created, err := r.entcli.Attachment.Create().
		SetID(in.ID).
		SetOwnerType(ownerType).
		SetOwnerID(owner.ID).
		SetParishID(owner.ParishID).
		SetUnitID(owner.UnitID).
		SetBucket(in.Bucket).
		SetObjectKey(in.ObjectKey).
		SetFilename(in.Filename).
		SetMimeType(in.MimeType).
		SetSize(in.Size).
		SetChecksum(in.Checksum).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	loaded, err := r.entcli.Attachment.Query().
		Where(entattachment.IDEQ(created.ID)).
		WithCreator().
		WithUnit().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	resp := toResponse(loaded)
	return &resp, nil
}

func (r Repository) ListByOwner(ctx context.Context, ownerType string, owner Owner) ([]Response, error) {
	m, err := LoadByOwners(ctx, r.entcli, ownerType, []uuid.UUID{owner.ID})
	if err != nil {
		return nil, err
	}
	res := m[owner.ID]
	if res == nil {
		res = []Response{}
	}
	return res, nil
}

func (r Repository) GetByID(ctx context.Context, ownerType string, owner Owner, id uuid.UUID) (*ent.Attachment, error) {
	att, err := r.entcli.Attachment.Query().
		Where(
			entattachment.IDEQ(id),
			entattachment.OwnerTypeEQ(ownerType),
			entattachment.OwnerIDEQ(owner.ID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	return att, nil
}

// LoadByOwners returns attachments grouped by owner ID. It lets owning
// modules embed attachments in their own responses without a hard edge.
func LoadByOwners(ctx context.Context, cli *ent.Client, ownerType string, ownerIDs []uuid.UUID) (map[uuid.UUID][]Response, error) {
	out := make(map[uuid.UUID][]Response, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return out, nil
	}
	rows, err := cli.Attachment.Query().
		Where(
			entattachment.OwnerTypeEQ(ownerType),
			entattachment.OwnerIDIn(ownerIDs...),
		).
		WithCreator().
		WithUnit().
		Order(ent.Asc(entattachment.FieldCreatedAt), ent.Asc(entattachment.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range rows {
		out[a.OwnerID] = append(out[a.OwnerID], toResponse(a))
	}
	return out, nil
}

func toResponse(a *ent.Attachment) Response {
	resp := Response{
		ID:        a.ID,
		OwnerType: a.OwnerType,
		OwnerID:   a.OwnerID,
		Filename:  a.Filename,
		MimeType:  a.MimeType,
		Size:      a.Size,
		Checksum:  a.Checksum,
		CreatedAt: a.CreatedAt,
	}
	if a.Edges.Unit != nil {
		resp.UnitID = a.Edges.Unit.ID
	}
	if a.Edges.Creator != nil {
		resp.CreatedBy = core.User{
			ID:     a.Edges.Creator.ID,
			Name:   a.Edges.Creator.Name,
			Email:  a.Edges.Creator.Email,
			Avatar: a.Edges.Creator.Avatar,
		}
	}
	return resp
}
