package test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent/diocese"
	"github.com/paroki/domus/api/ent/user"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/system/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/httpx"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type ValidateTestSuite struct {
	testutil.ApiTestSuite[core.AuthenticatedUser]
}

func (s *ValidateTestSuite) TestUnknownUser_Returns403() {
	unknownID := core.GenerateID()
	unknownEmail := "unknown-" + unknownID.String() + "@example.com"

	claims := jwt.MapClaims{
		"sub":            unknownID.String(),
		"email":          unknownEmail,
		"name":           "Unknown User",
		"email_verified": true,
	}

	token := testutil.SignToken(s.T(), claims)
	s.RequestWithToken("/ping", fiber.MethodGet, nil, token)
	s.AssertStatus(fiber.StatusForbidden)

	// Verify that user was NOT created in database
	st := testutil.GetState()
	exists, err := st.Ent.User.Query().
		Where(user.EmailEqualFold(unknownEmail)).
		Exist(context.Background())
	s.Require().NoError(err)
	s.False(exists, "unknown user must not be auto-created in database")
}

func (s *ValidateTestSuite) TestSoftDeletedUser_Returns403() {
	st := testutil.GetState()
	uid := core.GenerateID()
	authID := uuid.New()
	email := "deleted-" + uid.String() + "@example.com"

	u, err := st.Ent.User.Create().
		SetID(uid).
		SetName("Soft Deleted User").
		SetEmail(email).
		SetAuthUserID(authID).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	// Soft delete the user
	err = st.Ent.User.DeleteOne(u).Exec(context.Background())
	s.Require().NoError(err)

	claims := jwt.MapClaims{
		"sub":            authID.String(),
		"email":          email,
		"name":           "Soft Deleted User",
		"email_verified": true,
	}

	token := testutil.SignToken(s.T(), claims)
	s.RequestWithToken("/ping", fiber.MethodGet, nil, token)
	s.AssertStatus(fiber.StatusForbidden)
}

func (s *ValidateTestSuite) TestInvitedUser_UnverifiedEmail_Returns403() {
	st := testutil.GetState()
	internalID := core.GenerateID()
	email := "unverified-" + internalID.String() + "@example.com"

	_, err := st.Ent.User.Create().
		SetID(internalID).
		SetName("Unverified Invited User").
		SetEmail(email).
		SetStatus(user.StatusInvited).
		Save(context.Background())
	s.Require().NoError(err)

	betterAuthID := uuid.New()
	claims := jwt.MapClaims{
		"sub":            betterAuthID.String(),
		"email":          email,
		"name":           "Unverified Invited User",
		"email_verified": false, // not verified!
	}

	token := testutil.SignToken(s.T(), claims)
	s.RequestWithToken("/ping", fiber.MethodGet, nil, token)
	s.AssertStatus(fiber.StatusForbidden)

	// User must remain invited and unbound in database
	dbUser, err := st.Ent.User.Get(context.Background(), internalID)
	s.Require().NoError(err)
	s.Equal(user.StatusInvited, dbUser.Status)
	s.Nil(dbUser.AuthUserID)
}

func (s *ValidateTestSuite) TestInvitedUser_FirstLogin_BindsAuthIDAndActivates() {
	st := testutil.GetState()
	internalID := core.GenerateID()
	email := "invited-" + internalID.String() + "@example.com"

	_, err := st.Ent.User.Create().
		SetID(internalID).
		SetName("Invited User").
		SetEmail(email).
		SetStatus(user.StatusInvited).
		Save(context.Background())
	s.Require().NoError(err)

	betterAuthID := uuid.New()
	claims := jwt.MapClaims{
		"sub":            betterAuthID.String(),
		"email":          email,
		"name":           "Invited User",
		"email_verified": true,
	}

	token := testutil.SignToken(s.T(), claims)
	s.RequestWithToken("/ping", fiber.MethodGet, nil, token)
	s.OK()

	res := s.GetResponse()
	// Response user ID must be internal API ID, NOT Better Auth sub
	s.Equal(internalID, res.Data.ID)
	s.Equal(email, res.Data.Email)

	// In DB: user status is active and auth_user_id is bound to betterAuthID
	dbUser, err := st.Ent.User.Get(context.Background(), internalID)
	s.Require().NoError(err)
	s.Equal(user.StatusActive, dbUser.Status)
	s.Require().NotNil(dbUser.AuthUserID)
	s.Equal(betterAuthID, *dbUser.AuthUserID)
}

func (s *ValidateTestSuite) TestActiveUser_SubsequentLogin_SucceedsWithoutRebinding() {
	st := testutil.GetState()
	internalID := core.GenerateID()
	betterAuthID := uuid.New()
	email := "active-" + internalID.String() + "@example.com"

	_, err := st.Ent.User.Create().
		SetID(internalID).
		SetName("Active User").
		SetEmail(email).
		SetAuthUserID(betterAuthID).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	claims := jwt.MapClaims{
		"sub":            betterAuthID.String(),
		"email":          email,
		"name":           "Active User",
		"email_verified": true,
	}

	token := testutil.SignToken(s.T(), claims)
	s.RequestWithToken("/ping", fiber.MethodGet, nil, token)
	s.OK()

	res := s.GetResponse()
	s.Equal(internalID, res.Data.ID)

	// DB row remains active with same auth_user_id
	dbUser, err := st.Ent.User.Get(context.Background(), internalID)
	s.Require().NoError(err)
	s.Equal(user.StatusActive, dbUser.Status)
	s.Require().NotNil(dbUser.AuthUserID)
	s.Equal(betterAuthID, *dbUser.AuthUserID)
}

func (s *ValidateTestSuite) TestCasbinEnforcerUsesInternalIDNotSub() {
	st := testutil.GetState()
	internalID := core.GenerateID()
	betterAuthID := uuid.New()
	email := "casbin-" + internalID.String() + "@example.com"

	_, err := st.Ent.User.Create().
		SetID(internalID).
		SetName("Casbin User").
		SetEmail(email).
		SetAuthUserID(betterAuthID).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	// Grant superadmin at system:0 to internalID (NOT to betterAuthID!)
	err = st.Authz.AddMembership(internalID.String(), authz.RoleSuperadmin, authz.System)
	s.Require().NoError(err)

	claims := jwt.MapClaims{
		"sub":            betterAuthID.String(),
		"email":          email,
		"name":           "Casbin User",
		"email_verified": true,
	}

	token := testutil.SignToken(s.T(), claims)

	// Access route protected by httpx.RequirePermission(SystemScope, ResourceDiocese, ActionRead)
	req := &http.Request{}
	_ = req
	s.RequestWithToken("/dioceses", fiber.MethodGet, nil, token)
	s.OK()
}

func (s *ValidateTestSuite) TestAuditFieldsUseInternalIDNotSub() {
	st := testutil.GetState()
	internalID := core.GenerateID()
	betterAuthID := uuid.New()
	email := "audit-" + internalID.String() + "@example.com"

	_, err := st.Ent.User.Create().
		SetID(internalID).
		SetName("Audit User").
		SetEmail(email).
		SetAuthUserID(betterAuthID).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	// Grant superadmin to internalID
	err = st.Authz.AddMembership(internalID.String(), authz.RoleSuperadmin, authz.System)
	s.Require().NoError(err)

	claims := jwt.MapClaims{
		"sub":            betterAuthID.String(),
		"email":          email,
		"name":           "Audit User",
		"email_verified": true,
	}

	token := testutil.SignToken(s.T(), claims)

	// Create diocese
	createReq := model.CreateDioceseRequest{
		Name: "Keuskupan Audit " + internalID.String()[:8],
	}

	// We use testutil ApiTestSuite helper
	var resp httpx.WebResponse[model.DioceseResponse]
	_ = resp
	s.RequestWithToken("/dioceses", fiber.MethodPost, createReq, token)
	s.Created()

	// Verify that created_by and updated_by match internalID, NOT betterAuthID
	d, err := st.Ent.Diocese.Query().
		Where(diocese.NameEQ(createReq.Name)).
		Only(context.Background())
	s.Require().NoError(err)
	s.Equal(internalID, d.CreatedBy, "created_by must be internal API ID")
	s.NotEqual(betterAuthID, d.CreatedBy, "created_by must NOT be Better Auth sub")
	s.Equal(internalID, d.UpdatedBy, "updated_by must be internal API ID")
}

func TestValidateSuite(t *testing.T) {
	suite.Run(t, new(ValidateTestSuite))
}
