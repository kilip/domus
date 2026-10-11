package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AccountSuite struct {
	testutil.ApiTestSuite[model.AccountResponse]
}

func (s *AccountSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *AccountSuite) grantRole(parishID int, role string) {
	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parishID)}
	s.GrantRole(role, scope)
}

func (s *AccountSuite) createPosting(parishID int, accountID uuid.UUID) {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	trx, err := st.Ent.Transaction.Create().
		SetNumber(fmt.Sprintf("TRX-%s", uuid.New().String()[:8])).
		SetDate(time.Now()).
		SetKind("income").
		SetStatus("posted").
		SetParishID(parishID).
		Save(ctx)
	require.NoError(s.T(), err)

	entry, err := st.Ent.Entry.Create().
		SetTransactionID(trx.ID).
		SetDate(time.Now()).
		Save(ctx)
	require.NoError(s.T(), err)

	_, err = st.Ent.Posting.Create().
		SetEntryID(entry.ID).
		SetAccountID(accountID).
		SetAmount(1000).
		SetSide("debit").
		Save(ctx)
	require.NoError(s.T(), err)
}

func (s *AccountSuite) TestCreate_SuccessAndDerivedBalance() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	// 1. Asset type -> derived normal balance "debit"
	req := model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unit.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("1001", res.Data.Code)
	s.Equal("Kas", res.Data.Name)
	s.Equal("asset", res.Data.Type)
	s.Equal("debit", res.Data.NormalBalance)
	s.True(res.Data.Postable)
	s.Equal("active", res.Data.Status)
	s.Equal(parish.ID, res.Data.ParishID)
	s.Equal(unit.ID, res.Data.UnitID)
	s.Nil(res.Data.ParentID)

	// 2. Liability type -> derived normal balance "credit"
	reqLiab := model.CreateAccountRequest{
		Code:   "2001",
		Name:   "Hutang",
		Type:   "liability",
		UnitID: unit.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, reqLiab)
	s.Created()

	resLiab := s.GetResponse()
	s.Equal("credit", resLiab.Data.NormalBalance)
	s.Equal(unit.ID, resLiab.Data.UnitID)
}

func (s *AccountSuite) TestCreate_DuplicateCode() {
	parishA := CreateTestParish(s.T())
	unitA := CreateTestUnit(s.T(), parishA.ID, s.User.ID)
	parishB := CreateTestParish(s.T())
	unitB := CreateTestUnit(s.T(), parishB.ID, s.User.ID)

	s.grantRole(parishA.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(parishB.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	req := model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unitA.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishA.ID), fiber.MethodPost, req)
	s.Created()

	// Same parish -> 409 Conflict
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishA.ID), fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusConflict)

	// Different parish -> 201 Created
	reqB := model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unitB.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishB.ID), fiber.MethodPost, reqB)
	s.Created()
}

func (s *AccountSuite) TestCreate_ValidationErrors() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	otherParish := CreateTestParish(s.T())
	otherUnit := CreateTestUnit(s.T(), otherParish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	// Missing code -> 422
	reqMissingCode := map[string]any{
		"name":   "Kas",
		"type":   "asset",
		"unitId": unit.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, reqMissingCode)
	s.AssertStatus(fiber.StatusUnprocessableEntity)

	// Missing unitId -> 422
	reqMissingUnit := map[string]any{
		"code": "1001",
		"name": "Kas",
		"type": "asset",
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, reqMissingUnit)
	s.AssertStatus(fiber.StatusUnprocessableEntity)

	// Invalid type -> 422
	reqInvalidType := model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "invalid_type",
		UnitID: unit.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, reqInvalidType)
	s.AssertStatus(fiber.StatusUnprocessableEntity)

	// Unit from another parish -> 400
	reqOtherUnit := model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: otherUnit.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, reqOtherUnit)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AccountSuite) TestCreate_ParentRules() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	otherParish := CreateTestParish(s.T())
	otherUnit := CreateTestUnit(s.T(), otherParish.ID, s.User.ID)

	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(otherParish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	postableFalse := false
	postableTrue := true

	// 1. Create header parent (postable = false) -> success
	headerReq := model.CreateAccountRequest{
		Code:     "1000",
		Name:     "Aktiva Lancar",
		Type:     "asset",
		UnitID:   unit.ID,
		Postable: &postableFalse,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, headerReq)
	s.Created()
	headerAcc := s.GetResponse().Data

	// 2. Create child under header parent -> success
	childReq := model.CreateAccountRequest{
		Code:     "1001",
		Name:     "Kas",
		Type:     "asset",
		UnitID:   unit.ID,
		ParentID: &headerAcc.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, childReq)
	s.Created()
	s.Equal(&headerAcc.ID, s.GetResponse().Data.ParentID)

	// 3. Postable parent -> 400
	postableParentReq := model.CreateAccountRequest{
		Code:     "1100",
		Name:     "Bank Postable",
		Type:     "asset",
		UnitID:   unit.ID,
		Postable: &postableTrue,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, postableParentReq)
	s.Created()
	postableAcc := s.GetResponse().Data

	childWithPostableParent := model.CreateAccountRequest{
		Code:     "1101",
		Name:     "Sub Bank",
		Type:     "asset",
		UnitID:   unit.ID,
		ParentID: &postableAcc.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, childWithPostableParent)
	s.AssertStatus(fiber.StatusBadRequest)

	// 4. Type mismatch: child expense under asset parent -> 400
	mismatchChild := model.CreateAccountRequest{
		Code:     "5001",
		Name:     "Beban Kas",
		Type:     "expense",
		UnitID:   unit.ID,
		ParentID: &headerAcc.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, mismatchChild)
	s.AssertStatus(fiber.StatusBadRequest)

	// 5. Cross-parish parent: header in parish, child in otherParish -> 400
	crossParishChild := model.CreateAccountRequest{
		Code:     "1001",
		Name:     "Kas Other",
		Type:     "asset",
		UnitID:   otherUnit.ID,
		ParentID: &headerAcc.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", otherParish.ID), fiber.MethodPost, crossParishChild)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AccountSuite) TestUpdate_RulesAndCycles() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	postableFalse := false

	// Create account A
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:     "1000",
		Name:     "A",
		Type:     "asset",
		UnitID:   unit.ID,
		Postable: &postableFalse,
	})
	s.Created()
	accA := s.GetResponse().Data

	// Create account B under A
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:     "1100",
		Name:     "B",
		Type:     "asset",
		UnitID:   unit.ID,
		ParentID: &accA.ID,
		Postable: &postableFalse,
	})
	s.Created()
	accB := s.GetResponse().Data

	// Create account C under B
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:     "1110",
		Name:     "C",
		Type:     "asset",
		UnitID:   unit.ID,
		ParentID: &accB.ID,
		Postable: &postableFalse,
	})
	s.Created()
	accC := s.GetResponse().Data

	// 1. Self-parent: update A with parent A -> 400
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, accA.ID), fiber.MethodPut, model.UpdateAccountRequest{
		Code:     accA.Code,
		Name:     accA.Name,
		Type:     accA.Type,
		ParentID: &accA.ID,
	})
	s.AssertStatus(fiber.StatusBadRequest)

	// 2. Cycle: update A with parent C (which is descendant of A) -> 400
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, accA.ID), fiber.MethodPut, model.UpdateAccountRequest{
		Code:     accA.Code,
		Name:     accA.Name,
		Type:     accA.Type,
		ParentID: &accC.ID,
	})
	s.AssertStatus(fiber.StatusBadRequest)

	// 3. Update postable=true on account A that has children -> 400
	postableTrue := true
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, accA.ID), fiber.MethodPut, model.UpdateAccountRequest{
		Code:     accA.Code,
		Name:     accA.Name,
		Type:     accA.Type,
		Postable: &postableTrue,
	})
	s.AssertStatus(fiber.StatusBadRequest)

	// 4. Normal update success (without unitId/parishId)
	desc := "Updated description"
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, accC.ID), fiber.MethodPut, model.UpdateAccountRequest{
		Code:        "1115",
		Name:        "C Updated",
		Type:        "asset",
		Description: &desc,
	})
	s.OK()
	res := s.GetResponse().Data
	s.Equal("1115", res.Code)
	s.Equal("C Updated", res.Name)
	s.Equal(&desc, res.Description)
	s.Equal(unit.ID, res.UnitID)     // Unit unchanged!
	s.Equal(parish.ID, res.ParishID) // Parish unchanged!
}

func (s *AccountSuite) TestUpdate_PostingsRestrictions() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	// Create postable account
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unit.ID,
	})
	s.Created()
	acc := s.GetResponse().Data

	// Add posting to account
	s.createPosting(parish.ID, acc.ID)

	// 1. Changing type with postings -> 400
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, acc.ID), fiber.MethodPut, model.UpdateAccountRequest{
		Code: acc.Code,
		Name: acc.Name,
		Type: "expense",
	})
	s.AssertStatus(fiber.StatusBadRequest)

	// 2. Changing postable with postings -> 400
	postableFalse := false
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, acc.ID), fiber.MethodPut, model.UpdateAccountRequest{
		Code:     acc.Code,
		Name:     acc.Name,
		Type:     acc.Type,
		Postable: &postableFalse,
	})
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AccountSuite) TestList_WithFilters() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	postableFalse := false
	statusInactive := "inactive"

	// Create 1: asset, active, header
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:     "1000",
		Name:     "Aktiva",
		Type:     "asset",
		UnitID:   unit.ID,
		Postable: &postableFalse,
	})
	s.Created()

	// Create 2: asset, active, postable
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unit.ID,
	})
	s.Created()

	// Create 3: expense, inactive, postable
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "5001",
		Name:   "Beban Listrik",
		Type:   "expense",
		UnitID: unit.ID,
		Status: &statusInactive,
	})
	s.Created()

	// 1. List all sorted by code
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodGet, nil)
	s.OK()
	listAll := testutil.DecodeOther[[]model.AccountResponse](s).Data
	s.Len(listAll, 3)
	s.Equal("1000", listAll[0].Code)
	s.Equal("1001", listAll[1].Code)
	s.Equal("5001", listAll[2].Code)

	// 2. Filter type=asset
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts?type=asset", parish.ID), fiber.MethodGet, nil)
	s.OK()
	listAsset := testutil.DecodeOther[[]model.AccountResponse](s).Data
	s.Len(listAsset, 2)

	// 3. Filter status=inactive
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts?status=inactive", parish.ID), fiber.MethodGet, nil)
	s.OK()
	listInactive := testutil.DecodeOther[[]model.AccountResponse](s).Data
	s.Len(listInactive, 1)
	s.Equal("5001", listInactive[0].Code)

	// 4. Filter postable=false
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts?postable=false", parish.ID), fiber.MethodGet, nil)
	s.OK()
	listHeaders := testutil.DecodeOther[[]model.AccountResponse](s).Data
	s.Len(listHeaders, 1)
	s.Equal("1000", listHeaders[0].Code)
}

func (s *AccountSuite) TestGetByID_And_CrossParish() {
	parishA := CreateTestParish(s.T())
	unitA := CreateTestUnit(s.T(), parishA.ID, s.User.ID)
	parishB := CreateTestParish(s.T())
	unitB := CreateTestUnit(s.T(), parishB.ID, s.User.ID)
	_ = unitB

	s.grantRole(parishA.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(parishB.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishA.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unitA.ID,
	})
	s.Created()
	acc := s.GetResponse().Data

	// Get by ID success
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parishA.ID, acc.ID), fiber.MethodGet, nil)
	s.OK()
	getRes := s.GetResponse().Data
	s.Equal(acc.ID, getRes.ID)
	s.Equal(unitA.ID, getRes.UnitID)
	s.Equal(parishA.ID, getRes.ParishID)

	// Get from another parish -> 404 Not Found
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parishB.ID, acc.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Invalid ID parameter -> 400 Bad Request
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/not-a-uuid", parishA.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AccountSuite) TestDelete_RulesAndSoftDelete() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	postableFalse := false

	// Create parent A
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:     "1000",
		Name:     "Parent",
		Type:     "asset",
		UnitID:   unit.ID,
		Postable: &postableFalse,
	})
	s.Created()
	parentAcc := s.GetResponse().Data

	// Create child B
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:     "1001",
		Name:     "Child",
		Type:     "asset",
		UnitID:   unit.ID,
		ParentID: &parentAcc.ID,
	})
	s.Created()
	childAcc := s.GetResponse().Data

	// 1. Delete parent with children -> 400
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, parentAcc.ID), fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusBadRequest)

	// 2. Add posting to child B, then delete child -> 400
	s.createPosting(parish.ID, childAcc.ID)
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, childAcc.ID), fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusBadRequest)

	// 3. Create independent leaf account C and delete -> 204
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parish.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "9999",
		Name:   "Temporary",
		Type:   "expense",
		UnitID: unit.ID,
	})
	s.Created()
	leafAcc := s.GetResponse().Data

	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, leafAcc.ID), fiber.MethodDelete, nil)
	s.NoContent()

	// Normal query should not find deleted account -> 404
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts/%s", parish.ID, leafAcc.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Ent DB check with SkipSoftDelete: row still exists!
	st := testutil.GetState()
	skipCtx := core.SkipSoftDelete(context.Background())
	found, err := st.Ent.Account.Get(skipCtx, leafAcc.ID)
	require.NoError(s.T(), err)
	s.NotNil(found.DeletedAt, "deletedAt timestamp should be set")
}

func (s *AccountSuite) TestAuthz_Matrix() {
	parishA := CreateTestParish(s.T())
	unitA := CreateTestUnit(s.T(), parishA.ID, s.User.ID)
	parishB := CreateTestParish(s.T())
	unitB := CreateTestUnit(s.T(), parishB.ID, s.User.ID)

	// User has NO membership
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishA.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)

	// User has finance-reader in Parish A
	s.grantRole(parishA.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelReader))

	// Reader can GET
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishA.ID), fiber.MethodGet, nil)
	s.OK()

	// Reader cannot POST -> 403
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishA.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unitA.ID,
	})
	s.AssertStatus(fiber.StatusForbidden)

	// Writer in parish A on parish B -> 403
	s.grantRole(parishA.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.Request(fmt.Sprintf("/parishes/%d/finance/accounts", parishB.ID), fiber.MethodPost, model.CreateAccountRequest{
		Code:   "1001",
		Name:   "Kas",
		Type:   "asset",
		UnitID: unitB.ID,
	})
	s.AssertStatus(fiber.StatusForbidden)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(AccountSuite))
}
