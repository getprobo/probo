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

package cloudaccount_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
)

// awsRecordHelper builds a CloudAccountRecord whose decrypted credentials
// are a marshalled AWSCredentials envelope. Centralised so individual
// subtests can override fields without re-marshalling the envelope by
// hand.
func awsRecordHelper(t *testing.T) cloudaccount.CloudAccountRecord {
	t.Helper()

	creds := &cloudaccount.AWSCredentials{
		RoleARN:    "arn:aws:iam::123456789012:role/probo-cloud-account",
		ExternalID: "abcdef0123456789",
		ScopeKind:  coredata.CloudAccountScopeKindAWSAccount,
	}
	raw, err := json.Marshal(creds)
	require.NoError(t, err)

	return cloudaccount.CloudAccountRecord{
		ID:                   "cloud-account-aws",
		Provider:             coredata.CloudAccountProviderAWS,
		Kind:                 coredata.CloudAccountCredentialKindAWSAssumeRole,
		ScopeKind:            coredata.CloudAccountScopeKindAWSAccount,
		ScopeIdentifier:      "123456789012",
		ExternalID:           creds.ExternalID,
		DecryptedCredentials: raw,
	}
}

func gcpRecordHelper(t *testing.T) cloudaccount.CloudAccountRecord {
	t.Helper()

	creds := &cloudaccount.GCPCredentials{
		ServiceAccountJSON: []byte(`{"type":"service_account","project_id":"probo-test"}`),
		ScopeKind:          coredata.CloudAccountScopeKindGCPProject,
		ProjectID:          "probo-test",
	}
	raw, err := json.Marshal(creds)
	require.NoError(t, err)

	return cloudaccount.CloudAccountRecord{
		ID:                   "cloud-account-gcp",
		Provider:             coredata.CloudAccountProviderGCP,
		Kind:                 coredata.CloudAccountCredentialKindGCPServiceAccountKey,
		ScopeKind:            coredata.CloudAccountScopeKindGCPProject,
		ScopeIdentifier:      "probo-test",
		DecryptedCredentials: raw,
	}
}

func azureRecordHelper(t *testing.T) cloudaccount.CloudAccountRecord {
	t.Helper()

	creds := &cloudaccount.AzureCredentials{
		TenantID:       "00000000-0000-0000-0000-000000000000",
		ClientID:       "11111111-1111-1111-1111-111111111111",
		ClientSecret:   "very-secret",
		ScopeKind:      coredata.CloudAccountScopeKindAzureSubscription,
		SubscriptionID: "22222222-2222-2222-2222-222222222222",
	}
	raw, err := json.Marshal(creds)
	require.NoError(t, err)

	return cloudaccount.CloudAccountRecord{
		ID:                   "cloud-account-azure",
		Provider:             coredata.CloudAccountProviderAzure,
		Kind:                 coredata.CloudAccountCredentialKindAzureClientSecret,
		ScopeKind:            coredata.CloudAccountScopeKindAzureSubscription,
		ScopeIdentifier:      creds.SubscriptionID,
		DecryptedCredentials: raw,
	}
}

func newTestRegistry() *cloudaccount.Registry {
	return cloudaccount.NewRegistry(cloudaccount.Config{
		BaseAWSConfig: aws.Config{Region: "us-east-1"},
		HTTPClient:    &http.Client{},
	})
}

func TestRegistry_BuildAWSProvider(t *testing.T) {
	t.Parallel()

	t.Run("returns AWSProvider on matching record", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := awsRecordHelper(t)

		got, err := reg.BuildAWSProvider(rec)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec, got.Record())
	})

	t.Run("rejects non-AWS provider", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := awsRecordHelper(t)
		rec.Provider = coredata.CloudAccountProviderGCP

		got, err := reg.BuildAWSProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})

	t.Run("rejects mismatched envelope kind", func(t *testing.T) {
		t.Parallel()

		// Build a payload whose envelope advertises an Azure kind, but
		// stamped with provider=AWS on the record. UnmarshalCredentials
		// should reject this with ErrCredentialsInvalid.
		azureCreds := &cloudaccount.AzureCredentials{
			TenantID:     "00000000-0000-0000-0000-000000000000",
			ClientID:     "11111111-1111-1111-1111-111111111111",
			ClientSecret: "secret",
			ScopeKind:    coredata.CloudAccountScopeKindAzureSubscription,
		}
		raw, err := json.Marshal(azureCreds)
		require.NoError(t, err)

		rec := cloudaccount.CloudAccountRecord{
			ID:                   "cloud-account-aws",
			Provider:             coredata.CloudAccountProviderAWS,
			DecryptedCredentials: raw,
		}

		reg := newTestRegistry()
		got, err := reg.BuildAWSProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})

	t.Run("rejects malformed payload bytes", func(t *testing.T) {
		t.Parallel()

		rec := cloudaccount.CloudAccountRecord{
			ID:                   "cloud-account-aws",
			Provider:             coredata.CloudAccountProviderAWS,
			DecryptedCredentials: []byte("not json"),
		}

		reg := newTestRegistry()
		got, err := reg.BuildAWSProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})
}

func TestRegistry_BuildGCPProvider(t *testing.T) {
	t.Parallel()

	t.Run("returns GCPProvider on matching record", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := gcpRecordHelper(t)

		got, err := reg.BuildGCPProvider(rec)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec, got.Record())
	})

	t.Run("rejects non-GCP provider", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := gcpRecordHelper(t)
		rec.Provider = coredata.CloudAccountProviderAWS

		got, err := reg.BuildGCPProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})

	t.Run("rejects malformed payload bytes", func(t *testing.T) {
		t.Parallel()

		rec := cloudaccount.CloudAccountRecord{
			Provider:             coredata.CloudAccountProviderGCP,
			DecryptedCredentials: []byte("\x00\x01"),
		}

		reg := newTestRegistry()
		got, err := reg.BuildGCPProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})
}

func TestRegistry_BuildAzureProvider(t *testing.T) {
	t.Parallel()

	t.Run("returns AzureProvider on matching record", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := azureRecordHelper(t)

		got, err := reg.BuildAzureProvider(rec)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec, got.Record())
	})

	t.Run("rejects non-Azure provider", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := azureRecordHelper(t)
		rec.Provider = coredata.CloudAccountProviderAWS

		got, err := reg.BuildAzureProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})

	t.Run("rejects malformed payload bytes", func(t *testing.T) {
		t.Parallel()

		rec := cloudaccount.CloudAccountRecord{
			Provider:             coredata.CloudAccountProviderAzure,
			DecryptedCredentials: []byte("garbage"),
		}

		reg := newTestRegistry()
		got, err := reg.BuildAzureProvider(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})
}

func TestRegistry_BuildProbeable_PolymorphicDispatch(t *testing.T) {
	t.Parallel()

	t.Run("AWS dispatches to AWSProvider implementing Probeable", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := awsRecordHelper(t)

		got, err := reg.BuildProbeable(rec)
		require.NoError(t, err)
		require.NotNil(t, got)

		_, ok := got.(*cloudaccount.AWSProvider)
		assert.True(t, ok, "expected *AWSProvider, got %T", got)

		// Compile-time interface satisfaction is also checked here:
		// if Probeable were not satisfied this assignment would not
		// type-check.
		var _ cloudaccount.Probeable = got
	})

	t.Run("GCP dispatches to GCPProvider implementing Probeable", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := gcpRecordHelper(t)

		got, err := reg.BuildProbeable(rec)
		require.NoError(t, err)
		require.NotNil(t, got)

		_, ok := got.(*cloudaccount.GCPProvider)
		assert.True(t, ok, "expected *GCPProvider, got %T", got)

		var _ cloudaccount.Probeable = got
	})

	t.Run("Azure dispatches to AzureProvider implementing Probeable", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := azureRecordHelper(t)

		got, err := reg.BuildProbeable(rec)
		require.NoError(t, err)
		require.NotNil(t, got)

		_, ok := got.(*cloudaccount.AzureProvider)
		assert.True(t, ok, "expected *AzureProvider, got %T", got)

		var _ cloudaccount.Probeable = got
	})

	t.Run("unknown provider returns wrapped error", func(t *testing.T) {
		t.Parallel()

		reg := newTestRegistry()
		rec := cloudaccount.CloudAccountRecord{
			Provider: coredata.CloudAccountProvider("OBSCURE_PROVIDER"),
		}

		got, err := reg.BuildProbeable(rec)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "cannot build cloud account probeable:")
	})
}

// TestRegistry_CompileTimeInterfaceAssertions captures the var
// assertions in the typed provider files. If any of these stop being
// Probeable, the test binary won't compile -- which is the assertion
// itself.
func TestRegistry_CompileTimeInterfaceAssertions(t *testing.T) {
	t.Parallel()

	var (
		_ cloudaccount.Credentials = (*cloudaccount.AWSCredentials)(nil)
		_ cloudaccount.Credentials = (*cloudaccount.GCPCredentials)(nil)
		_ cloudaccount.Credentials = (*cloudaccount.AzureCredentials)(nil)
		_ cloudaccount.Probeable   = (*cloudaccount.AWSProvider)(nil)
		_ cloudaccount.Probeable   = (*cloudaccount.GCPProvider)(nil)
		_ cloudaccount.Probeable   = (*cloudaccount.AzureProvider)(nil)
	)

	// Sanity assertion so the test reports as run.
	assert.True(t, true)
}
