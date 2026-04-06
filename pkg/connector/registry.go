// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package connector

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"go.probo.inc/probo/pkg/gid"
)

type (
	ConnectorRegistry struct {
		sync.RWMutex
		connectors map[string]Connector
	}
)

func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[string]Connector),
	}
}

func (r *ConnectorRegistry) Register(provider string, c Connector) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.connectors[provider]; ok {
		return fmt.Errorf("cannot register connector %q: already registered", provider)
	}

	r.connectors[provider] = c
	return nil
}

func (r *ConnectorRegistry) Get(provider string) (Connector, error) {
	r.RLock()
	defer r.RUnlock()
	c, ok := r.connectors[provider]
	if !ok {
		return nil, fmt.Errorf("cannot find connector %q", provider)
	}
	return c, nil
}

func (r *ConnectorRegistry) Initiate(ctx context.Context, provider string, organizationID gid.GID, req *http.Request) (string, error) {
	c, err := r.Get(provider)
	if err != nil {
		return "", fmt.Errorf("cannot initiate connector: %w", err)
	}

	return c.Initiate(ctx, provider, organizationID, req)
}

// ExtractProviderFromState decodes the OAuth2 state token without
// verifying its signature and returns the provider name. This allows
// the callback handler to determine which connector to use for
// completing the OAuth2 flow, removing the need for a ?provider=
// query parameter on the redirect URI.
func ExtractProviderFromState(stateToken string) (string, error) {
	payload, err := DecodeOAuth2StatePayload(stateToken)
	if err != nil {
		return "", fmt.Errorf("cannot decode state token: %w", err)
	}

	if payload.Data.Provider == "" {
		return "", fmt.Errorf("cannot extract provider from state token: missing provider field")
	}

	return payload.Data.Provider, nil
}

func (r *ConnectorRegistry) Complete(ctx context.Context, provider string, req *http.Request) (Connection, *gid.GID, string, error) {
	c, err := r.Get(provider)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot complete connector: %w", err)
	}

	return c.Complete(ctx, req)
}

// CompleteWithState completes the OAuth2 flow and returns the full state
// including any reconnection context (ConnectorID).
func (r *ConnectorRegistry) CompleteWithState(ctx context.Context, provider string, req *http.Request) (Connection, *OAuth2State, error) {
	c, err := r.Get(provider)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot complete connector: %w", err)
	}

	oauth2Connector, ok := c.(*OAuth2Connector)
	if !ok {
		return nil, nil, fmt.Errorf("cannot complete connector %q: not an OAuth2 connector", provider)
	}

	return oauth2Connector.CompleteWithState(ctx, req)
}

// providerProbeURLs maps provider names to lightweight API endpoints
// used to verify OAuth token validity. Each URL must accept a GET
// request with a Bearer token and return 401/403 for invalid tokens.
var (
	providerProbeURLs = map[string]string{
		"SLACK":            "https://slack.com/api/users.list?limit=1",
		"GOOGLE_WORKSPACE": "https://admin.googleapis.com/admin/directory/v1/users?customer=my_customer&maxResults=1",
		"LINEAR":           "https://api.linear.app/graphql",
		"BREX":             "https://platform.brexapis.com/v2/users/me",
		"HUBSPOT":          "https://api.hubapi.com/account-info/v3/details",
		"DOCUSIGN":         "https://account-d.docusign.com/oauth/userinfo",
		"NOTION":           "https://api.notion.com/v1/users/me",
		"GITHUB":           "https://api.github.com/user",
		"SENTRY":           "https://sentry.io/api/0/organizations/",
		"INTERCOM":         "https://api.intercom.io/me",
		"CLOUDFLARE":       "https://api.cloudflare.com/client/v4/user/tokens/verify",
		"OPENAI":           "https://api.openai.com/v1/models",
		"SUPABASE":         "https://api.supabase.com/v1/organizations",
		"TALLY":            "https://api.tally.so/me",
		"RESEND":           "https://api.resend.com/domains",
		"ONE_PASSWORD":     "https://events.1password.com/api/v1/auditevents",
	}
)

// GetProbeURL returns the probe URL for a provider.
func (r *ConnectorRegistry) GetProbeURL(provider string) string {
	return providerProbeURLs[provider]
}

// GetOAuth2RefreshConfig returns the OAuth2 refresh configuration for a provider.
// Returns nil if the provider is not found or is not an OAuth2 connector.
func (r *ConnectorRegistry) GetOAuth2RefreshConfig(provider string) *OAuth2RefreshConfig {
	r.RLock()
	defer r.RUnlock()

	c, ok := r.connectors[provider]
	if !ok {
		return nil
	}

	oauth2Connector, ok := c.(*OAuth2Connector)
	if !ok {
		return nil
	}

	return &OAuth2RefreshConfig{
		ClientID:          oauth2Connector.ClientID,
		ClientSecret:      oauth2Connector.ClientSecret,
		TokenURL:          oauth2Connector.TokenURL,
		TokenEndpointAuth: oauth2Connector.TokenEndpointAuth,
	}
}
