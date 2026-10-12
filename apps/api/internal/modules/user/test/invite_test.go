package test

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/user/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

// InviteSuite tests the POST /users endpoint across various scenarios and roles.
type InviteSuite struct {
	testutil.ApiTestSuite[model.UserInviteResponse]
}

// ensureUserExists creates the test user in the database if not present, satisfying AuditMixins FK.
func (s *InviteSuite) ensureUserExists() {
	s.T().Helper()
	exists, err := testutil.GetState().Ent.User.Query().
		All(core.SkipSoftDelete(context.Background()))
	s.Require().NoError(err)
	for _, u := range exists {
		if u.ID == s.User.ID {
			return
		}
	}
	_, err = testutil.GetState().Ent.User.Create().
		SetID(s.User.ID).
		SetName(s.User.Name).
		SetEmail(s.User.Email).
		Save(context.Background())
	s.Require().NoError(err)
}

// createDiocese is a test helper that inserts a Diocese record.
func (s *InviteSuite) createDiocese(name string) *ent.Diocese {
	s.T().Helper()
	s.ensureUserExists()
	d, err := testutil.GetState().Ent.Diocese.Create().
		SetName(name).
		SetCreatedBy(s.User.ID).
		SetUpdatedBy(s.User.ID).
		Save(context.Background())
	s.Require().NoError(err)
	return d
}

// createParish is a test helper that inserts a Parish linked to a Diocese.
func (s *InviteSuite) createParish(name string, diocese *ent.Diocese) *ent.Parish {
	s.T().Helper()
	b := testutil.GetState().Ent.Parish.Create().SetName(name)
	if diocese != nil {
		b.SetDiocese(diocese)
	}
	p, err := b.Save(context.Background())
	s.Require().NoError(err)
	return p
}

// createWorkspace is a test helper that inserts a Workspace linked to a Parish.
func (s *InviteSuite) createWorkspace(name string, parish *ent.Parish) *ent.Workspace {
	s.T().Helper()
	w, err := testutil.GetState().Ent.Workspace.Create().
		SetName(name).
		SetParish(parish).
		SetCreatedBy(s.User.ID).
		SetUpdatedBy(s.User.ID).
		Save(context.Background())
	s.Require().NoError(err)
	return w
}

// TestSuperadminInvites_AnyScope verifies that superadmin can invite users at any scope level.
func (s *InviteSuite) TestSuperadminInvites_AnyScope() {
	s.GrantRole(authz.RoleSuperadmin, authz.System)

	dio := s.createDiocese("Diocese Superadmin Test")
	par := s.createParish("Parish Superadmin Test", dio)
	ws := s.createWorkspace("Workspace Superadmin Test", par)

	req := model.InviteUserRequest{
		Email: "invited-superadmin@example.com",
		Name:  "Invited By Superadmin",
		Memberships: []model.MembershipInput{
			{ScopeType: "diocese", ScopeID: int64(dio.ID), Role: "users-writer"},
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-admin"},
			{ScopeType: "workspace", ScopeID: int64(ws.ID), Role: "finance-writer"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("invited-superadmin@example.com", res.Data.Email)
	s.Equal("Invited By Superadmin", res.Data.Name)
	s.Equal("invited", res.Data.Status)
	s.Len(res.Data.Memberships, 3)

	// Verify enforcer has the memberships registered
	can, err := testutil.GetState().Authz.Can(context.Background(), res.Data.ID.String(), authz.Scope{Type: authz.ScopeParish, ID: int64(par.ID)}, authz.ResourceFinance, authz.ActionWrite)
	s.Require().NoError(err)
	s.True(can)
}

// TestParishAdminInvites_OwnParish verifies that parish admin can invite into their own parish.
func (s *InviteSuite) TestParishAdminInvites_OwnParish() {
	dio := s.createDiocese("Diocese Parish Admin Test")
	par := s.createParish("Parish Own", dio)

	s.GrantRole("users-admin", authz.Scope{Type: authz.ScopeParish, ID: int64(par.ID)})

	req := model.InviteUserRequest{
		Email: "invited-parish@example.com",
		Name:  "Invited Parish Member",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-reader"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("invited-parish@example.com", res.Data.Email)
	s.Equal("invited", res.Data.Status)
	s.Len(res.Data.Memberships, 1)
}

// TestParishAdminInvites_OtherParishRejected verifies that parish admin cannot invite into another parish.
func (s *InviteSuite) TestParishAdminInvites_OtherParishRejected() {
	dio := s.createDiocese("Diocese Parish Admin Rejection")
	parOwn := s.createParish("Parish Admin Own", dio)
	parOther := s.createParish("Parish Admin Other", dio)

	s.GrantRole("users-admin", authz.Scope{Type: authz.ScopeParish, ID: int64(parOwn.ID)})

	req := model.InviteUserRequest{
		Email: "parish-other-attempt@example.com",
		Name:  "Parish Other Attempt",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(parOther.ID), Role: "finance-reader"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusForbidden)
}

// TestDioceseAdminInvites_ParishOfThatDiocese verifies that diocese admin can invite into any parish within their diocese.
func (s *InviteSuite) TestDioceseAdminInvites_ParishOfThatDiocese() {
	dio := s.createDiocese("Diocese Admin Diocese")
	par := s.createParish("Parish In Diocese", dio)

	s.GrantRole("users-admin", authz.Scope{Type: authz.ScopeDiocese, ID: int64(dio.ID)})

	req := model.InviteUserRequest{
		Email: "invited-by-dio@example.com",
		Name:  "Invited By Diocese Admin",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-writer"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("invited-by-dio@example.com", res.Data.Email)
	s.Len(res.Data.Memberships, 1)
}

// TestDioceseAdminInvites_ParishOfOtherDioceseRejected verifies diocese admin is rejected for a parish in another diocese.
func (s *InviteSuite) TestDioceseAdminInvites_ParishOfOtherDioceseRejected() {
	dioOwn := s.createDiocese("Diocese Admin Own Dio")
	dioOther := s.createDiocese("Diocese Admin Other Dio")
	parOther := s.createParish("Parish In Other Dio", dioOther)

	s.GrantRole("users-admin", authz.Scope{Type: authz.ScopeDiocese, ID: int64(dioOwn.ID)})

	req := model.InviteUserRequest{
		Email: "dio-other-attempt@example.com",
		Name:  "Diocese Other Attempt",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(parOther.ID), Role: "finance-writer"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusForbidden)
}

// TestWorkspaceScope_ResolvedViaParish verifies authority over workspace scope resolves to parent parish rule.
func (s *InviteSuite) TestWorkspaceScope_ResolvedViaParish() {
	dio := s.createDiocese("Diocese Workspace Authority")
	par := s.createParish("Parish Workspace Owner", dio)
	ws := s.createWorkspace("Workspace Inside Parish", par)

	// Parish admin of parent parish should be able to grant workspace membership
	s.GrantRole("users-admin", authz.Scope{Type: authz.ScopeParish, ID: int64(par.ID)})

	req := model.InviteUserRequest{
		Email: "workspace-member@example.com",
		Name:  "Workspace Member",
		Memberships: []model.MembershipInput{
			{ScopeType: "workspace", ScopeID: int64(ws.ID), Role: "finance-reader"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("workspace-member@example.com", res.Data.Email)
	s.Len(res.Data.Memberships, 1)
	s.Equal("workspace", res.Data.Memberships[0].ScopeType)
	s.Equal(int64(ws.ID), res.Data.Memberships[0].ScopeID)
}

// TestGrantingSuperadmin_AlwaysRejected verifies that granting the superadmin role is always rejected.
func (s *InviteSuite) TestGrantingSuperadmin_AlwaysRejected() {
	s.GrantRole(authz.RoleSuperadmin, authz.System)

	req := model.InviteUserRequest{
		Email: "superadmin-grant-attempt@example.com",
		Name:  "Superadmin Grant Attempt",
		Memberships: []model.MembershipInput{
			{ScopeType: "system", ScopeID: 0, Role: "superadmin"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusForbidden)
}

// TestReinviteExistingEmail_Returns200AndAddsOnlyMissingMemberships tests idempotency and deduplication.
func (s *InviteSuite) TestReinviteExistingEmail_Returns200AndAddsOnlyMissingMemberships() {
	dio := s.createDiocese("Diocese Reinvite")
	par := s.createParish("Parish Reinvite", dio)

	s.GrantRole(authz.RoleSuperadmin, authz.System)

	// First invite
	req1 := model.InviteUserRequest{
		Email: "reinvite-test@example.com",
		Name:  "Initial User Name",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-reader"},
		},
	}
	s.Request("/users", fiber.MethodPost, req1)
	s.Created()
	initialID := s.GetResponse().Data.ID

	// Second invite with the duplicate membership plus a new one
	req2 := model.InviteUserRequest{
		Email: "reinvite-test@example.com",
		Name:  "Initial User Name",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-reader"},
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-writer"},
		},
	}
	s.Request("/users", fiber.MethodPost, req2)
	s.OK() // Status 200 for existing user

	res2 := s.GetResponse()
	s.Equal(initialID, res2.Data.ID)
	// Should have 2 unique memberships, not 3
	s.Len(res2.Data.Memberships, 2)
}

// TestInvalidEmail_Returns422 tests validation failure on bad email.
func (s *InviteSuite) TestInvalidEmail_Returns422() {
	s.GrantRole(authz.RoleSuperadmin, authz.System)

	req := model.InviteUserRequest{
		Email: "invalid-email-address",
		Name:  "Invalid Email User",
		Memberships: []model.MembershipInput{
			{ScopeType: "system", ScopeID: 0, Role: "finance-reader"},
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

// TestForbiddenMembershipInBatch_WritesNothing tests atomicity when one membership in batch is unauthorized.
func (s *InviteSuite) TestForbiddenMembershipInBatch_WritesNothing() {
	dio := s.createDiocese("Diocese Batch Test")
	parOwn := s.createParish("Parish Batch Own", dio)
	parOther := s.createParish("Parish Batch Other", dio)

	s.GrantRole("users-admin", authz.Scope{Type: authz.ScopeParish, ID: int64(parOwn.ID)})

	targetEmail := "batch-fail-atomic@example.com"
	req := model.InviteUserRequest{
		Email: targetEmail,
		Name:  "Batch Fail User",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(parOwn.ID), Role: "finance-reader"},
			{ScopeType: "parish", ScopeID: int64(parOther.ID), Role: "finance-reader"}, // Not allowed!
		},
	}

	s.Request("/users", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusForbidden)

	// Ensure the user was not written to database at all
	u, err := testutil.GetState().Ent.User.Query().
		All(core.SkipSoftDelete(context.Background()))
	s.Require().NoError(err)
	for _, userRecord := range u {
		s.NotEqual(targetEmail, userRecord.Email, "User should not exist after atomic rollback")
	}
}

// TestSoftDeletedEmail_CanBeReinvited tests that a soft-deleted email does not violate unique index constraints.
func (s *InviteSuite) TestSoftDeletedEmail_CanBeReinvited() {
	dio := s.createDiocese("Diocese SoftDelete Test")
	par := s.createParish("Parish SoftDelete", dio)

	s.GrantRole(authz.RoleSuperadmin, authz.System)

	email := "soft-deleted-user@example.com"
	req := model.InviteUserRequest{
		Email: email,
		Name:  "To Be Deleted",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-reader"},
		},
	}

	// 1. Create user
	s.Request("/users", fiber.MethodPost, req)
	s.Created()
	createdID := s.GetResponse().Data.ID

	// 2. Soft-delete user
	err := testutil.GetState().Ent.User.DeleteOneID(createdID).Exec(context.Background())
	s.Require().NoError(err)

	// 3. Invite the same email again -> should succeed with a new user (status 201)
	req2 := model.InviteUserRequest{
		Email: email,
		Name:  "Re-invited User",
		Memberships: []model.MembershipInput{
			{ScopeType: "parish", ScopeID: int64(par.ID), Role: "finance-writer"},
		},
	}
	s.Request("/users", fiber.MethodPost, req2)
	s.Created()

	res2 := s.GetResponse()
	s.NotEqual(createdID, res2.Data.ID, "New user ID should be generated after soft delete")
	s.Equal(email, res2.Data.Email)
}

func TestInviteSuite(t *testing.T) {
	suite.Run(t, new(InviteSuite))
}
