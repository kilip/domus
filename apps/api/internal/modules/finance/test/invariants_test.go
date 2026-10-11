package test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	ententry "github.com/paroki/domus/api/ent/entry"
	entparish "github.com/paroki/domus/api/ent/parish"
	entposting "github.com/paroki/domus/api/ent/posting"
	"github.com/paroki/domus/api/ent/transaction"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InvariantsSuite struct {
	testutil.ApiTestSuite[any]
}

func (s *InvariantsSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *InvariantsSuite) grantRole(parishID int, role string) {
	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parishID)}
	s.GrantRole(role, scope)
}

func (s *InvariantsSuite) activate(parishID, unitID int) {
	SetFinanceCookie(&s.ApiTestSuite, parishID, unitID)
}

func (s *InvariantsSuite) createAccountWithUnit(parishID int, unitID int, code, name, accType string, postable bool) *ent.Account {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	acc, err := st.Ent.Account.Create().
		SetCode(code).
		SetName(name).
		SetType(account.Type(accType)).
		SetNormalBalance(account.NormalBalance(model.DeriveNormalBalance(accType))).
		SetPostable(postable).
		SetStatus(account.StatusActive).
		SetParishID(parishID).
		SetUnitID(unitID).
		Save(ctx)
	require.NoError(s.T(), err)
	return acc
}

func (s *InvariantsSuite) createPostedTransaction(dateStr string, postings []model.PostingInput) model.TransactionResponse {
	req := model.CreateTransactionRequest{
		Date:     dateStr,
		Kind:     "adjustment",
		Postings: postings,
	}
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Created()
	res := testutil.DecodeOther[model.TransactionResponse](s)
	return res.Data
}

func (s *InvariantsSuite) voidTransaction(txID uuid.UUID) model.TransactionResponse {
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", txID), fiber.MethodPost, nil)
	s.OK()
	res := testutil.DecodeOther[model.TransactionResponse](s)
	return res.Data
}

// checkAccountingInvariants validates all four accounting invariants for a given parish:
// 1. global sum debit == sum credit over all postings in the parish
// 2. trial balance totals are equal (Totals.Debit == Totals.Credit)
// 3. every voided transaction has exactly 2 entries and nets to zero per account
// 4. every non-void transaction has exactly 1 entry
func (s *InvariantsSuite) checkAccountingInvariants(parishID int, asOf string) error {
	st := testutil.GetState()
	ctx := context.Background()

	// 1. Global sum of debit == credit across all postings in the parish
	postings, err := st.Ent.Posting.Query().
		Where(
			entposting.HasAccountWith(account.HasParishWith(entparish.IDEQ(parishID))),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query postings failed: %w", err)
	}

	var globalDebit, globalCredit int64
	for _, p := range postings {
		switch p.Side {
		case entposting.SideDebit:
			globalDebit += p.Amount
		case entposting.SideCredit:
			globalCredit += p.Amount
		default:
			return fmt.Errorf("unknown posting side: %s", p.Side)
		}
	}
	if globalDebit != globalCredit {
		return fmt.Errorf("global postings unbalanced: sum(debit)=%d != sum(credit)=%d", globalDebit, globalCredit)
	}

	// 2. Trial balance totals are equal
	tbURL := fmt.Sprintf("/fin/reports/trial-balance?asOf=%s", asOf)
	s.Request(tbURL, fiber.MethodGet, nil)
	if s.HttpResponse.StatusCode != fiber.StatusOK {
		return fmt.Errorf("trial balance returned status %d", s.HttpResponse.StatusCode)
	}
	tbRes := testutil.DecodeOther[model.TrialBalanceResponse](s)
	if tbRes.Data.Totals.Debit != tbRes.Data.Totals.Credit {
		return fmt.Errorf("trial balance totals unbalanced: debit=%d != credit=%d", tbRes.Data.Totals.Debit, tbRes.Data.Totals.Credit)
	}

	// 3. Every voided transaction has exactly two entries and nets to zero per account
	voidTxs, err := st.Ent.Transaction.Query().
		Where(
			transaction.HasParishWith(entparish.IDEQ(parishID)),
			transaction.StatusEQ(transaction.StatusVoid),
		).
		WithEntries(func(q *ent.EntryQuery) {
			q.WithPostings(func(pq *ent.PostingQuery) {
				pq.WithAccount()
			})
		}).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query void transactions failed: %w", err)
	}

	for _, tx := range voidTxs {
		if len(tx.Edges.Entries) != 2 {
			return fmt.Errorf("void transaction %s has %d entries, expected 2", tx.ID, len(tx.Edges.Entries))
		}
		accountNets := make(map[uuid.UUID]int64)
		for _, entry := range tx.Edges.Entries {
			for _, p := range entry.Edges.Postings {
				if p.Edges.Account == nil {
					return fmt.Errorf("posting %s missing account edge", p.ID)
				}
				accID := p.Edges.Account.ID
				if p.Side == entposting.SideDebit {
					accountNets[accID] += p.Amount
				} else {
					accountNets[accID] -= p.Amount
				}
			}
		}
		for accID, net := range accountNets {
			if net != 0 {
				return fmt.Errorf("void transaction %s account %s does not net to zero: net=%d", tx.ID, accID, net)
			}
		}
	}

	// 4. Every non-void transaction has exactly one entry
	nonVoidTxs, err := st.Ent.Transaction.Query().
		Where(
			transaction.HasParishWith(entparish.IDEQ(parishID)),
			transaction.StatusNEQ(transaction.StatusVoid),
		).
		WithEntries().
		All(ctx)
	if err != nil {
		return fmt.Errorf("query non-void transactions failed: %w", err)
	}

	for _, tx := range nonVoidTxs {
		if len(tx.Edges.Entries) != 1 {
			return fmt.Errorf("non-void transaction %s has %d entries, expected 1", tx.ID, len(tx.Edges.Entries))
		}
	}

	return nil
}

func (s *InvariantsSuite) TestAccountingInvariants_AllHold() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	// Chart of Accounts
	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "Cash", "asset", true)
	bank := s.createAccountWithUnit(parish.ID, unit.ID, "1002", "Bank", "asset", true)
	ar := s.createAccountWithUnit(parish.ID, unit.ID, "1201", "Accounts Receivable", "asset", true)
	kolekte := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "Kolekte", "income", true)
	donation := s.createAccountWithUnit(parish.ID, unit.ID, "4002", "Donations", "income", true)
	opExpense := s.createAccountWithUnit(parish.ID, unit.ID, "5001", "Operational Expense", "expense", true)
	utilExpense := s.createAccountWithUnit(parish.ID, unit.ID, "5002", "Utilities Expense", "expense", true)

	// 1. Simple transaction (1 debit, 1 credit)
	s.createPostedTransaction("2026-10-01", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 500000},
		{AccountID: kolekte.ID, Side: "credit", Amount: 500000},
	})

	// 2. Split transaction (1 debit, 2 credits)
	s.createPostedTransaction("2026-10-02", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 1500000},
		{AccountID: kolekte.ID, Side: "credit", Amount: 1000000},
		{AccountID: donation.ID, Side: "credit", Amount: 500000},
	})

	// 3. Split transaction (2 debits, 1 credit) - to be voided later
	tx3 := s.createPostedTransaction("2026-10-03", []model.PostingInput{
		{AccountID: opExpense.ID, Side: "debit", Amount: 300000},
		{AccountID: utilExpense.ID, Side: "debit", Amount: 200000},
		{AccountID: bank.ID, Side: "credit", Amount: 500000},
	})

	// 4. Split transaction (multi-debit, multi-credit)
	s.createPostedTransaction("2026-10-04", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 400000},
		{AccountID: ar.ID, Side: "debit", Amount: 600000},
		{AccountID: kolekte.ID, Side: "credit", Amount: 700000},
		{AccountID: donation.ID, Side: "credit", Amount: 300000},
	})

	// 5. Simple transaction - to be voided later
	tx5 := s.createPostedTransaction("2026-10-05", []model.PostingInput{
		{AccountID: opExpense.ID, Side: "debit", Amount: 150000},
		{AccountID: cash.ID, Side: "credit", Amount: 150000},
	})

	// Void tx3 and tx5
	s.voidTransaction(tx3.ID)
	s.voidTransaction(tx5.ID)

	// Invariants check
	err := s.checkAccountingInvariants(parish.ID, "2026-10-31")
	s.Require().NoError(err, "all accounting invariants must hold after transactions and voids")
}

func (s *InvariantsSuite) TestAccountingInvariants_CorruptedPostingFails() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "Cash", "asset", true)
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "Income", "income", true)

	// Create valid transaction
	tx := s.createPostedTransaction("2026-10-01", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 100000},
		{AccountID: income.ID, Side: "credit", Amount: 100000},
	})

	// Verify invariants hold initially
	err := s.checkAccountingInvariants(parish.ID, "2026-10-31")
	s.Require().NoError(err, "invariants must hold before corruption")

	// Get a posting ID from the created transaction
	st := testutil.GetState()
	postings, err := st.Ent.Posting.Query().
		Where(
			entposting.HasEntryWith(
				ententry.HasTransactionWith(transaction.IDEQ(tx.ID)),
			),
		).
		All(context.Background())
	s.Require().NoError(err)
	s.Require().NotEmpty(postings)
	targetPostingID := postings[0].ID

	// Manually corrupt the posting in SQLite (negative control)
	db, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&_busy_timeout=5000")
	s.Require().NoError(err)
	defer db.Close()

	res, err := db.Exec("UPDATE postings SET amount = amount + 9999 WHERE id = ?", targetPostingID)
	s.Require().NoError(err)
	rowsAffected, err := res.RowsAffected()
	s.Require().NoError(err)
	s.Equal(int64(1), rowsAffected, "exactly one posting row should be corrupted")

	// Verify that invariants check detects corruption and returns an error
	err = s.checkAccountingInvariants(parish.ID, "2026-10-31")
	s.Require().Error(err, "accounting invariants must fail when a posting is manually corrupted")
	s.Contains(err.Error(), "unbalanced", "error message should mention unbalanced postings")
}

func TestInvariantsSuite(t *testing.T) {
	suite.Run(t, new(InvariantsSuite))
}
