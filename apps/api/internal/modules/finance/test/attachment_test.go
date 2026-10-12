package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/attachment"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/storage"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AttachmentSuite struct {
	testutil.ApiTestSuite[attachment.Response]
}

func (s *AttachmentSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUser(&s.ApiTestSuite)
}

func (s *AttachmentSuite) grantRole(parishID int, role string) {
	GrantFinanceRole(&s.ApiTestSuite, parishID, role)
}

func (s *AttachmentSuite) grantUnitRole(unitID int, role string) {
	s.GrantRole(role, authz.Scope{Type: authz.ScopeWorkspace, ID: int64(unitID)})
}

func (s *AttachmentSuite) activate(parishID, unitID int) {
	SetFinanceCookie(&s.ApiTestSuite, parishID, unitID)
}

// setupTx creates a parish, a workspace and a posted transaction in that workspace.
// The user can write finance at parish scope (to create the transaction)
// and at workspace scope (attachments are authorized per workspace).
func (s *AttachmentSuite) setupTx() (*ent.Parish, *ent.Workspace, model.TransactionResponse) {
	parish := CreateTestParish(s.T())
	writer := authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter)
	s.grantRole(parish.ID, writer)
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantUnitRole(unit.ID, writer)
	s.activate(parish.ID, unit.ID)

	cash := s.createAccount(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, unit.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 50000)
	return parish, unit, tx
}

func (s *AttachmentSuite) createAccount(parishID int, unitID int, code string, accType string, postable bool, status string) *ent.Account {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	acc, err := st.Ent.Account.Create().
		SetCode(code).
		SetName("Account " + code).
		SetType(account.Type(accType)).
		SetNormalBalance(account.NormalBalance(model.DeriveNormalBalance(accType))).
		SetPostable(postable).
		SetStatus(account.Status(status)).
		SetParishID(parishID).
		SetWorkspaceID(unitID).
		Save(ctx)
	require.NoError(s.T(), err)
	return acc
}

func (s *AttachmentSuite) createPostedTransaction(cashID, incomeID uuid.UUID, dateStr string, amount int64) model.TransactionResponse {
	req := model.CreateTransactionRequest{
		Date: dateStr,
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cashID, Side: "debit", Amount: amount},
			{AccountID: incomeID, Side: "credit", Amount: amount},
		},
	}
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Created()
	return testutil.DecodeOther[model.TransactionResponse](&s.ApiTestSuite).Data
}

var (
	validPDFBytes = []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF\n")
	validPNGBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4")
)

func (s *AttachmentSuite) TestUpload_PDF_Success() {
	_, unit, tx := s.setupTx()
	_ = unit

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt.pdf", validPDFBytes)
	s.Created()

	resp := s.GetResponse().Data
	s.Equal(tx.ID, resp.OwnerID)
	s.Equal("transaction", resp.OwnerType)
	s.Equal(unit.ID, resp.WorkspaceID)
	s.Equal("receipt.pdf", resp.Filename)
	s.Equal("application/pdf", resp.MimeType)
	s.Equal(int64(len(validPDFBytes)), resp.Size)
	s.NotEmpty(resp.CreatedAt)
	s.Equal(s.User.ID, resp.CreatedBy.ID)
}

func (s *AttachmentSuite) TestUpload_PNG_Success() {
	_, unit, tx := s.setupTx()
	_ = unit

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt.png", validPNGBytes)
	s.Created()

	resp := s.GetResponse().Data
	s.Equal("receipt.png", resp.Filename)
	s.Equal("image/png", resp.MimeType)
}

func (s *AttachmentSuite) TestUpload_InvalidMimeType_Returns400() {
	_, _, tx := s.setupTx()

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	textBytes := []byte("plain text content which is not allowed")
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt.txt", textBytes)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AttachmentSuite) TestUpload_FileTooLarge_Returns413() {
	_, _, tx := s.setupTx()

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	largeBytes := make([]byte, 11*1024*1024) // 11 MB > 10 MB limit
	copy(largeBytes, validPDFBytes)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "large.pdf", largeBytes)
	s.AssertStatus(fiber.StatusRequestEntityTooLarge)
}

func (s *AttachmentSuite) TestUpload_ZeroBytes_Returns400() {
	_, _, tx := s.setupTx()

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "empty.pdf", []byte{})
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AttachmentSuite) TestUpload_MissingFileField_Returns400() {
	_, _, tx := s.setupTx()

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "wrong_field", "file.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AttachmentSuite) TestList_Attachments_SortedChronologically() {
	_, _, tx := s.setupTx()

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "first.pdf", validPDFBytes)
	s.Created()
	att1 := s.GetResponse().Data

	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "second.png", validPNGBytes)
	s.Created()
	att2 := s.GetResponse().Data

	s.Request(uploadURL, fiber.MethodGet, nil)
	s.OK()
	listResp := testutil.DecodeOther[[]attachment.Response](&s.ApiTestSuite).Data
	s.Len(listResp, 2)
	s.Equal(att1.ID, listResp[0].ID)
	s.Equal(att2.ID, listResp[1].ID)
}

func (s *AttachmentSuite) TestDownload_ReturnsPresignedURL() {
	_, _, tx := s.setupTx()

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt.pdf", validPDFBytes)
	s.Created()
	att1 := s.GetResponse().Data

	downloadURL := fmt.Sprintf("/fin/transactions/%s/attachments/%s/download", tx.ID, att1.ID)
	s.Request(downloadURL, fiber.MethodGet, nil)
	s.OK()
	dlResp := testutil.DecodeOther[attachment.DownloadResponse](&s.ApiTestSuite).Data
	s.NotEmpty(dlResp.URL)
	s.True(dlResp.ExpiresAt.After(time.Now()))
}

func (s *AttachmentSuite) TestIsolation_CrossParishReturns404() {
	parishA, _, txA := s.setupTx()
	parishB := CreateTestParish(s.T())
	unitB := CreateTestUnit(s.T(), parishB.ID, s.User.ID)
	s.grantRole(parishB.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	uploadURL_A := fmt.Sprintf("/fin/transactions/%s/attachments", txA.ID)
	s.RequestMultipart(uploadURL_A, fiber.MethodPost, "file", "file.pdf", validPDFBytes)
	s.Created()
	attA := s.GetResponse().Data

	// Switch to parish B
	s.activate(parishB.ID, unitB.ID)

	// Try upload on parish B with txA ID -> 404
	s.RequestMultipart(uploadURL_A, fiber.MethodPost, "file", "file2.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusNotFound)

	// Try list on parish B with txA ID -> 404
	s.Request(uploadURL_A, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Try download on parish B with txA ID and attA ID -> 404
	dlURL_B := fmt.Sprintf("/fin/transactions/%s/attachments/%s/download", txA.ID, attA.ID)
	s.Request(dlURL_B, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Switch back to parish A: try download with mismatched transaction ID -> 404
	s.activate(parishA.ID, *txA.WorkspaceID)
	randomTxID := core.GenerateID()
	dlURL_MismatchedTx := fmt.Sprintf("/fin/transactions/%s/attachments/%s/download", randomTxID, attA.ID)
	s.Request(dlURL_MismatchedTx, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *AttachmentSuite) TestStorageRollback_OnInvalidTransaction() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	fakeTxID := core.GenerateID()
	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", fakeTxID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "orphaned.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusNotFound)

	// Verify no objects with fakeTxID prefix exist in storage
	memStorage := testutil.GetState().Storage.(*storage.MemoryStorage)
	prefix := fmt.Sprintf("transaction/%d/%s/", parish.ID, fakeTxID.String())
	_, _, err := memStorage.Get(prefix + "orphaned.pdf")
	s.ErrorIs(err, storage.ErrNotFound)
}

func (s *AttachmentSuite) TestUpload_VoidTransactionAllowed() {
	_, _, tx := s.setupTx()

	// Void transaction
	voidURL := fmt.Sprintf("/fin/transactions/%s/void", tx.ID)
	s.Request(voidURL, fiber.MethodPost, model.VoidTransactionRequest{})
	s.OK()

	// Upload attachment to voided transaction
	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "void_receipt.pdf", validPDFBytes)
	s.Created()
	resp := s.GetResponse().Data
	s.Equal("void_receipt.pdf", resp.Filename)
}

func (s *AttachmentSuite) TestAuthz_ReaderVsWriter() {
	_, unit, tx := s.setupTx()
	_ = unit

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "doc.pdf", validPDFBytes)
	s.Created()
	att := s.GetResponse().Data

	// Switch to a new user with reader-only role
	readerUser := core.AuthenticatedUser{
		ID:    core.GenerateID(),
		Email: "reader@example.com",
		Name:  "Reader User",
		Roles: []core.UserRole{core.UserRoleAdmin},
	}
	s.User = &readerUser
	EnsureUser(&s.ApiTestSuite)
	s.grantUnitRole(unit.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelReader))

	// Reader can list
	s.Request(uploadURL, fiber.MethodGet, nil)
	s.OK()

	// Reader can download
	dlURL := fmt.Sprintf("/fin/transactions/%s/attachments/%s/download", tx.ID, att.ID)
	s.Request(dlURL, fiber.MethodGet, nil)
	s.OK()

	// Reader cannot upload (403)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "doc2.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusForbidden)
}

func (s *AttachmentSuite) TestAuthz_ParishRoleOnlyIsForbidden() {
	parish, unit, tx := s.setupTx()

	other := core.AuthenticatedUser{ID: core.GenerateID(), Email: "parish-only@example.com", Name: "Parish Only", Roles: []core.UserRole{core.UserRoleAdmin}}
	s.User = &other
	s.GrantRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter), authz.Scope{Type: authz.ScopeParish, ID: int64(parish.ID)})
	s.activate(parish.ID, unit.ID)

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.Request(uploadURL, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "doc.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusForbidden)
}

func (s *AttachmentSuite) TestAuthz_OtherUnitRoleIsForbidden() {
	parish, unit, tx := s.setupTx()

	other := core.AuthenticatedUser{ID: core.GenerateID(), Email: "other-unit@example.com", Name: "Other Unit", Roles: []core.UserRole{core.UserRoleAdmin}}
	s.User = &other
	EnsureUser(&s.ApiTestSuite)
	otherUnit := CreateTestUnit(s.T(), parish.ID, other.ID)
	s.grantUnitRole(otherUnit.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	uploadURL := fmt.Sprintf("/fin/transactions/%s/attachments", tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "doc.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusForbidden)
}

func TestAttachmentSuite(t *testing.T) {
	suite.Run(t, new(AttachmentSuite))
}
