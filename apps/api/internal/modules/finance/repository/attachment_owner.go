package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/transaction"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/attachment"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

// AttachmentOwnerTransaction is the attachment owner_type for transactions.
const AttachmentOwnerTransaction = "transaction"

// NewTransactionOwnerResolver resolves a transaction (scoped to its parish)
// into an attachment Owner. A transaction without a unit cannot have
// attachments because attachment permission is checked at unit scope.
func NewTransactionOwnerResolver(cli *ent.Client) attachment.OwnerResolver {
	return func(ctx context.Context, parishID int, id uuid.UUID) (*attachment.Owner, error) {
		tx, err := cli.Transaction.Query().
			Where(
				transaction.IDEQ(id),
				transaction.HasParishWith(parish.IDEQ(parishID)),
			).
			WithUnit().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, core.ErrItemNotFound
			}
			return nil, err
		}
		if tx.Edges.Unit == nil {
			return nil, model.ErrTransactionHasNoUnit
		}
		return &attachment.Owner{ID: tx.ID, ParishID: parishID, UnitID: tx.Edges.Unit.ID}, nil
	}
}

func loadTransactionAttachments(ctx context.Context, cli *ent.Client, resp *model.TransactionResponse) error {
	m, err := attachment.LoadByOwners(ctx, cli, AttachmentOwnerTransaction, []uuid.UUID{resp.ID})
	if err != nil {
		return err
	}
	if atts := m[resp.ID]; atts != nil {
		resp.Attachments = atts
	}
	return nil
}
