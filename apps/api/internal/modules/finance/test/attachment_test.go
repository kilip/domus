package test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/storage"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AttachmentSuite struct {
	testutil.ApiTestSuite[model.AttachmentResponse]
}

func (s *AttachmentSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *AttachmentSuite) grantRole(parishID int, role string) {
	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parishID)}
	s.GrantRole(role, scope)
}

func (s *AttachmentSuite) createAccount(parishID int, code string, accType string, postable bool, status string) *ent.Account {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	unit := CreateTestUnit(s.T(), parishID, s.User.ID)
	acc, err := st.Ent.Account.Create().
		SetCode(code).
		SetName("Account " + code).
		SetType(account.Type(accType)).
		SetNormalBalance(account.NormalBalance(model.DeriveNormalBalance(accType))).
		SetPostable(postable).
		SetStatus(account.Status(status)).
		SetParishID(parishID).
		SetUnitID(unit.ID).
		Save(ctx)
	require.NoError(s.T(), err)
	return acc
}

func (s *AttachmentSuite) createPostedTransaction(parishID int, cashID, incomeID uuid.UUID, dateStr string, amount int64) model.TransactionResponse {
	req := model.CreateTransactionRequest{
		Date: dateStr,
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cashID, Side: "debit", Amount: amount},
			{AccountID: incomeID, Side: "credit", Amount: amount},
		},
	}
	s.Request(fmt.Sprintf("/parishes/%d/finance/transactions", parishID), fiber.MethodPost, req)
	s.Created()
	return testutil.DecodeOther[model.TransactionResponse](&s.ApiTestSuite).Data
}

var (
	validPDFBytes = []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF\n")
	validPNGBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4")
)

func (s *AttachmentSuite) TestUpload_PDF_Success() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt.pdf", validPDFBytes)
	s.Created()

	resp := s.GetResponse().Data
	s.Equal(tx.ID, resp.TransactionID)
	s.Equal("receipt.pdf", resp.Filename)
	s.Equal("application/pdf", resp.MimeType)
	s.Equal(int64(len(validPDFBytes)), resp.Size)

	hasher := sha256.New()
	hasher.Write(validPDFBytes)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))
	s.Equal(expectedChecksum, resp.Checksum)
	s.Equal(s.User.ID, resp.CreatedBy.ID)

	// Verify object exists in MemoryStorage under expected key
	memStorage := testutil.GetState().Storage.(*storage.MemoryStorage)
	expectedKey := fmt.Sprintf("finance/%d/%s/%s-receipt.pdf", parish.ID, tx.ID.String(), resp.ID.String())
	data, contentType, err := memStorage.Get(expectedKey)
	s.NoError(err)
	s.Equal("application/pdf", contentType)
	s.Equal(validPDFBytes, data)
}

func (s *AttachmentSuite) TestUpload_PNG_Success() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "invoice.png", validPNGBytes)
	s.Created()

	resp := s.GetResponse().Data
	s.Equal("invoice.png", resp.Filename)
	s.Equal("image/png", resp.MimeType)
	s.Equal(int64(len(validPNGBytes)), resp.Size)

	hasher := sha256.New()
	hasher.Write(validPNGBytes)
	s.Equal(hex.EncodeToString(hasher.Sum(nil)), resp.Checksum)
}

func (s *AttachmentSuite) TestUpload_MIMESniffing_RejectsFakeExtension() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	// Filename says .pdf but content is HTML
	fakeContent := []byte("<!DOCTYPE html><html><body><h1>Fake PDF</h1></body></html>")
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "fake.pdf", fakeContent)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AttachmentSuite) TestUpload_Oversized_Returns413() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	oversized := make([]byte, 10*1024*1024+1)
	copy(oversized, "%PDF-1.4\n")

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "huge.pdf", oversized)
	s.AssertStatus(fiber.StatusRequestEntityTooLarge)
}

func (s *AttachmentSuite) TestUpload_MissingFileField_Returns400() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	// Pass empty fieldName to simulate missing "file" field
	s.RequestMultipart(uploadURL, fiber.MethodPost, "", "", nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *AttachmentSuite) TestUpload_PathTraversalFilename_Sanitized() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "../../secret/receipt.pdf", validPDFBytes)
	s.Created()

	resp := s.GetResponse().Data
	s.Equal("receipt.pdf", resp.Filename)

	memStorage := testutil.GetState().Storage.(*storage.MemoryStorage)
	expectedKey := fmt.Sprintf("finance/%d/%s/%s-receipt.pdf", parish.ID, tx.ID.String(), resp.ID.String())
	_, _, err := memStorage.Get(expectedKey)
	s.NoError(err)
}

func (s *AttachmentSuite) TestListAndDownload_Success() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt1.pdf", validPDFBytes)
	s.Created()
	att1 := s.GetResponse().Data

	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "receipt2.png", validPNGBytes)
	s.Created()
	att2 := s.GetResponse().Data

	// List
	s.Request(uploadURL, fiber.MethodGet, nil)
	s.OK()
	listResp := testutil.DecodeOther[[]model.AttachmentResponse](&s.ApiTestSuite).Data
	s.Len(listResp, 2)
	s.Equal(att1.ID, listResp[0].ID)
	s.Equal(att2.ID, listResp[1].ID)

	// Download att1
	downloadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments/%s/download", parish.ID, tx.ID, att1.ID)
	s.Request(downloadURL, fiber.MethodGet, nil)
	s.OK()
	dlResp := testutil.DecodeOther[model.DownloadResponse](&s.ApiTestSuite).Data
	s.NotEmpty(dlResp.URL)
	s.True(dlResp.ExpiresAt.After(time.Now()))
}

func (s *AttachmentSuite) TestIsolation_CrossParishReturns404() {
	parishA := CreateTestParish(s.T())
	parishB := CreateTestParish(s.T())
	s.grantRole(parishA.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(parishB.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cashA := s.createAccount(parishA.ID, "1001", "asset", true, "active")
	incomeA := s.createAccount(parishA.ID, "4001", "income", true, "active")
	txA := s.createPostedTransaction(parishA.ID, cashA.ID, incomeA.ID, "2026-10-15", 50000)

	uploadURL_A := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parishA.ID, txA.ID)
	s.RequestMultipart(uploadURL_A, fiber.MethodPost, "file", "file.pdf", validPDFBytes)
	s.Created()
	attA := s.GetResponse().Data

	// Try upload on parish B with txA ID -> 404
	uploadURL_B := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parishB.ID, txA.ID)
	s.RequestMultipart(uploadURL_B, fiber.MethodPost, "file", "file2.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusNotFound)

	// Try list on parish B with txA ID -> 404
	s.Request(uploadURL_B, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Try download on parish B with txA ID and attA ID -> 404
	dlURL_B := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments/%s/download", parishB.ID, txA.ID, attA.ID)
	s.Request(dlURL_B, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)

	// Try download on parish A but with mismatched transaction ID -> 404
	randomTxID := core.GenerateID()
	dlURL_MismatchedTx := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments/%s/download", parishA.ID, randomTxID, attA.ID)
	s.Request(dlURL_MismatchedTx, fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *AttachmentSuite) TestStorageRollback_OnInvalidTransaction() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	fakeTxID := core.GenerateID()
	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, fakeTxID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "orphaned.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusNotFound)

	// Verify no objects with fakeTxID prefix exist in storage
	memStorage := testutil.GetState().Storage.(*storage.MemoryStorage)
	prefix := fmt.Sprintf("finance/%d/%s/", parish.ID, fakeTxID.String())
	// In MemoryStorage, attempting to get any key with that fakeTxID returns ErrNotFound
	_, _, err := memStorage.Get(prefix + "orphaned.pdf")
	s.ErrorIs(err, storage.ErrNotFound)
}

func (s *AttachmentSuite) TestUpload_VoidTransactionAllowed() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	// Void transaction
	voidURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/void", parish.ID, tx.ID)
	s.Request(voidURL, fiber.MethodPost, model.VoidTransactionRequest{})
	s.OK()

	// Upload attachment to voided transaction
	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "void_receipt.pdf", validPDFBytes)
	s.Created()
	resp := s.GetResponse().Data
	s.Equal("void_receipt.pdf", resp.Filename)
}

func (s *AttachmentSuite) TestAuthz_ReaderVsWriter() {
	parish := CreateTestParish(s.T())
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash := s.createAccount(parish.ID, "1001", "asset", true, "active")
	income := s.createAccount(parish.ID, "4001", "income", true, "active")
	tx := s.createPostedTransaction(parish.ID, cash.ID, income.ID, "2026-10-15", 50000)

	uploadURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments", parish.ID, tx.ID)
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
	s.GrantRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelReader), authz.Scope{Type: authz.ScopeParish, ID: int64(parish.ID)})

	// Reader can list
	s.Request(uploadURL, fiber.MethodGet, nil)
	s.OK()

	// Reader can download
	dlURL := fmt.Sprintf("/parishes/%d/finance/transactions/%s/attachments/%s/download", parish.ID, tx.ID, att.ID)
	s.Request(dlURL, fiber.MethodGet, nil)
	s.OK()

	// Reader cannot upload (403)
	s.RequestMultipart(uploadURL, fiber.MethodPost, "file", "doc2.pdf", validPDFBytes)
	s.AssertStatus(fiber.StatusForbidden)
}

func TestAttachmentSuite(t *testing.T) {
	suite.Run(t, new(AttachmentSuite))
}
