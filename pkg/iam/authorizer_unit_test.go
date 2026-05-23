// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package iam

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

func TestAuthorizer_ValidateInputs(t *testing.T) {
	t.Parallel()

	t.Run("authorize rejects unsupported principal type", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{}
		_, err := a.Authorize(
			context.Background(),
			AuthorizeParams{
				Principal: gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
			},
		)
		require.Error(t, err)

		errUnsupported, ok := errors.AsType[*ErrUnsupportedPrincipalType](err)
		require.True(t, ok)
		assert.Equal(t, coredata.OrganizationEntityType, errUnsupported.EntityType)
	})

	t.Run("authorize batch rejects unsupported principal type", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{}
		_, err := a.AuthorizeBatch(
			context.Background(),
			AuthorizeBatchParams{
				Principal: gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
				Resources: []gid.GID{gid.New(gid.NewTenantID(), coredata.FrameworkEntityType)},
			},
		)
		require.Error(t, err)

		errUnsupported, ok := errors.AsType[*ErrUnsupportedPrincipalType](err)
		require.True(t, ok)
		assert.Equal(t, coredata.OrganizationEntityType, errUnsupported.EntityType)
	})

	t.Run("authorize batch rejects empty resources", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{}
		_, err := a.AuthorizeBatch(
			context.Background(),
			AuthorizeBatchParams{
				Principal: gid.New(gid.NilTenant, coredata.IdentityEntityType),
			},
		)
		require.Error(t, err)
		_, ok := errors.AsType[*ErrEmptyResourceBatch](err)
		require.True(t, ok)
	})

}

func TestAuthorizer_InternalErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	unknownResourceID := gid.New(gid.NewTenantID(), 65535)
	unsupportedResourceID := gid.New(gid.NewTenantID(), coredata.OAuth2AccessTokenEntityType)

	a := &Authorizer{
		evaluator: policy.NewEvaluator(),
		policySet: NewPolicySet(),
	}

	t.Run("authorize batch returns wrapped resource attributes batch error", func(t *testing.T) {
		t.Parallel()

		_, err := a.authorizeMulti(
			ctx,
			nil,
			AuthorizeMultiParams{
				Principal: identityID,
				Items: []MultiAuthorizeItem{
					{
						Resource: unknownResourceID,
						Action:   "core:test:list",
					},
				},
			},
			nil,
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot build resource attributes batch")
	})

	t.Run("authorize batch rejects mixed entity types", func(t *testing.T) {
		t.Parallel()

		firstResourceID := gid.New(gid.NewTenantID(), coredata.FrameworkEntityType)
		secondResourceID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)

		_, err := a.AuthorizeBatch(
			ctx,
			AuthorizeBatchParams{
				Principal: identityID,
				Action:    "core:test:list",
				Resources: []gid.GID{firstResourceID, secondResourceID},
			},
		)
		require.Error(t, err)

		errMixedEntityType, ok := errors.AsType[*ErrMixedEntityTypeBatch](err)
		require.True(t, ok)
		assert.Equal(
			t,
			[]uint16{coredata.OrganizationEntityType, coredata.FrameworkEntityType},
			errMixedEntityType.EntityTypes,
		)
	})

	t.Run("build principal attributes rejects unsupported batch interface type", func(t *testing.T) {
		t.Parallel()

		_, err := a.buildPrincipalAttributes(
			ctx,
			nil,
			unsupportedResourceID,
			map[string]string{"role": "OWNER"},
		)
		require.Error(t, err)
		errUnsupported, ok := errors.AsType[*ErrBatchAuthorizationUnsupportedResourceType](err)
		require.True(t, ok)
		assert.Equal(t, coredata.OAuth2AccessTokenEntityType, errUnsupported.EntityType)
	})

	t.Run("build principal attributes keeps defaults when entity type is unknown", func(t *testing.T) {
		t.Parallel()

		attrs, err := a.buildPrincipalAttributes(
			ctx,
			nil,
			unknownResourceID,
			map[string]string{"role": "OWNER"},
		)
		require.NoError(t, err)
		assert.Equal(t, unknownResourceID.String(), attrs["id"])
		assert.Equal(t, "OWNER", attrs["role"])
	})
}

func TestAuthorizer_HelperMethods(t *testing.T) {
	t.Parallel()

	t.Run("check assumption short-circuits", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{}
		identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
		sessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)
		membership := &coredata.Membership{ID: gid.New(gid.NewTenantID(), coredata.MembershipEntityType)}

		require.NoError(t, a.checkAssumption(context.Background(), nil, identityID, nil, membership, false))
		require.NoError(t, a.checkAssumption(context.Background(), nil, identityID, &sessionID, nil, false))
		require.NoError(t, a.checkAssumption(context.Background(), nil, identityID, &sessionID, membership, true))
	})

	t.Run("build policies for role includes identity scoped policies", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{
			policySet: NewPolicySet().
				AddIdentityScopedPolicy(policy.NewPolicy("identity", "Identity", policy.Allow("identity:read"))).
				AddRolePolicy("OWNER", policy.NewPolicy("owner", "Owner", policy.Allow("core:*"))),
		}

		withRole := a.buildPoliciesForRole("OWNER")
		require.Len(t, withRole, 2)

		withoutRole := a.buildPoliciesForRole("VIEWER")
		require.Len(t, withoutRole, 1)
	})

	t.Run("unique sorted strings deduplicates and sorts", func(t *testing.T) {
		t.Parallel()

		got := uniqueSortedStrings([]string{"b", "a", "b", "", "c", "a"})
		assert.Equal(t, []string{"", "a", "b", "c"}, got)
	})

	t.Run("unique sorted entity types deduplicates and sorts", func(t *testing.T) {
		t.Parallel()

		got := uniqueSortedEntityTypes(
			[]uint16{
				coredata.FrameworkEntityType,
				coredata.OrganizationEntityType,
				coredata.FrameworkEntityType,
				coredata.OrganizationEntityType,
			},
		)
		assert.Equal(
			t,
			[]uint16{coredata.OrganizationEntityType, coredata.FrameworkEntityType},
			got,
		)
	})

	t.Run("resource type from action parses segments", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "ThirdParty", resourceTypeFromAction("core:third-party:get"))
		assert.Equal(t, "WebhookSubscription", resourceTypeFromAction("core:webhook-subscription:delete"))
		assert.Equal(t, "Unknown", resourceTypeFromAction("invalid-action"))
	})
}

func TestAuthorizer_LoadMembership(t *testing.T) {
	t.Parallel()

	a := &Authorizer{}

	t.Run("empty organization id returns nil membership", func(t *testing.T) {
		t.Parallel()

		membership, err := a.loadMembership(
			context.Background(),
			nil,
			gid.New(gid.NilTenant, coredata.IdentityEntityType),
			"",
		)
		require.NoError(t, err)
		assert.Nil(t, membership)
	})

	t.Run("invalid organization id returns parse error", func(t *testing.T) {
		t.Parallel()

		_, err := a.loadMembership(
			context.Background(),
			nil,
			gid.New(gid.NilTenant, coredata.IdentityEntityType),
			"not-a-gid",
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot parse gid")
	})
}

func TestAuthorizer_BuildAuditLogEntry(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	principalID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	resourceID := gid.New(tenantID, coredata.FrameworkEntityType)
	sessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)

	tests := []struct {
		name          string
		params        AuthorizeParams
		resourceAttrs policy.Attributes
		wantNil       bool
		wantActorType coredata.AuditLogActorType
	}{
		{
			name: "dry run returns nil",
			params: AuthorizeParams{
				Principal: principalID,
				Resource:  resourceID,
				Action:    "core:framework:get",
				DryRun:    true,
			},
			resourceAttrs: policy.Attributes{
				"organization_id": orgID.String(),
			},
			wantNil: true,
		},
		{
			name: "missing organization id returns nil",
			params: AuthorizeParams{
				Principal: principalID,
				Resource:  resourceID,
				Action:    "core:framework:get",
			},
			resourceAttrs: policy.Attributes{
				"id": resourceID.String(),
			},
			wantNil: true,
		},
		{
			name: "unparseable organization id returns nil",
			params: AuthorizeParams{
				Principal: principalID,
				Resource:  resourceID,
				Action:    "core:framework:get",
			},
			resourceAttrs: policy.Attributes{
				"organization_id": "invalid-gid",
			},
			wantNil: true,
		},
		{
			name: "session present sets user actor type",
			params: AuthorizeParams{
				Principal: principalID,
				Resource:  resourceID,
				Action:    "core:framework:get",
				Session:   &sessionID,
			},
			resourceAttrs: policy.Attributes{
				"organization_id": orgID.String(),
			},
			wantActorType: coredata.AuditLogActorTypeUser,
		},
		{
			name: "nil session sets api key actor type",
			params: AuthorizeParams{
				Principal: principalID,
				Resource:  resourceID,
				Action:    "core:framework:get",
			},
			resourceAttrs: policy.Attributes{
				"organization_id": orgID.String(),
			},
			wantActorType: coredata.AuditLogActorTypeAPIKey,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &Authorizer{
				logger: log.NewLogger(log.WithOutput(io.Discard)),
			}

			entry := a.buildAuditLogEntry(context.Background(), tt.params, tt.resourceAttrs)
			if tt.wantNil {
				assert.Nil(t, entry)

				return
			}

			require.NotNil(t, entry)
			assert.Equal(t, tt.wantActorType, entry.ActorType)
			assert.Equal(t, tt.params.Principal, entry.ActorID)
			assert.Equal(t, tt.params.Resource, entry.ResourceID)
			assert.Equal(t, tt.params.Action, entry.Action)
		})
	}
}

func TestAuthorizer_WrappedInternalErrors(t *testing.T) {
	t.Parallel()

	a := &Authorizer{
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}

	queryErr := errors.New("query failed")
	tx := &errorTx{queryErr: queryErr}
	principalID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	membershipID := gid.New(gid.NewTenantID(), coredata.MembershipEntityType)
	sessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)
	resourceOrgID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType).String()

	t.Run("load membership wraps load errors", func(t *testing.T) {
		t.Parallel()

		_, err := a.loadMembership(context.Background(), tx, principalID, resourceOrgID)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot load active membership")
	})

	t.Run("get active child session wraps load errors", func(t *testing.T) {
		t.Parallel()

		_, err := a.getActiveChildSessionForMembership(
			context.Background(),
			tx,
			sessionID,
			membershipID,
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot load child session")
	})

	t.Run("check assumption wraps non-assumption errors", func(t *testing.T) {
		t.Parallel()

		err := a.checkAssumption(
			context.Background(),
			tx,
			principalID,
			&sessionID,
			&coredata.Membership{ID: membershipID},
			false,
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot get active child session for membership")

		_, ok := errors.AsType[*ErrAssumptionRequired](err)
		assert.False(t, ok)
	})

	t.Run("build principal attributes wraps load errors", func(t *testing.T) {
		t.Parallel()

		_, err := a.buildPrincipalAttributes(context.Background(), tx, principalID, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot load principal attributes")
	})

	t.Run("load resource attributes by type wraps load errors", func(t *testing.T) {
		t.Parallel()

		resourceID := gid.New(gid.NewTenantID(), coredata.FrameworkEntityType)

		_, err := a.loadResourceAttributesByType(context.Background(), tx, []gid.GID{resourceID})
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot load batched resource attributes")
	})
}

type errorRow struct {
	err error
}

func (r *errorRow) Scan(...any) error {
	return r.err
}

type errorTx struct {
	queryErr error
}

var _ pg.Tx = (*errorTx)(nil)

func (tx *errorTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, tx.queryErr
}

func (tx *errorTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, tx.queryErr
}

func (tx *errorTx) QueryRow(context.Context, string, ...any) pgx.Row {
	return &errorRow{err: tx.queryErr}
}

func (tx *errorTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, tx.queryErr
}

func (tx *errorTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (tx *errorTx) Savepoint(context.Context, pg.ExecFunc[pg.Tx]) error {
	return tx.queryErr
}
