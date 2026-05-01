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

package coredata

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

// TestCloudAccount_CursorKey covers each supported order-by field.
// The cursor-key value drives keyset pagination; a regression here
// silently breaks list pagination with no compile-time signal.
func TestCloudAccount_CursorKey(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	id := gid.New(tenantID, CloudAccountEntityType)
	createdAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	account := &CloudAccount{
		ID:        id,
		CreatedAt: createdAt,
		Status:    CloudAccountStatusVerified,
		Provider:  CloudAccountProviderAWS,
	}

	t.Run("created_at", func(t *testing.T) {
		t.Parallel()

		key := account.CursorKey(CloudAccountOrderFieldCreatedAt)
		assert.NotEmpty(t, key)
	})

	t.Run("status", func(t *testing.T) {
		t.Parallel()

		key := account.CursorKey(CloudAccountOrderFieldStatus)
		assert.NotEmpty(t, key)
	})

	t.Run("provider", func(t *testing.T) {
		t.Parallel()

		key := account.CursorKey(CloudAccountOrderFieldProvider)
		assert.NotEmpty(t, key)
	})
}

// TestCloudAccount_CursorKey_Panics asserts that asking for an
// unknown order field panics at runtime. This is the agreed-upon
// fail-loud contract: a new order field MUST be added to the
// switch in CursorKey at the same time it is defined.
func TestCloudAccount_CursorKey_Panics(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	account := &CloudAccount{
		ID:        gid.New(tenantID, CloudAccountEntityType),
		CreatedAt: time.Now(),
	}

	assert.Panics(t, func() {
		_ = account.CursorKey(CloudAccountOrderField("BOGUS"))
	})
}

// TestCloudAccountOrderField_Column asserts the Column() helper
// returns the field's underlying string. The pagination SQL fragment
// in pkg/page substitutes Column() into the ORDER BY clause -- a
// drift between this and the actual DB column name silently breaks
// pagination ordering.
func TestCloudAccountOrderField_Column(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field    CloudAccountOrderField
		wantText string
	}{
		{field: CloudAccountOrderFieldCreatedAt, wantText: "CREATED_AT"},
		{field: CloudAccountOrderFieldStatus, wantText: "STATUS"},
		{field: CloudAccountOrderFieldProvider, wantText: "PROVIDER"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.wantText, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantText, tt.field.Column())
			assert.Equal(t, tt.wantText, tt.field.String())

			marshalled, err := tt.field.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, string(marshalled))

			var roundtrip CloudAccountOrderField
			require.NoError(t, roundtrip.UnmarshalText([]byte(tt.wantText)))
			assert.Equal(t, tt.field, roundtrip)
		})
	}
}

// TestCloudAccountFilter_NilFragment asserts the nil-receiver path
// returns "TRUE" so the filter is a no-op when no list endpoint
// passes in a filter object. The SQL must remain a static literal
// (no dynamic clause assembly) -- this is an approval blocker.
func TestCloudAccountFilter_NilFragment(t *testing.T) {
	t.Parallel()

	var f *CloudAccountFilter
	assert.Equal(t, "TRUE", f.SQLFragment())
	assert.Equal(t, 0, len(f.SQLArguments()))
}

// TestCloudAccountFilter_StaticSQLFragment asserts the SQL fragment
// produced is independent of which Filter.WithX setters were called.
// The fragment is hard-coded with @filter_* placeholders evaluated by
// PG; this test pins that contract so a future refactor cannot
// silently switch to dynamic SQL assembly.
func TestCloudAccountFilter_StaticSQLFragment(t *testing.T) {
	t.Parallel()

	bare := NewCloudAccountFilter().SQLFragment()
	withProvider := NewCloudAccountFilter().WithProvider(CloudAccountProviderAWS).SQLFragment()
	withStatus := NewCloudAccountFilter().WithStatus(CloudAccountStatusVerified).SQLFragment()
	withScope := NewCloudAccountFilter().WithScopeKind(CloudAccountScopeKindAWSAccount).SQLFragment()
	all := NewCloudAccountFilter().
		WithProvider(CloudAccountProviderGCP).
		WithStatus(CloudAccountStatusErrored).
		WithScopeKind(CloudAccountScopeKindGCPProject).
		SQLFragment()

	assert.Equal(t, bare, withProvider, "fragment must not depend on which fields are set")
	assert.Equal(t, bare, withStatus)
	assert.Equal(t, bare, withScope)
	assert.Equal(t, bare, all)

	// Sanity-check the fragment references each of the three
	// expected named arguments -- a typo in the column-name binding
	// would cause silent filter no-ops in production.
	for _, expectedArg := range []string{
		"@filter_cloud_account_provider",
		"@filter_cloud_account_status",
		"@filter_cloud_account_scope_kind",
	} {
		assert.True(
			t,
			strings.Contains(bare, expectedArg),
			"fragment must reference %s",
			expectedArg,
		)
	}
}

// TestCloudAccountFilter_SQLArguments asserts the arguments map
// always carries all three filter slots. Setters write the string
// form; unset slots stay nil so PG's CASE WHEN ... IS NULL THEN TRUE
// branch fires.
func TestCloudAccountFilter_SQLArguments(t *testing.T) {
	t.Parallel()

	t.Run("nil filter", func(t *testing.T) {
		t.Parallel()

		var f *CloudAccountFilter
		assert.Empty(t, f.SQLArguments())
	})

	t.Run("bare filter has all-nil slots", func(t *testing.T) {
		t.Parallel()

		args := NewCloudAccountFilter().SQLArguments()
		assert.Nil(t, args["filter_cloud_account_provider"])
		assert.Nil(t, args["filter_cloud_account_status"])
		assert.Nil(t, args["filter_cloud_account_scope_kind"])
	})

	t.Run("provider set populates only provider", func(t *testing.T) {
		t.Parallel()

		args := NewCloudAccountFilter().WithProvider(CloudAccountProviderAWS).SQLArguments()
		assert.Equal(t, "AWS", args["filter_cloud_account_provider"])
		assert.Nil(t, args["filter_cloud_account_status"])
		assert.Nil(t, args["filter_cloud_account_scope_kind"])
	})

	t.Run("status set populates only status", func(t *testing.T) {
		t.Parallel()

		args := NewCloudAccountFilter().WithStatus(CloudAccountStatusVerified).SQLArguments()
		assert.Nil(t, args["filter_cloud_account_provider"])
		assert.Equal(t, "VERIFIED", args["filter_cloud_account_status"])
		assert.Nil(t, args["filter_cloud_account_scope_kind"])
	})

	t.Run("scope kind set populates only scope kind", func(t *testing.T) {
		t.Parallel()

		args := NewCloudAccountFilter().WithScopeKind(CloudAccountScopeKindGCPProject).SQLArguments()
		assert.Nil(t, args["filter_cloud_account_provider"])
		assert.Nil(t, args["filter_cloud_account_status"])
		assert.Equal(t, "GCP_PROJECT", args["filter_cloud_account_scope_kind"])
	})

	t.Run("all set", func(t *testing.T) {
		t.Parallel()

		args := NewCloudAccountFilter().
			WithProvider(CloudAccountProviderAzure).
			WithStatus(CloudAccountStatusErrored).
			WithScopeKind(CloudAccountScopeKindAzureSubscription).
			SQLArguments()
		assert.Equal(t, "AZURE", args["filter_cloud_account_provider"])
		assert.Equal(t, "ERRORED", args["filter_cloud_account_status"])
		assert.Equal(t, "AZURE_SUBSCRIPTION", args["filter_cloud_account_scope_kind"])
	})
}

// TestCloudAccountProviderEnum pins the on-the-wire string values.
// The DB enum, the GraphQL enum, the MCP spec, and the CLI all key
// off these strings -- a rename here cascades into a five-place
// migration.
func TestCloudAccountProviderEnum(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "AWS", CloudAccountProviderAWS.String())
	assert.Equal(t, "GCP", CloudAccountProviderGCP.String())
	assert.Equal(t, "AZURE", CloudAccountProviderAzure.String())

	all := CloudAccountProviders()
	assert.Equal(
		t,
		[]CloudAccountProvider{
			CloudAccountProviderAWS,
			CloudAccountProviderGCP,
			CloudAccountProviderAzure,
		},
		all,
	)
}

// TestCloudAccountStatusEnum pins the on-the-wire status values
// driving the disconnected/recovery state machine.
func TestCloudAccountStatusEnum(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "PENDING_VERIFICATION", CloudAccountStatusPendingVerification.String())
	assert.Equal(t, "VERIFIED", CloudAccountStatusVerified.String())
	assert.Equal(t, "ERRORED", CloudAccountStatusErrored.String())
	assert.Equal(t, "DISCONNECTED", CloudAccountStatusDisconnected.String())
}

// TestCloudAccountStatus_Scan asserts the database scanner accepts
// both string and []byte representations and rejects unknown values
// with an error rather than zero-valuing the receiver.
func TestCloudAccountStatus_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    CloudAccountStatus
		wantErr bool
	}{
		{name: "verified string", input: "VERIFIED", want: CloudAccountStatusVerified},
		{name: "errored bytes", input: []byte("ERRORED"), want: CloudAccountStatusErrored},
		{name: "pending string", input: "PENDING_VERIFICATION", want: CloudAccountStatusPendingVerification},
		{name: "disconnected string", input: "DISCONNECTED", want: CloudAccountStatusDisconnected},
		{name: "invalid string", input: "BOGUS", wantErr: true},
		{name: "wrong type", input: 42, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s CloudAccountStatus
			err := s.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, s)
		})
	}
}

// TestCloudAccountProvider_Scan covers the same Scan contract for
// the provider enum.
func TestCloudAccountProvider_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    CloudAccountProvider
		wantErr bool
	}{
		{name: "aws", input: "AWS", want: CloudAccountProviderAWS},
		{name: "gcp", input: "GCP", want: CloudAccountProviderGCP},
		{name: "azure", input: "AZURE", want: CloudAccountProviderAzure},
		{name: "azure bytes", input: []byte("AZURE"), want: CloudAccountProviderAzure},
		{name: "invalid", input: "ORACLE", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var p CloudAccountProvider
			err := p.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, p)
		})
	}
}

// TestCloudAccountScopeKind_Scan covers the same Scan contract for
// the scope-kind enum, including the per-cloud variants used to
// drive the per-(provider, scope) install / verify dispatch.
func TestCloudAccountScopeKind_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    CloudAccountScopeKind
		wantErr bool
	}{
		{name: "aws account", input: "AWS_ACCOUNT", want: CloudAccountScopeKindAWSAccount},
		{name: "aws organization", input: "AWS_ORGANIZATION", want: CloudAccountScopeKindAWSOrganization},
		{name: "gcp project", input: "GCP_PROJECT", want: CloudAccountScopeKindGCPProject},
		{name: "gcp organization", input: "GCP_ORGANIZATION", want: CloudAccountScopeKindGCPOrganization},
		{name: "azure subscription", input: "AZURE_SUBSCRIPTION", want: CloudAccountScopeKindAzureSubscription},
		{name: "azure mg", input: "AZURE_MANAGEMENT_GROUP", want: CloudAccountScopeKindAzureManagementGroup},
		{name: "azure tenant", input: "AZURE_TENANT", want: CloudAccountScopeKindAzureTenant},
		{name: "invalid", input: "GCP_FOLDER", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s CloudAccountScopeKind
			err := s.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, s)
		})
	}
}

// TestCloudAccountCredentialKind_Scan covers the same Scan contract
// for the credential-kind enum, including v2 placeholders that ship
// in the enum but are not yet wired through Validate.
func TestCloudAccountCredentialKind_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    CloudAccountCredentialKind
		wantErr bool
	}{
		{name: "aws assume role", input: "AWS_ASSUME_ROLE", want: CloudAccountCredentialKindAWSAssumeRole},
		{name: "gcp service account key", input: "GCP_SERVICE_ACCOUNT_KEY", want: CloudAccountCredentialKindGCPServiceAccountKey},
		{name: "azure client secret", input: "AZURE_CLIENT_SECRET", want: CloudAccountCredentialKindAzureClientSecret},
		{name: "gcp wif placeholder", input: "GCP_WORKLOAD_IDENTITY_FEDERATION", want: CloudAccountCredentialKindGCPWorkloadIdentityFederation},
		{name: "azure federated placeholder", input: "AZURE_FEDERATED_CREDENTIAL", want: CloudAccountCredentialKindAzureFederatedCredential},
		{name: "invalid", input: "PASSWORD", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var k CloudAccountCredentialKind
			err := k.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, k)
		})
	}
}
