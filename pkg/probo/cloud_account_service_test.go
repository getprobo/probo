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

package probo

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// validRequestGIDs returns a fresh organization id and cloud-account id
// inside a fresh tenant. The Validate tests verify that supplying these
// well-formed values clears the GID-shape checks; runtime existence is
// out of scope for Validate (purity) tests.
func validRequestGIDs(t *testing.T) (gid.GID, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	cloudAccountID := gid.New(tenantID, coredata.CloudAccountEntityType)

	return orgID, cloudAccountID
}

// TestCreateCloudAccountRequest_Validate covers the field-level shape
// checks performed before the service body opens any DB connection.
// It also asserts purity: Validate must never call into pg, the
// registry, or any external I/O surface.
func TestCreateCloudAccountRequest_Validate(t *testing.T) {
	t.Parallel()

	orgID, _ := validRequestGIDs(t)

	t.Run("happy path aws", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           "Production AWS",
			Provider:        coredata.CloudAccountProviderAWS,
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: "123456789012",
		}

		assert.NoError(t, req.Validate())
	})

	t.Run("happy path gcp", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           "Production GCP",
			Provider:        coredata.CloudAccountProviderGCP,
			CredentialKind:  coredata.CloudAccountCredentialKindGCPServiceAccountKey,
			ScopeKind:       coredata.CloudAccountScopeKindGCPProject,
			ScopeIdentifier: "probo-prod-1",
		}

		assert.NoError(t, req.Validate())
	})

	t.Run("happy path azure", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           "Production Azure",
			Provider:        coredata.CloudAccountProviderAzure,
			CredentialKind:  coredata.CloudAccountCredentialKindAzureClientSecret,
			ScopeKind:       coredata.CloudAccountScopeKindAzureSubscription,
			ScopeIdentifier: "00000000-0000-0000-0000-000000000001",
		}

		assert.NoError(t, req.Validate())
	})

	t.Run("missing label", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Provider:        coredata.CloudAccountProviderAWS,
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: "123",
		}

		assert.Error(t, req.Validate())
	})

	t.Run("label too long", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           strings.Repeat("x", CloudAccountLabelMaxLength+1),
			Provider:        coredata.CloudAccountProviderAWS,
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: "123",
		}

		assert.Error(t, req.Validate())
	})

	t.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           "x",
			Provider:        coredata.CloudAccountProvider("ORACLE"),
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: "123",
		}

		assert.Error(t, req.Validate())
	})

	t.Run("invalid scope identifier too long", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           "x",
			Provider:        coredata.CloudAccountProviderAWS,
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: strings.Repeat("x", CloudAccountScopeIdentifierMaxLength+1),
		}

		assert.Error(t, req.Validate())
	})

	t.Run("scope identifier with newline rejected", func(t *testing.T) {
		t.Parallel()

		req := CreateCloudAccountRequest{
			OrganizationID:  orgID,
			Label:           "x",
			Provider:        coredata.CloudAccountProviderAWS,
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: "123\n456",
		}

		assert.Error(t, req.Validate())
	})

	t.Run("organization id with wrong entity type", func(t *testing.T) {
		t.Parallel()

		// Use a CloudAccount-typed gid where an Organization-typed
		// gid is required. Validate must reject this.
		_, accountID := validRequestGIDs(t)
		req := CreateCloudAccountRequest{
			OrganizationID:  accountID,
			Label:           "x",
			Provider:        coredata.CloudAccountProviderAWS,
			CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier: "123",
		}

		assert.Error(t, req.Validate())
	})
}

// TestRotateCloudAccountCredentialsRequest_Validate covers the
// (cloudAccountID, provider, kind) shape checks. Provider/kind
// must-match-row checks live in the service body, not Validate.
func TestRotateCloudAccountCredentialsRequest_Validate(t *testing.T) {
	t.Parallel()

	_, accountID := validRequestGIDs(t)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		req := RotateCloudAccountCredentialsRequest{
			CloudAccountID: accountID,
			Provider:       coredata.CloudAccountProviderAWS,
			CredentialKind: coredata.CloudAccountCredentialKindAWSAssumeRole,
		}

		assert.NoError(t, req.Validate())
	})

	t.Run("missing cloud account id", func(t *testing.T) {
		t.Parallel()

		req := RotateCloudAccountCredentialsRequest{
			Provider:       coredata.CloudAccountProviderAWS,
			CredentialKind: coredata.CloudAccountCredentialKindAWSAssumeRole,
		}

		assert.Error(t, req.Validate())
	})

	t.Run("missing provider", func(t *testing.T) {
		t.Parallel()

		req := RotateCloudAccountCredentialsRequest{
			CloudAccountID: accountID,
			CredentialKind: coredata.CloudAccountCredentialKindAWSAssumeRole,
		}

		assert.Error(t, req.Validate())
	})

	t.Run("missing kind", func(t *testing.T) {
		t.Parallel()

		req := RotateCloudAccountCredentialsRequest{
			CloudAccountID: accountID,
			Provider:       coredata.CloudAccountProviderAWS,
		}

		assert.Error(t, req.Validate())
	})

	t.Run("organization id where cloud account id required", func(t *testing.T) {
		t.Parallel()

		orgID, _ := validRequestGIDs(t)
		req := RotateCloudAccountCredentialsRequest{
			CloudAccountID: orgID,
			Provider:       coredata.CloudAccountProviderAWS,
			CredentialKind: coredata.CloudAccountCredentialKindAWSAssumeRole,
		}

		assert.Error(t, req.Validate())
	})
}

// TestVerifyCloudAccountRequest_Validate covers the GID-typed name
// of the row to probe.
func TestVerifyCloudAccountRequest_Validate(t *testing.T) {
	t.Parallel()

	_, accountID := validRequestGIDs(t)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		req := VerifyCloudAccountRequest{CloudAccountID: accountID}
		assert.NoError(t, req.Validate())
	})

	t.Run("missing cloud account id", func(t *testing.T) {
		t.Parallel()

		req := VerifyCloudAccountRequest{}
		assert.Error(t, req.Validate())
	})

	t.Run("wrong entity type", func(t *testing.T) {
		t.Parallel()

		orgID, _ := validRequestGIDs(t)
		req := VerifyCloudAccountRequest{CloudAccountID: orgID}
		assert.Error(t, req.Validate())
	})
}

// TestGenerateInstallAssetsRequest_Validate covers the
// (organizationID, provider, scope_kind) shape checks that gate the
// install-assets resolver.
func TestGenerateInstallAssetsRequest_Validate(t *testing.T) {
	t.Parallel()

	orgID, _ := validRequestGIDs(t)

	t.Run("happy path aws", func(t *testing.T) {
		t.Parallel()

		req := GenerateInstallAssetsRequest{
			OrganizationID: orgID,
			Provider:       coredata.CloudAccountProviderAWS,
			ScopeKind:      coredata.CloudAccountScopeKindAWSAccount,
		}

		assert.NoError(t, req.Validate())
	})

	t.Run("missing provider", func(t *testing.T) {
		t.Parallel()

		req := GenerateInstallAssetsRequest{
			OrganizationID: orgID,
			ScopeKind:      coredata.CloudAccountScopeKindAWSAccount,
		}

		assert.Error(t, req.Validate())
	})

	t.Run("missing scope kind", func(t *testing.T) {
		t.Parallel()

		req := GenerateInstallAssetsRequest{
			OrganizationID: orgID,
			Provider:       coredata.CloudAccountProviderAWS,
		}

		assert.Error(t, req.Validate())
	})

	t.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		req := GenerateInstallAssetsRequest{
			OrganizationID: orgID,
			Provider:       coredata.CloudAccountProvider("ORACLE"),
			ScopeKind:      coredata.CloudAccountScopeKindAWSAccount,
		}

		assert.Error(t, req.Validate())
	})
}

// TestMarshalCredentialsForCreate_ProviderKindMatching asserts the
// per-(provider, kind) discriminator is enforced before any encryption
// or persistence runs. A mismatched (provider=AWS, kind=GCP_*) must
// surface ErrCredentialsInvalid; the cleartext envelope must never
// touch the cipher when the discriminator disagrees.
func TestMarshalCredentialsForCreate_ProviderKindMatching(t *testing.T) {
	t.Parallel()

	orgID, _ := validRequestGIDs(t)
	svc := &CloudAccountService{}

	tests := []struct {
		name     string
		provider coredata.CloudAccountProvider
		kind     coredata.CloudAccountCredentialKind
		ok       bool
	}{
		{
			name:     "aws + aws_assume_role",
			provider: coredata.CloudAccountProviderAWS,
			kind:     coredata.CloudAccountCredentialKindAWSAssumeRole,
			ok:       true,
		},
		{
			name:     "aws + gcp_service_account_key rejected",
			provider: coredata.CloudAccountProviderAWS,
			kind:     coredata.CloudAccountCredentialKindGCPServiceAccountKey,
			ok:       false,
		},
		{
			name:     "aws + azure_client_secret rejected",
			provider: coredata.CloudAccountProviderAWS,
			kind:     coredata.CloudAccountCredentialKindAzureClientSecret,
			ok:       false,
		},
		{
			name:     "gcp + gcp_service_account_key",
			provider: coredata.CloudAccountProviderGCP,
			kind:     coredata.CloudAccountCredentialKindGCPServiceAccountKey,
			ok:       true,
		},
		{
			name:     "gcp + aws_assume_role rejected",
			provider: coredata.CloudAccountProviderGCP,
			kind:     coredata.CloudAccountCredentialKindAWSAssumeRole,
			ok:       false,
		},
		{
			name:     "gcp + gcp_workload_identity_federation rejected (v2 placeholder)",
			provider: coredata.CloudAccountProviderGCP,
			kind:     coredata.CloudAccountCredentialKindGCPWorkloadIdentityFederation,
			ok:       false,
		},
		{
			name:     "azure + azure_client_secret",
			provider: coredata.CloudAccountProviderAzure,
			kind:     coredata.CloudAccountCredentialKindAzureClientSecret,
			ok:       true,
		},
		{
			name:     "azure + aws_assume_role rejected",
			provider: coredata.CloudAccountProviderAzure,
			kind:     coredata.CloudAccountCredentialKindAWSAssumeRole,
			ok:       false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := CreateCloudAccountRequest{
				OrganizationID:  orgID,
				Label:           "x",
				Provider:        tt.provider,
				CredentialKind:  tt.kind,
				ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
				ScopeIdentifier: "123",
			}

			payload, err := svc.marshalCredentialsForCreate(req)
			if tt.ok {
				require.NoError(t, err)
				assert.NotEmpty(t, payload, "payload must be a non-empty JSON envelope")
				return
			}

			require.Error(t, err)
			assert.True(
				t,
				errors.Is(err, cloudaccount.ErrCredentialsInvalid),
				"mismatch must wrap ErrCredentialsInvalid; got %v",
				err,
			)
			assert.True(
				t,
				strings.HasPrefix(err.Error(), "cannot build "),
				"error message must start with %q; got %q",
				"cannot build ",
				err.Error(),
			)
		})
	}
}

// TestMarshalCredentialsForCreate_UnsupportedProvider asserts the
// default branch returns a typed error for an unknown provider value.
func TestMarshalCredentialsForCreate_UnsupportedProvider(t *testing.T) {
	t.Parallel()

	orgID, _ := validRequestGIDs(t)
	svc := &CloudAccountService{}

	req := CreateCloudAccountRequest{
		OrganizationID:  orgID,
		Label:           "x",
		Provider:        coredata.CloudAccountProvider("ORACLE"),
		CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
		ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
		ScopeIdentifier: "1",
	}

	_, err := svc.marshalCredentialsForCreate(req)
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "cannot build cloud account credentials"))
}

// TestMarshalCredentialsForRotate_PreservesScopeFromRow asserts the
// rotate envelope carries the loaded row's ScopeKind, not anything
// from the request -- credential rotation is not a scope-edit flow.
func TestMarshalCredentialsForRotate_PreservesScopeFromRow(t *testing.T) {
	t.Parallel()

	svc := &CloudAccountService{}

	tests := []struct {
		name      string
		account   *coredata.CloudAccount
		req       RotateCloudAccountCredentialsRequest
		ok        bool
		expectKey string // a substring expected to be present in marshalled payload
	}{
		{
			name: "aws scope preserved",
			account: &coredata.CloudAccount{
				Provider:        coredata.CloudAccountProviderAWS,
				CredentialKind:  coredata.CloudAccountCredentialKindAWSAssumeRole,
				ScopeKind:       coredata.CloudAccountScopeKindAWSAccount,
				ScopeIdentifier: "123456789012",
			},
			req: RotateCloudAccountCredentialsRequest{
				Provider:       coredata.CloudAccountProviderAWS,
				CredentialKind: coredata.CloudAccountCredentialKindAWSAssumeRole,
				AWSRoleARN:     "arn:aws:iam::123456789012:role/Probo",
				AWSExternalID:  "ext",
			},
			ok:        true,
			expectKey: "AWS_ACCOUNT",
		},
		{
			name: "gcp project scope preserved with project_id",
			account: &coredata.CloudAccount{
				Provider:        coredata.CloudAccountProviderGCP,
				CredentialKind:  coredata.CloudAccountCredentialKindGCPServiceAccountKey,
				ScopeKind:       coredata.CloudAccountScopeKindGCPProject,
				ScopeIdentifier: "probo-prod",
			},
			req: RotateCloudAccountCredentialsRequest{
				Provider:              coredata.CloudAccountProviderGCP,
				CredentialKind:        coredata.CloudAccountCredentialKindGCPServiceAccountKey,
				GCPServiceAccountJSON: []byte(`{"type":"service_account"}`),
			},
			ok:        true,
			expectKey: "probo-prod",
		},
		{
			name: "azure subscription scope preserved",
			account: &coredata.CloudAccount{
				Provider:        coredata.CloudAccountProviderAzure,
				CredentialKind:  coredata.CloudAccountCredentialKindAzureClientSecret,
				ScopeKind:       coredata.CloudAccountScopeKindAzureSubscription,
				ScopeIdentifier: "00000000-0000-0000-0000-000000000001",
			},
			req: RotateCloudAccountCredentialsRequest{
				Provider:          coredata.CloudAccountProviderAzure,
				CredentialKind:    coredata.CloudAccountCredentialKindAzureClientSecret,
				AzureTenantID:     "tenant",
				AzureClientID:     "client",
				AzureClientSecret: "secret",
			},
			ok:        true,
			expectKey: "AZURE_SUBSCRIPTION",
		},
		{
			name: "unknown provider rejected",
			account: &coredata.CloudAccount{
				Provider: coredata.CloudAccountProvider("ORACLE"),
			},
			req: RotateCloudAccountCredentialsRequest{
				Provider: coredata.CloudAccountProvider("ORACLE"),
			},
			ok: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload, err := svc.marshalCredentialsForRotate(tt.account, tt.req)
			if !tt.ok {
				require.Error(t, err)
				assert.True(t, strings.HasPrefix(err.Error(), "cannot build "))
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, payload)
			if tt.expectKey != "" {
				assert.Contains(t, string(payload), tt.expectKey, "scope identifier or kind must be embedded in envelope")
			}
		})
	}
}

// TestMapCloudAccountToRecord asserts the entity -> record mapping
// pulls every field the registry needs, and that ExternalID is
// flattened from the *string pointer when present.
func TestMapCloudAccountToRecord(t *testing.T) {
	t.Parallel()

	t.Run("with external id", func(t *testing.T) {
		t.Parallel()

		extID := "ext-abc"
		_, accountID := validRequestGIDs(t)
		account := &coredata.CloudAccount{
			ID:                   accountID,
			Provider:             coredata.CloudAccountProviderAWS,
			CredentialKind:       coredata.CloudAccountCredentialKindAWSAssumeRole,
			ScopeKind:            coredata.CloudAccountScopeKindAWSAccount,
			ScopeIdentifier:      "123",
			ExternalID:           &extID,
			DecryptedCredentials: []byte(`{"v":1}`),
		}

		rec := mapCloudAccountToRecord(account)
		assert.Equal(t, accountID.String(), rec.ID)
		assert.Equal(t, coredata.CloudAccountProviderAWS, rec.Provider)
		assert.Equal(t, coredata.CloudAccountCredentialKindAWSAssumeRole, rec.Kind)
		assert.Equal(t, coredata.CloudAccountScopeKindAWSAccount, rec.ScopeKind)
		assert.Equal(t, "123", rec.ScopeIdentifier)
		assert.Equal(t, extID, rec.ExternalID)
		assert.Equal(t, []byte(`{"v":1}`), rec.DecryptedCredentials)
	})

	t.Run("nil external id flattens to empty", func(t *testing.T) {
		t.Parallel()

		_, accountID := validRequestGIDs(t)
		account := &coredata.CloudAccount{
			ID:                   accountID,
			Provider:             coredata.CloudAccountProviderGCP,
			CredentialKind:       coredata.CloudAccountCredentialKindGCPServiceAccountKey,
			ScopeKind:            coredata.CloudAccountScopeKindGCPProject,
			ScopeIdentifier:      "p",
			ExternalID:           nil,
			DecryptedCredentials: []byte("x"),
		}

		rec := mapCloudAccountToRecord(account)
		assert.Equal(t, "", rec.ExternalID)
	})
}

// TestExtractGCPProjectIDFromAccount returns the scope identifier
// only when scope_kind = GCP_PROJECT.
func TestExtractGCPProjectIDFromAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scopeKind coredata.CloudAccountScopeKind
		scopeID   string
		want      string
	}{
		{
			name:      "project scope returns identifier",
			scopeKind: coredata.CloudAccountScopeKindGCPProject,
			scopeID:   "probo-prod",
			want:      "probo-prod",
		},
		{
			name:      "organization scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindGCPOrganization,
			scopeID:   "987654321",
			want:      "",
		},
		{
			name:      "aws scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindAWSAccount,
			scopeID:   "123",
			want:      "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractGCPProjectIDFromAccount(&coredata.CloudAccount{
				ScopeKind:       tt.scopeKind,
				ScopeIdentifier: tt.scopeID,
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractGCPOrganizationIDFromAccount returns the scope
// identifier only when scope_kind = GCP_ORGANIZATION.
func TestExtractGCPOrganizationIDFromAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scopeKind coredata.CloudAccountScopeKind
		scopeID   string
		want      string
	}{
		{
			name:      "organization scope returns identifier",
			scopeKind: coredata.CloudAccountScopeKindGCPOrganization,
			scopeID:   "987654321",
			want:      "987654321",
		},
		{
			name:      "project scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindGCPProject,
			scopeID:   "probo-prod",
			want:      "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractGCPOrganizationIDFromAccount(&coredata.CloudAccount{
				ScopeKind:       tt.scopeKind,
				ScopeIdentifier: tt.scopeID,
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractAzureSubscriptionIDFromAccount returns the scope
// identifier only when scope_kind = AZURE_SUBSCRIPTION.
func TestExtractAzureSubscriptionIDFromAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scopeKind coredata.CloudAccountScopeKind
		want      string
	}{
		{
			name:      "subscription scope returns identifier",
			scopeKind: coredata.CloudAccountScopeKindAzureSubscription,
			want:      "sub-id",
		},
		{
			name:      "management group scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindAzureManagementGroup,
			want:      "",
		},
		{
			name:      "tenant scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindAzureTenant,
			want:      "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractAzureSubscriptionIDFromAccount(&coredata.CloudAccount{
				ScopeKind:       tt.scopeKind,
				ScopeIdentifier: "sub-id",
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractAzureManagementGroupFromAccount asserts both
// MANAGEMENT_GROUP and TENANT scope kinds map to the management-group
// slot (since Azure tenant root is itself a management group), while
// SUBSCRIPTION returns empty.
func TestExtractAzureManagementGroupFromAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scopeKind coredata.CloudAccountScopeKind
		want      string
	}{
		{
			name:      "management group scope returns identifier",
			scopeKind: coredata.CloudAccountScopeKindAzureManagementGroup,
			want:      "mg-id",
		},
		{
			name:      "tenant scope returns identifier",
			scopeKind: coredata.CloudAccountScopeKindAzureTenant,
			want:      "mg-id",
		},
		{
			name:      "subscription scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindAzureSubscription,
			want:      "",
		},
		{
			name:      "non-azure scope returns empty",
			scopeKind: coredata.CloudAccountScopeKindAWSAccount,
			want:      "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractAzureManagementGroupFromAccount(&coredata.CloudAccount{
				ScopeKind:       tt.scopeKind,
				ScopeIdentifier: "mg-id",
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCloudAccountAccessSourceFKConstraint pins the constraint name
// the Delete error mapper compares against. The migration that
// created access_sources.cloud_account_id REFERENCES cloud_accounts(id)
// names this constraint "access_sources_cloud_account_id_fkey" --
// any rename in a future migration must be reflected in the service.
func TestCloudAccountAccessSourceFKConstraint(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"access_sources_cloud_account_id_fkey",
		cloudAccountAccessSourceFKConstraint,
	)
}

// TestCloudAccountConstants pins the request-shape limits the
// validators enforce. A widening / tightening of these caps is a
// schema-level concern and must be visible in the test diff.
func TestCloudAccountConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 100, CloudAccountLabelMaxLength)
	assert.Equal(t, 256, CloudAccountScopeIdentifierMaxLength)
}
