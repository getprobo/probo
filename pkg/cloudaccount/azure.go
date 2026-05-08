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
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// AzureCredentials carry the v1 client-secret triple: the
	// tenant id, the App Registration's client id, and the secret.
	// SubscriptionID/ManagementGroupID disambiguate which scope
	// the customer attached the registration to.
	AzureCredentials struct {
		TenantID          string                         `json:"tenant_id"`
		ClientID          string                         `json:"client_id"`
		ClientSecret      string                         `json:"client_secret"`
		ScopeKind         coredata.CloudAccountScopeKind `json:"scope_kind"`
		SubscriptionID    string                         `json:"subscription_id,omitempty"`
		ManagementGroupID string                         `json:"management_group_id,omitempty"`
	}

	// azureSubscriptionsAPI is the narrow seam AzureProvider depends
	// on for subscription-list probes. Tests inject a stub that
	// returns a one-page paginator without any HTTP traffic.
	azureSubscriptionsAPI interface {
		// HasOnePage returns nil when at least one subscription
		// page can be fetched, mapping the SDK error otherwise.
		HasOnePage(ctx context.Context) error
	}

	// AzureProvider builds typed Azure SDK clients pinned to a
	// single CloudAccountRecord's client-secret credentials.
	AzureProvider struct {
		httpClient  *http.Client
		record      CloudAccountRecord
		credentials *AzureCredentials

		// Test seam. Production builds leave it nil and the real
		// SDK client is built lazily on demand.
		subscriptions azureSubscriptionsAPI
	}

	// azureHTTPClientAdapter wraps a *http.Client into the
	// policy.Transporter interface the Azure SDK expects.
	azureHTTPClientAdapter struct {
		client *http.Client
	}
)

// Compile-time interface assertions.
var (
	_ Credentials = (*AzureCredentials)(nil)
	_ Probeable   = (*AzureProvider)(nil)
)

func (c *AzureCredentials) Provider() coredata.CloudAccountProvider {
	return coredata.CloudAccountProviderAzure
}

func (c *AzureCredentials) Kind() coredata.CloudAccountCredentialKind {
	return coredata.CloudAccountCredentialKindAzureClientSecret
}

// MarshalJSON wraps the Azure credentials payload in the versioned
// envelope.
func (c *AzureCredentials) MarshalJSON() ([]byte, error) {
	return MarshalEnvelope(c.Kind(), struct {
		TenantID          string                         `json:"tenant_id"`
		ClientID          string                         `json:"client_id"`
		ClientSecret      string                         `json:"client_secret"`
		ScopeKind         coredata.CloudAccountScopeKind `json:"scope_kind"`
		SubscriptionID    string                         `json:"subscription_id,omitempty"`
		ManagementGroupID string                         `json:"management_group_id,omitempty"`
	}{
		TenantID:          c.TenantID,
		ClientID:          c.ClientID,
		ClientSecret:      c.ClientSecret,
		ScopeKind:         c.ScopeKind,
		SubscriptionID:    c.SubscriptionID,
		ManagementGroupID: c.ManagementGroupID,
	})
}

// UnmarshalJSON accepts either the bare payload or the full
// envelope shape; rejects non-Azure-compatible kinds with
// ErrCredentialsInvalid.
func (c *AzureCredentials) UnmarshalJSON(data []byte) error {
	var env credentialsEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("cannot unmarshal azure credentials envelope: %w", err)
	}

	payloadBytes := data
	if env.V != 0 || env.Kind != "" {
		if env.Kind != coredata.CloudAccountCredentialKindAzureClientSecret {
			return fmt.Errorf("cannot unmarshal azure credentials: kind %q: %w", env.Kind, ErrCredentialsInvalid)
		}
		payloadBytes = env.Payload
	}

	var payload struct {
		TenantID          string                         `json:"tenant_id"`
		ClientID          string                         `json:"client_id"`
		ClientSecret      string                         `json:"client_secret"`
		ScopeKind         coredata.CloudAccountScopeKind `json:"scope_kind"`
		SubscriptionID    string                         `json:"subscription_id,omitempty"`
		ManagementGroupID string                         `json:"management_group_id,omitempty"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("cannot unmarshal azure credentials payload: %w", err)
	}

	c.TenantID = payload.TenantID
	c.ClientID = payload.ClientID
	c.ClientSecret = payload.ClientSecret
	c.ScopeKind = payload.ScopeKind
	c.SubscriptionID = payload.SubscriptionID
	c.ManagementGroupID = payload.ManagementGroupID

	return nil
}

// newAzureProvider is the package-internal constructor invoked by
// Registry.BuildAzureProvider.
func newAzureProvider(httpClient *http.Client, rec CloudAccountRecord, creds *AzureCredentials) *AzureProvider {
	return &AzureProvider{
		httpClient:  httpClient,
		record:      rec,
		credentials: creds,
	}
}

// TokenCredential builds an azcore.TokenCredential from the
// record's client-secret triple. Callers that need an arbitrary
// ARM/Graph SDK client use this; common consumers prefer the typed
// helpers below.
func (p *AzureProvider) TokenCredential() (azcore.TokenCredential, error) {
	opts := &azidentity.ClientSecretCredentialOptions{}
	if p.httpClient != nil {
		opts.ClientOptions = azcore.ClientOptions{
			Transport: &azureHTTPClientAdapter{client: p.httpClient},
		}
	}

	cred, err := azidentity.NewClientSecretCredential(
		p.credentials.TenantID,
		p.credentials.ClientID,
		p.credentials.ClientSecret,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build azure client secret credential: %w", err)
	}

	return cred, nil
}

// armClientOptions returns the *arm.ClientOptions the typed
// helpers below use to wire the Probo HTTP client (SSRF-protected)
// into every Azure SDK request.
func (p *AzureProvider) armClientOptions() *arm.ClientOptions {
	if p.httpClient == nil {
		return nil
	}

	return &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Transport: &azureHTTPClientAdapter{client: p.httpClient},
		},
	}
}

// Subscriptions returns an *armsubscription.SubscriptionsClient
// pinned to the record's credentials. Used by the access-review
// driver and the probe path.
func (p *AzureProvider) Subscriptions() (*armsubscription.SubscriptionsClient, error) {
	cred, err := p.TokenCredential()
	if err != nil {
		return nil, err
	}

	client, err := armsubscription.NewSubscriptionsClient(cred, p.armClientOptions())
	if err != nil {
		return nil, fmt.Errorf("cannot build azure subscriptions client: %w", err)
	}

	return client, nil
}

// Probe verifies the client-secret triple by listing one page of
// subscriptions on behalf of the principal. Errors are mapped to
// typed package sentinels via MapSDKError.
func (p *AzureProvider) Probe(ctx context.Context) error {
	if p.subscriptions != nil {
		if err := p.subscriptions.HasOnePage(ctx); err != nil {
			return MapSDKError(fmt.Errorf("cannot probe azure subscriptions: %w", err))
		}

		return nil
	}

	client, err := p.Subscriptions()
	if err != nil {
		return err
	}

	pager := client.NewListPager(nil)
	if !pager.More() {
		return nil
	}

	if _, err := pager.NextPage(ctx); err != nil {
		return MapSDKError(fmt.Errorf("cannot probe azure subscriptions: %w", err))
	}

	return nil
}

// Record returns the underlying CloudAccountRecord. Useful for
// drivers that need to discriminate on ScopeKind without
// re-threading the value through their constructor.
func (p *AzureProvider) Record() CloudAccountRecord {
	return p.record
}

// Do implements policy.Transporter so the Azure SDK can route
// every request through Probo's SSRF-protected HTTP client.
func (a *azureHTTPClientAdapter) Do(req *http.Request) (*http.Response, error) {
	return a.client.Do(req)
}
