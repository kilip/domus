package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/entry"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/posting"
	"github.com/paroki/domus/api/ent/transaction"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

func (r TransactionRepository) Void(ctx context.Context, id uuid.UUID, reversalDate *time.Time, reason *string) (*model.TransactionResponse, error) {
	if r.voidMu != nil {
		r.voidMu.Lock()
		defer r.voidMu.Unlock()
	}

	fc, _ := model.FinanceContextFromContext(ctx)
	var resp *model.TransactionResponse

	err := database.WithTx(ctx, r.entcli, func(tx *ent.Tx) error {
		// Atomically move the transaction: Update().Where(IDEQ, parish & unit match, StatusEQ(posted)).SetStatus(void).Save
		// and require exactly 1 affected row, else 409 or 404.
		n, err := tx.Transaction.Update().
			Where(
				transaction.IDEQ(id),
				transaction.HasParishWith(parish.IDEQ(fc.ParishID)),
				transaction.HasUnitWith(entunit.IDEQ(fc.UnitID)),
				transaction.StatusEQ(transaction.StatusPosted),
			).
			SetStatus(transaction.StatusVoid).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			// Check if the transaction actually exists in this parish & unit
			txExists, err := tx.Transaction.Query().
				Where(
					transaction.IDEQ(id),
					transaction.HasParishWith(parish.IDEQ(fc.ParishID)),
					transaction.HasUnitWith(entunit.IDEQ(fc.UnitID)),
				).
				Exist(ctx)
			if err != nil {
				return err
			}
			if !txExists {
				return core.ErrItemNotFound
			}
			return model.ErrTxNotPosted
		}

		// 2. Find the active entry (the one that is not a reversal and has no reversed_by).
		activeEntry, err := tx.Entry.Query().
			Where(
				entry.HasTransactionWith(
					transaction.IDEQ(id),
					transaction.HasParishWith(parish.IDEQ(fc.ParishID)),
					transaction.HasUnitWith(entunit.IDEQ(fc.UnitID)),
				),
				entry.Not(entry.HasReverses()),
				entry.Not(entry.HasReversedBy()),
			).
			WithTransaction().
			WithPostings(func(pq *ent.PostingQuery) {
				pq.WithAccount()
			}).
			Only(ctx)
		if err != nil {
			return err
		}

		// Reversal date = request date, default = original entry date.
		revDate := activeEntry.Date
		if reversalDate != nil {
			revDate = *reversalDate
		}

		// Description: "Reversal of <number>" plus reason if given.
		desc := fmt.Sprintf("Reversal of %s", activeEntry.Edges.Transaction.Number)
		if reason != nil && *reason != "" {
			desc = fmt.Sprintf("Reversal of %s: %s", activeEntry.Edges.Transaction.Number, *reason)
		}

		// 3. Create the reversal Entry (reverses = original, date per D5)
		reversalEntry, err := tx.Entry.Create().
			SetTransactionID(id).
			SetDate(revDate).
			SetDescription(desc).
			SetReverses(activeEntry).
			Save(ctx)
		if err != nil {
			return err
		}

		// Mirror postings with sides swapped and identical amounts, accounts, and memos.
		postingBuilders := make([]*ent.PostingCreate, len(activeEntry.Edges.Postings))
		for i, p := range activeEntry.Edges.Postings {
			newSide := posting.SideDebit
			if p.Side == posting.SideDebit {
				newSide = posting.SideCredit
			}

			b := tx.Posting.Create().
				SetEntry(reversalEntry).
				SetAccountID(p.Edges.Account.ID).
				SetSide(newSide).
				SetAmount(p.Amount)
			if p.Memo != nil && *p.Memo != "" {
				b.SetMemo(*p.Memo)
			}
			postingBuilders[i] = b
		}

		if len(postingBuilders) > 0 {
			_, err = tx.Posting.CreateBulk(postingBuilders...).Save(ctx)
			if err != nil {
				return err
			}
		}

		// Reload full transaction response
		loaded, err := tx.Transaction.Query().
			Where(transaction.IDEQ(id)).
			WithParish().
			WithUnit().
			WithCreator().
			WithEntries(func(eq *ent.EntryQuery) {
				eq.WithPostings(func(pq *ent.PostingQuery) {
					pq.WithAccount()
				}).WithReverses()
			}).
			Only(ctx)
		if err != nil {
			return err
		}

		res := toTransactionResponse(loaded)
		if err := loadTransactionAttachments(ctx, tx.Client(), &res); err != nil {
			return err
		}
		resp = &res
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil
}
