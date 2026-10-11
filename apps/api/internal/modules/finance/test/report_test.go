package test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ReportSuite struct {
	testutil.ApiTestSuite[any]
}

func (s *ReportSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *ReportSuite) grantRole(parishID int, role string) {
	GrantFinanceRole(&s.ApiTestSuite, parishID, role)
}

func (s *ReportSuite) activate(parishID, unitID int) {
	SetFinanceCookie(&s.ApiTestSuite, parishID, unitID)
}

func (s *ReportSuite) createAccount(parishID int, code string, accType string, postable bool, parentID *uuid.UUID, unitID int) *ent.Account {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	b := st.Ent.Account.Create().
		SetCode(code).
		SetName("Account " + code).
		SetType(account.Type(accType)).
		SetNormalBalance(account.NormalBalance(model.DeriveNormalBalance(accType))).
		SetPostable(postable).
		SetStatus(account.StatusActive).
		SetParishID(parishID).
		SetUnitID(unitID)
	if parentID != nil {
		b.SetParentID(*parentID)
	}
	acc, err := b.Save(ctx)
	require.NoError(s.T(), err)
	return acc
}

func (s *ReportSuite) createPostedTransaction(dateStr string, postings []model.PostingInput) model.TransactionResponse {
	req := model.CreateTransactionRequest{
		Date:     dateStr,
		Kind:     "income",
		Postings: postings,
	}
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Created()
	return testutil.DecodeOther[model.TransactionResponse](s).Data
}

func (s *ReportSuite) voidTransaction(txID uuid.UUID) {
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", txID), fiber.MethodPost, nil)
	s.OK()
}

func (s *ReportSuite) TestLedger_RunningBalanceAndPagination() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccount(parish.ID, "1001", "asset", true, nil, unit.ID)
	income := s.createAccount(parish.ID, "4001", "income", true, nil, unit.ID)

	// Pre-period transaction (2026-01-10): Cash +100_000
	s.createPostedTransaction("2026-01-10", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 100000},
		{AccountID: income.ID, Side: "credit", Amount: 100000},
	})

	// In-period transactions (2026-02-01 to 2026-02-15)
	// Tx 1: 2026-02-01 Cash +50_000
	s.createPostedTransaction("2026-02-01", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 50000},
		{AccountID: income.ID, Side: "credit", Amount: 50000},
	})
	// Tx 2: 2026-02-05 Cash -20_000
	s.createPostedTransaction("2026-02-05", []model.PostingInput{
		{AccountID: income.ID, Side: "debit", Amount: 20000},
		{AccountID: cash.ID, Side: "credit", Amount: 20000},
	})
	// Tx 3: 2026-02-10 Cash +30_000
	s.createPostedTransaction("2026-02-10", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 30000},
		{AccountID: income.ID, Side: "credit", Amount: 30000},
	})

	// 1. Query Cash (Debit-Normal) with limit=2 (Page 1)
	ledgerURL := fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-02-01&to=2026-02-28&limit=2", cash.ID)
	s.Request(ledgerURL, fiber.MethodGet, nil)
	s.OK()

	p1 := testutil.DecodeOther[model.LedgerResponse](s)
	s.Equal(int64(100000), p1.Data.OpeningBalance, "opening balance should equal pre-period transaction sum")
	s.Len(p1.Data.Rows, 2)
	s.Equal(int64(150000), p1.Data.Rows[0].Balance) // 100k + 50k
	s.Equal(int64(130000), p1.Data.Rows[1].Balance) // 150k - 20k
	s.Equal(int64(130000), p1.Data.ClosingBalance)
	s.NotEmpty(p1.Meta.Cursor)

	// Page 2 with cursor
	p2URL := fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-02-01&to=2026-02-28&limit=2&cursor=%s", cash.ID, url.QueryEscape(p1.Meta.Cursor))
	s.Request(p2URL, fiber.MethodGet, nil)
	s.OK()

	p2 := testutil.DecodeOther[model.LedgerResponse](s)
	s.Equal(int64(130000), p2.Data.OpeningBalance, "page 2 opening balance must match page 1 closing balance")
	s.Len(p2.Data.Rows, 1)
	s.Equal(int64(160000), p2.Data.Rows[0].Balance) // 130k + 30k
	s.Equal(int64(160000), p2.Data.ClosingBalance)
	s.Empty(p2.Meta.Cursor)

	// 2. Query Income (Credit-Normal) in the same period
	incomeURL := fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-02-01&to=2026-02-28", income.ID)
	s.Request(incomeURL, fiber.MethodGet, nil)
	s.OK()

	incResp := testutil.DecodeOther[model.LedgerResponse](s)
	s.Equal(int64(100000), incResp.Data.OpeningBalance)
	s.Len(incResp.Data.Rows, 3)
	// Tx 1: credit 50k -> balance = 100k + 50k = 150k
	s.Equal(int64(150000), incResp.Data.Rows[0].Balance)
	// Tx 2: debit 20k -> balance = 150k - 20k = 130k
	s.Equal(int64(130000), incResp.Data.Rows[1].Balance)
	// Tx 3: credit 30k -> balance = 130k + 30k = 160k
	s.Equal(int64(160000), incResp.Data.Rows[2].Balance)
	s.Equal(int64(160000), incResp.Data.ClosingBalance)
}

func (s *ReportSuite) TestLedger_VoidInclusionAndNetting() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccount(parish.ID, "1001", "asset", true, nil, unit.ID)
	income := s.createAccount(parish.ID, "4001", "income", true, nil, unit.ID)

	// Initial balance: 50k
	s.createPostedTransaction("2026-03-01", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 50000},
		{AccountID: income.ID, Side: "credit", Amount: 50000},
	})

	// Transaction to be voided: 25k
	tx2 := s.createPostedTransaction("2026-03-05", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 25000},
		{AccountID: income.ID, Side: "credit", Amount: 25000},
	})

	// Void tx2
	s.voidTransaction(tx2.ID)

	ledgerURL := fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-03-01&to=2026-03-31", cash.ID)
	s.Request(ledgerURL, fiber.MethodGet, nil)
	s.OK()

	res := testutil.DecodeOther[model.LedgerResponse](s)
	s.Equal(int64(0), res.Data.OpeningBalance)
	s.Len(res.Data.Rows, 3) // Initial, Void original, Void reversal
	s.Equal(int64(50000), res.Data.Rows[0].Balance)
	s.Equal(int64(75000), res.Data.Rows[1].Balance)
	s.Equal(int64(50000), res.Data.Rows[2].Balance, "balance must return to pre-transaction balance after reversal")
	s.Equal(int64(50000), res.Data.ClosingBalance)
}

func (s *ReportSuite) TestTrialBalance_DebitsEqualCreditsAndAsOf() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccount(parish.ID, "1001", "asset", true, nil, unit.ID)
	bank := s.createAccount(parish.ID, "1002", "asset", true, nil, unit.ID)
	income := s.createAccount(parish.ID, "4001", "income", true, nil, unit.ID)
	expense := s.createAccount(parish.ID, "5001", "expense", true, nil, unit.ID)

	// 1. Tx on 2026-04-01: Cash debit 100k, Income credit 100k
	s.createPostedTransaction("2026-04-01", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 100000},
		{AccountID: income.ID, Side: "credit", Amount: 100000},
	})

	// 2. Tx on 2026-04-05: Expense debit 30k, Bank credit 30k
	tx2 := s.createPostedTransaction("2026-04-05", []model.PostingInput{
		{AccountID: expense.ID, Side: "debit", Amount: 30000},
		{AccountID: bank.ID, Side: "credit", Amount: 30000},
	})

	// 3. Void tx2 on 2026-04-05
	s.voidTransaction(tx2.ID)

	// 4. Tx on 2026-04-20 (after asOf 2026-04-10): Cash debit 50k, Income credit 50k
	s.createPostedTransaction("2026-04-20", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 50000},
		{AccountID: income.ID, Side: "credit", Amount: 50000},
	})

	// Query Trial Balance asOf 2026-04-10
	tbURL := "/fin/reports/trial-balance?asOf=2026-04-10"
	s.Request(tbURL, fiber.MethodGet, nil)
	s.OK()

	res := testutil.DecodeOther[model.TrialBalanceResponse](s)
	s.Equal("2026-04-10", res.Data.AsOf)
	s.Equal(res.Data.Totals.Debit, res.Data.Totals.Credit, "totals.debit must equal totals.credit")

	// Verify order by code ASC
	for i := 1; i < len(res.Data.Rows); i++ {
		s.True(res.Data.Rows[i-1].Code < res.Data.Rows[i].Code, "rows must be ordered by code ASC")
	}

	// Verify later transaction (2026-04-20) is excluded: Cash debit should be 100k, not 150k
	for _, row := range res.Data.Rows {
		if row.AccountID == cash.ID {
			s.Equal(int64(100000), row.Debit)
		}
	}
}

func (s *ReportSuite) TestCollections_SubtreeUnitAndValidation() {
	parish := CreateTestParish(s.T())
	unit1 := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	unit2 := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	// Hierarchy for unit1:
	// 4000 Header Income
	//   ├── 4001 Kolekte 1 (postable)
	//   └── 4002 Kolekte 2 (postable)
	// 4900 Other Income (postable, outside 4000)
	cash := s.createAccount(parish.ID, "1001", "asset", true, nil, unit1.ID)
	headerInc := s.createAccount(parish.ID, "4000", "income", false, nil, unit1.ID)
	child1 := s.createAccount(parish.ID, "4001", "income", true, &headerInc.ID, unit1.ID)
	child2 := s.createAccount(parish.ID, "4002", "income", true, &headerInc.ID, unit1.ID)
	otherInc := s.createAccount(parish.ID, "4900", "income", true, nil, unit1.ID)

	s.activate(parish.ID, unit1.ID)

	// 1. Kolekte 1 with unit1: 60k
	s.createPostedTransaction("2026-05-01", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 60000},
		{AccountID: child1.ID, Side: "credit", Amount: 60000},
	})

	// 2. Kolekte 2 with unit1: 40k
	s.createPostedTransaction("2026-05-02", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 40000},
		{AccountID: child2.ID, Side: "credit", Amount: 40000},
	})

	// 3. Other Income with unit1: 50k
	s.createPostedTransaction("2026-05-03", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 50000},
		{AccountID: otherInc.ID, Side: "credit", Amount: 50000},
	})

	// 4. Voided Kolekte 1 transaction: 15k
	voidTx := s.createPostedTransaction("2026-05-04", []model.PostingInput{
		{AccountID: cash.ID, Side: "debit", Amount: 15000},
		{AccountID: child1.ID, Side: "credit", Amount: 15000},
	})
	s.voidTransaction(voidTx.ID)

	// A. Query without parent filter: all income accounts (child1: 60k, child2: 40k, other: 50k) -> grandTotal = 150k
	urlA := "/fin/reports/collections?from=2026-05-01&to=2026-05-31"
	s.Request(urlA, fiber.MethodGet, nil)
	s.OK()
	resA := testutil.DecodeOther[model.CollectionsResponse](s)
	s.Equal(int64(150000), resA.Data.GrandTotal)
	s.Len(resA.Data.Rows, 3)

	// B. Query filtered by parentAccountId = headerInc.ID -> only child1 & child2 -> 100k
	urlB := fmt.Sprintf("/fin/reports/collections?from=2026-05-01&to=2026-05-31&parentAccountId=%s", headerInc.ID)
	s.Request(urlB, fiber.MethodGet, nil)
	s.OK()
	resB := testutil.DecodeOther[model.CollectionsResponse](s)
	s.Equal(int64(100000), resB.Data.GrandTotal)
	s.Len(resB.Data.Rows, 2)

	// C. When activating unit2 (which has no accounts/transactions): empty collections
	s.activate(parish.ID, unit2.ID)
	s.Request(urlA, fiber.MethodGet, nil)
	s.OK()
	resC := testutil.DecodeOther[model.CollectionsResponse](s)
	s.Equal(int64(0), resC.Data.GrandTotal)
	s.Empty(resC.Data.Rows)

	// Switch back to unit1
	s.activate(parish.ID, unit1.ID)

	// D. Validation: date range > 366 days -> 400
	urlD := "/fin/reports/collections?from=2024-01-01&to=2026-01-01"
	s.Request(urlD, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)

	// E. Validation: parentAccountId is postable -> 400
	urlE := fmt.Sprintf("/fin/reports/collections?from=2026-05-01&to=2026-05-31&parentAccountId=%s", child1.ID)
	s.Request(urlE, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)

	// F. Validation: parentAccountId not found -> 404
	urlF := fmt.Sprintf("/fin/reports/collections?from=2026-05-01&to=2026-05-31&parentAccountId=%s", uuid.New())
	s.Request(urlF, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ReportSuite) TestTenantIsolation_CrossParish() {
	parishA := CreateTestParish(s.T())
	unitA := CreateTestUnit(s.T(), parishA.ID, s.User.ID)
	parishB := CreateTestParish(s.T())
	unitB := CreateTestUnit(s.T(), parishB.ID, s.User.ID)

	s.grantRole(parishA.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(parishB.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	s.activate(parishB.ID, unitB.ID)
	cashB := s.createAccount(parishB.ID, "1001", "asset", true, nil, unitB.ID)
	incB := s.createAccount(parishB.ID, "4001", "income", true, nil, unitB.ID)

	// Create transaction in Parish B
	s.createPostedTransaction("2026-06-01", []model.PostingInput{
		{AccountID: cashB.ID, Side: "debit", Amount: 200000},
		{AccountID: incB.ID, Side: "credit", Amount: 200000},
	})

	// Switch to Parish A
	s.activate(parishA.ID, unitA.ID)

	// Parish A trial balance should be empty
	s.Request("/fin/reports/trial-balance?asOf=2026-06-30", fiber.MethodGet, nil)
	s.OK()
	tbA := testutil.DecodeOther[model.TrialBalanceResponse](s)
	s.Empty(tbA.Data.Rows)
	s.Equal(int64(0), tbA.Data.Totals.Debit)

	// Parish A collections should be empty
	s.Request("/fin/reports/collections?from=2026-06-01&to=2026-06-30", fiber.MethodGet, nil)
	s.OK()
	collA := testutil.DecodeOther[model.CollectionsResponse](s)
	s.Empty(collA.Data.Rows)
	s.Equal(int64(0), collA.Data.GrandTotal)

	// Querying Parish B account in Parish A ledger -> 404 Not Found
	s.Request(fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-06-01&to=2026-06-30", cashB.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ReportSuite) TestAuthz_ReaderAllowedAndNoMembershipForbidden() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	cash := s.createAccount(parish.ID, "1001", "asset", true, nil, unit.ID)

	s.activate(parish.ID, unit.ID)

	// User without membership -> 403 Forbidden
	s.Request(fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-01-01&to=2026-01-31", cash.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)

	s.Request("/fin/reports/trial-balance?asOf=2026-01-31", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)

	s.Request("/fin/reports/collections?from=2026-01-01&to=2026-01-31", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)

	// Grant Reader role -> 200 OK
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelReader))

	s.Request(fmt.Sprintf("/fin/reports/ledger?accountId=%s&from=2026-01-01&to=2026-01-31", cash.ID), fiber.MethodGet, nil)
	s.OK()

	s.Request("/fin/reports/trial-balance?asOf=2026-01-31", fiber.MethodGet, nil)
	s.OK()

	s.Request("/fin/reports/collections?from=2026-01-01&to=2026-01-31", fiber.MethodGet, nil)
	s.OK()
}

func TestReportSuite(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}
