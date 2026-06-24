// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

func (r *ConnectorRegistry) Initiate(
	ctx context.Context,
	provider string,
	organizationID gid.GID,
	opts InitiateOptions,
	req *http.Request,
) (string, error) {
	c, err := r.Get(provider)
	if err != nil {
		return "", fmt.Errorf("cannot initiate connector: %w", err)
	}

	return c.Initiate(ctx, provider, organizationID, opts, req)
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
