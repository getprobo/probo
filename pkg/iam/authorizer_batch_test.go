// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/iam/scopeset"
	"go.probo.inc/probo/pkg/mail"
)

type batchAuthorizeFixture struct {
	tenantID        gid.TenantID
	identityID      gid.GID
	membershipID    gid.GID
	organizationID  gid.GID
	organization2ID gid.GID
	frameworkID1    gid.GID
	frameworkID2    gid.GID
	frameworkID3    gid.GID
}

func TestAuthorizer_AuthorizeBatch(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)

		scope, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		assert.Equal(t, fixture.tenantID, scope.GetTenantID())
		assert.Equal(t, 2, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("mixed organization batch", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID3,
				},
			},
		)
		require.Error(t, err)

		errMixedOrg, ok := err.(*iam.ErrMixedOrganizationBatch)
		require.True(t, ok)
		assert.ElementsMatch(
			t,
			[]string{
				fixture.organizationID.String(),
				fixture.organization2ID.String(),
			},
			errMixedOrg.OrganizationIDs,
		)
	})

	t.Run("mixed entity type batch", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.organizationID,
				},
			},
		)
		require.Error(t, err)

		errMixedEntityType, ok := err.(*iam.ErrMixedEntityTypeBatch)
		require.True(t, ok)
		assert.Equal(
			t,
			[]uint16{coredata.OrganizationEntityType, coredata.FrameworkEntityType},
			errMixedEntityType.EntityTypes,
		)
	})

	t.Run("unsupported resource type for batch attributes", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					gid.New(fixture.tenantID, coredata.OAuth2RefreshTokenEntityType),
				},
			},
		)
		require.Error(t, err)

		errUnsupported, ok := errors.AsType[*iam.ErrBatchAuthorizationUnsupportedResourceType](err)
		require.True(t, ok)
		assert.Equal(t, coredata.OAuth2RefreshTokenEntityType, errUnsupported.EntityType)
	})

	t.Run("single deny rolls back entire batch", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, &fixture.frameworkID1)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
			},
		)
		require.Error(t, err)

		_, ok := err.(*iam.ErrInsufficientPermissions)
		require.True(t, ok)
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("duplicate resources in batch", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)

		scope, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID1,
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		assert.Equal(t, fixture.tenantID, scope.GetTenantID())
		assert.Equal(t, 2, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()

		authorizer := iam.NewAuthorizer(nil, log.NewLogger(log.WithOutput(io.Discard)), scopeset.New())

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: gid.New(gid.NilTenant, coredata.IdentityEntityType),
				Action:    newBatchTestAction(),
			},
		)
		require.Error(t, err)

		_, ok := err.(*iam.ErrEmptyResourceBatch)
		require.True(t, ok)
	})

	t.Run("dry-run does not write audit logs", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)

		scope, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
				DryRun: true,
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		assert.Equal(t, fixture.tenantID, scope.GetTenantID())
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("missing principal identity returns wrapped principal attributes error", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)
		missingPrincipalID := gid.New(gid.NilTenant, coredata.IdentityEntityType)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: missingPrincipalID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
				},
			},
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot build principal attributes")
		assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
	})

	t.Run("bulk insert failure aborts transaction", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithIdentityScopedStatements(
			client,
			policy.Allow(action).WithSID("allow-audit-log-failure"),
		)
		nonExistentOrgID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)

		scope, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
				ResourceAttributes: policy.Attributes{
					"organization_id": nonExistentOrgID.String(),
				},
			},
		)
		require.Error(t, err)
		assert.Nil(t, scope)
		assert.ErrorContains(t, err, "cannot commit transaction")
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("assumption required", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)
		sessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Session:   &sessionID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
			},
		)
		require.Error(t, err)

		_, ok := err.(*iam.ErrAssumptionRequired)
		require.True(t, ok)
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("assumption succeeds with active child session", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)
		rootSessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)

		insertBatchTestChildSession(
			t,
			context.Background(),
			client,
			fixture,
			rootSessionID,
			false,
		)

		scope, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Session:   &rootSessionID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		assert.Equal(t, fixture.tenantID, scope.GetTenantID())
		assert.Equal(t, 2, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("assumption fails when child session is expired", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)
		rootSessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)

		insertBatchTestChildSession(
			t,
			context.Background(),
			client,
			fixture,
			rootSessionID,
			true,
		)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: fixture.identityID,
				Session:   &rootSessionID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
					fixture.frameworkID2,
				},
			},
		)
		require.Error(t, err)

		_, ok := errors.AsType[*iam.ErrAssumptionRequired](err)
		require.True(t, ok)
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("no membership ignores assumption check", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizer(client, action, nil)
		rootSessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)
		identityWithoutMembershipID := insertBatchTestIdentity(
			t,
			context.Background(),
			client,
			fixture.tenantID,
		)

		_, err := authorizer.AuthorizeBatch(
			context.Background(),
			iam.AuthorizeBatchParams{
				Principal: identityWithoutMembershipID,
				Session:   &rootSessionID,
				Action:    action,
				Resources: []gid.GID{
					fixture.frameworkID1,
				},
			},
		)
		require.Error(t, err)

		_, hasAssumptionErr := errors.AsType[*iam.ErrAssumptionRequired](err)
		assert.False(t, hasAssumptionErr)

		insufficientPermissionsErr, ok := errors.AsType[*iam.ErrInsufficientPermissions](err)
		require.True(t, ok)
		assert.Equal(t, identityWithoutMembershipID, insufficientPermissionsErr.IdentityID)
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})
}

func TestAuthorizer_AuthorizeMulti(t *testing.T) {
	t.Parallel()

	t.Run("returns nil decisions when every item is allowed", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).WithSID("allow-evaluate-multi-all"),
		)

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource: fixture.frameworkID1,
						Action:   action,
					},
					{
						Resource: fixture.frameworkID2,
						Action:   action,
					},
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		assert.Equal(t, fixture.tenantID, scope.GetTenantID())
		require.Len(t, decisions, 2)
		assert.NoError(t, decisions[0])
		assert.NoError(t, decisions[1])
		assert.Equal(t, 2, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("returns per-item decisions on partial denial without aborting", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).
				WithSID("allow-only-first-framework").
				When(policy.Equals("resource.id", fixture.frameworkID1.String())),
		)

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource: fixture.frameworkID1,
						Action:   action,
					},
					{
						Resource: fixture.frameworkID2,
						Action:   action,
					},
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		assert.Equal(t, fixture.tenantID, scope.GetTenantID())
		require.Len(t, decisions, 2)
		assert.NoError(t, decisions[0])
		require.Error(t, decisions[1])

		denied, ok := errors.AsType[*iam.ErrInsufficientPermissions](decisions[1])
		require.True(t, ok)
		assert.Equal(t, fixture.frameworkID2, denied.EntityID)

		assert.Equal(t, 1, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("writes no audit logs when every item is denied", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).
				WithSID("allow-no-frameworks").
				When(policy.Equals("resource.id", "missing")),
		)

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource: fixture.frameworkID1,
						Action:   action,
					},
					{
						Resource: fixture.frameworkID2,
						Action:   action,
					},
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		require.Len(t, decisions, 2)
		require.Error(t, decisions[0])
		require.Error(t, decisions[1])
		assert.Equal(t, 0, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("skips audit log entries for dry-run allowed items", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).WithSID("allow-evaluate-dry-run"),
		)

		_, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource: fixture.frameworkID1,
						Action:   action,
						DryRun:   true,
					},
					{
						Resource: fixture.frameworkID2,
						Action:   action,
					},
				},
			},
		)
		require.NoError(t, err)
		require.Len(t, decisions, 2)
		assert.NoError(t, decisions[0])
		assert.NoError(t, decisions[1])
		assert.Equal(t, 1, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("rejects mixed organization batch", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).WithSID("allow-evaluate-mixed-org"),
		)

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource: fixture.frameworkID1,
						Action:   action,
					},
					{
						Resource: fixture.frameworkID3,
						Action:   action,
					},
				},
			},
		)
		require.Error(t, err)
		assert.Nil(t, scope)
		assert.Nil(t, decisions)

		_, ok := errors.AsType[*iam.ErrMixedOrganizationBatch](err)
		require.True(t, ok)
	})

	t.Run("rejects empty items", func(t *testing.T) {
		t.Parallel()

		authorizer := iam.NewAuthorizer(nil, log.NewLogger(log.WithOutput(io.Discard)), scopeset.New())

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: gid.New(gid.NilTenant, coredata.IdentityEntityType),
			},
		)
		require.Error(t, err)
		assert.Nil(t, scope)
		assert.Nil(t, decisions)

		_, ok := errors.AsType[*iam.ErrEmptyResourceBatch](err)
		require.True(t, ok)
	})

	t.Run("rejects unsupported principal type", func(t *testing.T) {
		t.Parallel()

		authorizer := iam.NewAuthorizer(nil, log.NewLogger(log.WithOutput(io.Discard)), scopeset.New())

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
				Items: []iam.MultiAuthorizeItem{
					{
						Resource: gid.New(gid.NewTenantID(), coredata.FrameworkEntityType),
						Action:   "core:test:get",
					},
				},
			},
		)
		require.Error(t, err)
		assert.Nil(t, scope)
		assert.Nil(t, decisions)

		_, ok := errors.AsType[*iam.ErrUnsupportedPrincipalType](err)
		require.True(t, ok)
	})

	t.Run("assumption error is recorded only on items that require the check", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).WithSID("allow-evaluate-mixed-skip-assumption"),
		)
		sessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Session:   &sessionID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource:            fixture.frameworkID1,
						Action:              action,
						SkipAssumptionCheck: true,
					},
					{
						Resource: fixture.frameworkID2,
						Action:   action,
					},
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		require.Len(t, decisions, 2)
		assert.NoError(t, decisions[0])
		require.Error(t, decisions[1])

		_, ok := errors.AsType[*iam.ErrAssumptionRequired](decisions[1])
		require.True(t, ok)

		assert.Equal(t, 1, countAuditLogsForAction(t, context.Background(), client, action))
	})

	t.Run("per-item resource attributes are honoured by the policy", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		fixture := seedBatchAuthorizeFixture(t, context.Background(), client)
		action := newBatchTestAction()
		authorizer := newTestAuthorizerWithStatements(
			client,
			policy.Allow(action).
				WithSID("allow-when-resource-flag-set").
				When(policy.Equals("resource.flag", "on")),
		)

		scope, decisions, err := authorizer.AuthorizeMulti(
			context.Background(),
			iam.AuthorizeMultiParams{
				Principal: fixture.identityID,
				Items: []iam.MultiAuthorizeItem{
					{
						Resource:           fixture.frameworkID1,
						Action:             action,
						ResourceAttributes: policy.Attributes{"flag": "on"},
					},
					{
						Resource:           fixture.frameworkID2,
						Action:             action,
						ResourceAttributes: policy.Attributes{"flag": "off"},
					},
				},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, scope)
		require.Len(t, decisions, 2)
		assert.NoError(t, decisions[0])
		require.Error(t, decisions[1])

		denied, ok := errors.AsType[*iam.ErrInsufficientPermissions](decisions[1])
		require.True(t, ok)
		assert.Equal(t, fixture.frameworkID2, denied.EntityID)

		assert.Equal(t, 1, countAuditLogsForAction(t, context.Background(), client, action))
	})
}

func newTestAuthorizer(client *pg.Client, action string, allowResourceID *gid.GID) *iam.Authorizer {
	statement := policy.Allow(action).WithSID("allow-test-action")
	if allowResourceID != nil {
		statement = statement.When(policy.Equals("resource.id", allowResourceID.String()))
	}

	return newTestAuthorizerWithStatements(client, statement)
}

func newTestAuthorizerWithStatements(client *pg.Client, statements ...policy.Statement) *iam.Authorizer {
	authorizer := iam.NewAuthorizer(client, log.NewLogger(log.WithOutput(io.Discard)), scopeset.New())
	authorizer.RegisterPolicySet(
		iam.NewPolicySet().AddRolePolicy(
			string(coredata.MembershipRoleOwner),
			policy.NewPolicy("batch-authorize-test", "Batch Authorize Test", statements...),
		),
	)

	return authorizer
}

func newTestAuthorizerWithIdentityScopedStatements(client *pg.Client, statements ...policy.Statement) *iam.Authorizer {
	authorizer := iam.NewAuthorizer(client, log.NewLogger(log.WithOutput(io.Discard)), scopeset.New())
	authorizer.RegisterPolicySet(
		iam.NewPolicySet().AddIdentityScopedPolicy(
			policy.NewPolicy("batch-authorize-identity-test", "Batch Authorize Identity Test", statements...),
		),
	)

	return authorizer
}

func seedBatchAuthorizeFixture(t *testing.T, ctx context.Context, client *pg.Client) batchAuthorizeFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	organization2ID := gid.New(tenantID, coredata.OrganizationEntityType)
	membershipID := gid.New(tenantID, coredata.MembershipEntityType)
	profileID := gid.New(tenantID, coredata.MembershipProfileEntityType)
	frameworkID1 := gid.New(tenantID, coredata.FrameworkEntityType)
	frameworkID2 := gid.New(tenantID, coredata.FrameworkEntityType)
	frameworkID3 := gid.New(tenantID, coredata.FrameworkEntityType)
	now := time.Now().UTC()

	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", tenantID))
	require.NoError(t, err)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		identity := coredata.Identity{
			ID:                   identityID,
			EmailAddress:         emailAddress,
			FullName:             "Batch Test User",
			EmailAddressVerified: true,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := identity.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert identity: %w", err)
		}

		organization := coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Batch Test Org A",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := organization.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert first organization: %w", err)
		}

		organization2 := coredata.Organization{
			ID:        organization2ID,
			TenantID:  tenantID,
			Name:      "Batch Test Org B",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := organization2.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert second organization: %w", err)
		}

		membership := coredata.Membership{
			ID:             membershipID,
			IdentityID:     identityID,
			OrganizationID: organizationID,
			Role:           coredata.MembershipRoleOwner,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := membership.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert membership: %w", err)
		}

		profile := coredata.MembershipProfile{
			ID:             profileID,
			IdentityID:     identityID,
			OrganizationID: organizationID,
			Source:         coredata.ProfileSourceManual,
			State:          coredata.ProfileStateActive,
			FullName:       "Batch Test User",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := profile.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert membership profile: %w", err)
		}

		framework1 := coredata.Framework{
			ID:             frameworkID1,
			OrganizationID: organizationID,
			ReferenceID:    fmt.Sprintf("batch-test-framework-1-%s", tenantID),
			Name:           "Batch Test Framework 1",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := framework1.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert first framework: %w", err)
		}

		framework2 := coredata.Framework{
			ID:             frameworkID2,
			OrganizationID: organizationID,
			ReferenceID:    fmt.Sprintf("batch-test-framework-2-%s", tenantID),
			Name:           "Batch Test Framework 2",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := framework2.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert second framework: %w", err)
		}

		framework3 := coredata.Framework{
			ID:             frameworkID3,
			OrganizationID: organization2ID,
			ReferenceID:    fmt.Sprintf("batch-test-framework-3-%s", tenantID),
			Name:           "Batch Test Framework 3",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := framework3.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert third framework: %w", err)
		}

		return nil
	}))

	return batchAuthorizeFixture{
		tenantID:        tenantID,
		identityID:      identityID,
		membershipID:    membershipID,
		organizationID:  organizationID,
		organization2ID: organization2ID,
		frameworkID1:    frameworkID1,
		frameworkID2:    frameworkID2,
		frameworkID3:    frameworkID3,
	}
}

func insertBatchTestIdentity(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	tenantID gid.TenantID,
) gid.GID {
	t.Helper()

	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	now := time.Now().UTC()

	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s-no-membership@example.com", tenantID))
	require.NoError(t, err)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		identity := coredata.Identity{
			ID:                   identityID,
			EmailAddress:         emailAddress,
			FullName:             "Batch Test No Membership User",
			EmailAddressVerified: true,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := identity.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert identity without membership: %w", err)
		}

		return nil
	}))

	return identityID
}

func insertBatchTestChildSession(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fixture batchAuthorizeFixture,
	rootSessionID gid.GID,
	expired bool,
) gid.GID {
	t.Helper()

	childSessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)
	now := time.Now().UTC()
	expiredAt := now.Add(30 * time.Minute)

	var expireReason *coredata.ExpireReason

	if expired {
		reason := coredata.ExpireReasonRevoked
		expireReason = &reason
		expiredAt = now.Add(-1 * time.Minute)
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		rootSession := coredata.Session{
			ID:              rootSessionID,
			IdentityID:      fixture.identityID,
			Data:            coredata.SessionData{},
			AuthMethod:      coredata.AuthMethodPassword,
			AuthenticatedAt: now,
			UserAgent:       "batch-test-root-agent",
			IPAddress:       net.ParseIP("127.0.0.1"),
			ExpiredAt:       now.Add(30 * time.Minute),
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := rootSession.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert root session: %w", err)
		}

		session := coredata.Session{
			ID:              childSessionID,
			IdentityID:      fixture.identityID,
			TenantID:        &fixture.tenantID,
			MembershipID:    &fixture.membershipID,
			ParentSessionID: &rootSessionID,
			Data:            coredata.SessionData{},
			AuthMethod:      coredata.AuthMethodPassword,
			AuthenticatedAt: now,
			UserAgent:       "batch-test-agent",
			IPAddress:       net.ParseIP("127.0.0.1"),
			ExpireReason:    expireReason,
			ExpiredAt:       expiredAt,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := session.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert child session: %w", err)
		}

		return nil
	}))

	return childSessionID
}

func countAuditLogsForAction(t *testing.T, ctx context.Context, client *pg.Client, action string) int {
	t.Helper()

	var count int

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := tx.QueryRow(
			ctx,
			"SELECT COUNT(id) FROM audit_log_entries WHERE action = $1",
			action,
		).Scan(&count); err != nil {
			return fmt.Errorf("cannot query audit log count: %w", err)
		}

		return nil
	}))

	return count
}

func newBatchTestAction() string {
	return fmt.Sprintf("test:framework-%d:get", time.Now().UnixNano())
}
