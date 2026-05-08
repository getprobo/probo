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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
)

func TestUnmarshalCredentials_ProviderKindMismatch(t *testing.T) {
	t.Parallel()

	azureEnvelope, err := json.Marshal(&cloudaccount.AzureCredentials{
		TenantID:     "00000000-0000-0000-0000-000000000000",
		ClientID:     "11111111-1111-1111-1111-111111111111",
		ClientSecret: "secret",
		ScopeKind:    coredata.CloudAccountScopeKindAzureSubscription,
	})
	require.NoError(t, err)

	t.Run("azure payload submitted under provider=AWS is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderAWS, azureEnvelope)
		require.Error(t, err)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})

	t.Run("azure payload submitted under provider=GCP is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderGCP, azureEnvelope)
		require.Error(t, err)
		assert.ErrorIs(t, err, cloudaccount.ErrCredentialsInvalid)
	})
}

func TestUnmarshalCredentials_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("aws", func(t *testing.T) {
		t.Parallel()

		original := &cloudaccount.AWSCredentials{
			RoleARN:    "arn:aws:iam::123456789012:role/probo-cloud-account",
			ExternalID: "abcdef",
			ScopeKind:  coredata.CloudAccountScopeKindAWSAccount,
		}

		raw, err := json.Marshal(original)
		require.NoError(t, err)

		decoded, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderAWS, raw)
		require.NoError(t, err)

		got, ok := decoded.(*cloudaccount.AWSCredentials)
		require.True(t, ok, "expected *AWSCredentials, got %T", decoded)
		assert.Equal(t, original.RoleARN, got.RoleARN)
		assert.Equal(t, original.ExternalID, got.ExternalID)
		assert.Equal(t, original.ScopeKind, got.ScopeKind)
	})

	t.Run("gcp", func(t *testing.T) {
		t.Parallel()

		original := &cloudaccount.GCPCredentials{
			ServiceAccountJSON: []byte(`{"type":"service_account"}`),
			ScopeKind:          coredata.CloudAccountScopeKindGCPProject,
			ProjectID:          "probo-test-project",
		}

		raw, err := json.Marshal(original)
		require.NoError(t, err)

		decoded, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderGCP, raw)
		require.NoError(t, err)

		got, ok := decoded.(*cloudaccount.GCPCredentials)
		require.True(t, ok, "expected *GCPCredentials, got %T", decoded)
		assert.Equal(t, original.ServiceAccountJSON, got.ServiceAccountJSON)
		assert.Equal(t, original.ScopeKind, got.ScopeKind)
		assert.Equal(t, original.ProjectID, got.ProjectID)
	})

	t.Run("azure", func(t *testing.T) {
		t.Parallel()

		original := &cloudaccount.AzureCredentials{
			TenantID:          "00000000-0000-0000-0000-000000000000",
			ClientID:          "11111111-1111-1111-1111-111111111111",
			ClientSecret:      "very-secret",
			ScopeKind:         coredata.CloudAccountScopeKindAzureManagementGroup,
			ManagementGroupID: "mg-root",
		}

		raw, err := json.Marshal(original)
		require.NoError(t, err)

		decoded, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderAzure, raw)
		require.NoError(t, err)

		got, ok := decoded.(*cloudaccount.AzureCredentials)
		require.True(t, ok, "expected *AzureCredentials, got %T", decoded)
		assert.Equal(t, original.TenantID, got.TenantID)
		assert.Equal(t, original.ClientID, got.ClientID)
		assert.Equal(t, original.ClientSecret, got.ClientSecret)
		assert.Equal(t, original.ScopeKind, got.ScopeKind)
		assert.Equal(t, original.ManagementGroupID, got.ManagementGroupID)
	})
}

func TestUnmarshalCredentials_UnknownVersion(t *testing.T) {
	t.Parallel()

	envelope := []byte(`{"v":99,"kind":"AWS_ASSUME_ROLE","payload":{}}`)

	_, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderAWS, envelope)
	require.Error(t, err)
	assert.True(t, errors.Is(err, cloudaccount.ErrCredentialsInvalid), "expected ErrCredentialsInvalid, got %v", err)
}

func TestUnmarshalCredentials_UnknownKind(t *testing.T) {
	t.Parallel()

	envelope := []byte(`{"v":1,"kind":"DEFINITELY_NOT_A_KIND","payload":{}}`)

	_, err := cloudaccount.UnmarshalCredentials(coredata.CloudAccountProviderAWS, envelope)
	require.Error(t, err)
	assert.True(t, errors.Is(err, cloudaccount.ErrCredentialsInvalid), "expected ErrCredentialsInvalid, got %v", err)
}
