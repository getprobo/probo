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

package cloudaccount

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// fakeAzureSubscriptions implements the unexported
	// azureSubscriptionsAPI seam declared in azure.go.
	fakeAzureSubscriptions struct {
		err   error
		calls int
	}
)

func (f *fakeAzureSubscriptions) HasOnePage(_ context.Context) error {
	f.calls++
	return f.err
}

func newAzureProviderForTest(
	scope coredata.CloudAccountScopeKind,
	subs *fakeAzureSubscriptions,
) *AzureProvider {
	creds := &AzureCredentials{
		TenantID:       "00000000-0000-0000-0000-000000000000",
		ClientID:       "11111111-1111-1111-1111-111111111111",
		ClientSecret:   "secret",
		ScopeKind:      scope,
		SubscriptionID: "22222222-2222-2222-2222-222222222222",
	}

	rec := CloudAccountRecord{
		ID:              "01HXYZ-cloud-account-azure",
		Provider:        coredata.CloudAccountProviderAzure,
		Kind:            coredata.CloudAccountCredentialKindAzureClientSecret,
		ScopeKind:       scope,
		ScopeIdentifier: creds.SubscriptionID,
	}

	p := newAzureProvider(&http.Client{}, rec, creds)
	p.subscriptions = subs

	return p
}

func TestAzureProvider_Probe_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("subscription scope routes through HasOnePage seam", func(t *testing.T) {
		t.Parallel()

		subs := &fakeAzureSubscriptions{}
		p := newAzureProviderForTest(coredata.CloudAccountScopeKindAzureSubscription, subs)

		err := p.Probe(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, subs.calls, "HasOnePage must be called exactly once")
	})

	t.Run("management group scope routes through HasOnePage seam", func(t *testing.T) {
		t.Parallel()

		// v1 Probe is uniform across Azure scope kinds: it lists one
		// page of subscriptions on behalf of the principal. The MG
		// scope thus also exercises the subscriptions seam.
		subs := &fakeAzureSubscriptions{}
		p := newAzureProviderForTest(coredata.CloudAccountScopeKindAzureManagementGroup, subs)

		err := p.Probe(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, subs.calls)
	})
}

func TestAzureProvider_Probe_MapsResponseError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		status   int
		expectIs error
	}{
		{"401 -> ErrCredentialsInvalid", http.StatusUnauthorized, ErrCredentialsInvalid},
		{"403 -> ErrInsufficientPermissions", http.StatusForbidden, ErrInsufficientPermissions},
		{"404 -> ErrScopeUnreachable", http.StatusNotFound, ErrScopeUnreachable},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subs := &fakeAzureSubscriptions{
				err: &azcore.ResponseError{StatusCode: tc.status, ErrorCode: "stub"},
			}
			p := newAzureProviderForTest(coredata.CloudAccountScopeKindAzureSubscription, subs)

			err := p.Probe(context.Background())
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.expectIs)
		})
	}
}

func TestAzureProvider_Probe_AADStringFingerprint(t *testing.T) {
	t.Parallel()

	// AAD-string fingerprint path: a non-typed error whose message
	// contains an AADSTS code is recognised by MapSDKError as
	// ErrCredentialsInvalid.
	subs := &fakeAzureSubscriptions{
		err: errors.New("AADSTS7000215: Invalid client secret provided"),
	}
	p := newAzureProviderForTest(coredata.CloudAccountScopeKindAzureSubscription, subs)

	err := p.Probe(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCredentialsInvalid)
}

func TestAzureProvider_TokenCredential_BuildsClientSecretCredential(t *testing.T) {
	t.Parallel()

	creds := &AzureCredentials{
		TenantID:     "00000000-0000-0000-0000-000000000000",
		ClientID:     "11111111-1111-1111-1111-111111111111",
		ClientSecret: "secret",
		ScopeKind:    coredata.CloudAccountScopeKindAzureSubscription,
	}
	rec := CloudAccountRecord{
		Provider:  coredata.CloudAccountProviderAzure,
		Kind:      coredata.CloudAccountCredentialKindAzureClientSecret,
		ScopeKind: creds.ScopeKind,
	}

	p := newAzureProvider(&http.Client{}, rec, creds)
	cred, err := p.TokenCredential()
	require.NoError(t, err)
	require.NotNil(t, cred)

	// azidentity.NewClientSecretCredential returns a typed
	// *azidentity.ClientSecretCredential. Asserting the dynamic type
	// pins the construction path.
	_, ok := cred.(*azidentity.ClientSecretCredential)
	assert.True(t, ok, "TokenCredential must be built via azidentity.NewClientSecretCredential, got %T", cred)
}

func TestAzureCredentials_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &AzureCredentials{
		TenantID:          "00000000-0000-0000-0000-000000000000",
		ClientID:          "11111111-1111-1111-1111-111111111111",
		ClientSecret:      "very-secret",
		ScopeKind:         coredata.CloudAccountScopeKindAzureManagementGroup,
		ManagementGroupID: "mg-root",
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	got := &AzureCredentials{}
	require.NoError(t, json.Unmarshal(raw, got))
	assert.Equal(t, original.TenantID, got.TenantID)
	assert.Equal(t, original.ClientID, got.ClientID)
	assert.Equal(t, original.ClientSecret, got.ClientSecret)
	assert.Equal(t, original.ScopeKind, got.ScopeKind)
	assert.Equal(t, original.ManagementGroupID, got.ManagementGroupID)
}

func TestAzureCredentials_UnmarshalRejectsForeignKind(t *testing.T) {
	t.Parallel()

	envelope := []byte(`{"v":1,"kind":"GCP_SERVICE_ACCOUNT_KEY","payload":{}}`)

	got := &AzureCredentials{}
	err := got.UnmarshalJSON(envelope)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCredentialsInvalid))
}

func TestAzureCredentials_EnvelopeIdentity(t *testing.T) {
	t.Parallel()

	c := &AzureCredentials{}
	assert.Equal(t, coredata.CloudAccountProviderAzure, c.Provider())
	assert.Equal(t, coredata.CloudAccountCredentialKindAzureClientSecret, c.Kind())
}

// TestAzureCredentials_EnvelopeRejectsMismatchedProvider mirrors the
// GCP/AWS counterparts: an Azure payload submitted under provider=GCP
// is rejected by UnmarshalCredentials.
func TestAzureCredentials_EnvelopeRejectsMismatchedProvider(t *testing.T) {
	t.Parallel()

	azure := &AzureCredentials{
		TenantID:     "00000000-0000-0000-0000-000000000000",
		ClientID:     "11111111-1111-1111-1111-111111111111",
		ClientSecret: "secret",
		ScopeKind:    coredata.CloudAccountScopeKindAzureSubscription,
	}
	raw, err := json.Marshal(azure)
	require.NoError(t, err)

	_, err = UnmarshalCredentials(coredata.CloudAccountProviderGCP, raw)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCredentialsInvalid))
}
