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
	"slices"
	"strings"
	"sync"

	"go.probo.inc/probo/pkg/gid"
)

type (
	ConnectionConfigurer interface {
		ConfigureConnection(Connection) error
	}

	RuntimeConfigurationRequired interface {
		RequiresRuntimeConfiguration()
	}

	Registry struct {
		sync.RWMutex
		connectors         map[string]Connector
		protocolConnectors map[string]map[ProtocolType]Connector
	}
)

func NewConnectorRegistry() *Registry {
	return &Registry{
		connectors:         make(map[string]Connector),
		protocolConnectors: make(map[string]map[ProtocolType]Connector),
	}
}

func (r *Registry) Register(provider string, c Connector) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.connectors[provider]; ok {
		return fmt.Errorf("cannot register connector %q: already registered", provider)
	}

	r.connectors[provider] = c

	return nil
}

func (r *Registry) RegisterProtocol(provider string, protocol ProtocolType, c Connector) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.protocolConnectors[provider]; !ok {
		r.protocolConnectors[provider] = make(map[ProtocolType]Connector)
	}

	if _, ok := r.protocolConnectors[provider][protocol]; ok {
		return fmt.Errorf(
			"cannot register connector %q with protocol %q: already registered",
			provider,
			protocol,
		)
	}

	r.protocolConnectors[provider][protocol] = c

	return nil
}

func (r *Registry) Get(provider string) (Connector, error) {
	r.RLock()
	defer r.RUnlock()

	c, ok := r.connectors[provider]
	if !ok {
		return nil, fmt.Errorf("cannot find connector %q", provider)
	}

	return c, nil
}

func (r *Registry) GetProtocol(provider string, protocol ProtocolType) (Connector, error) {
	r.RLock()
	defer r.RUnlock()

	connectors, ok := r.protocolConnectors[provider]
	if !ok {
		return nil, fmt.Errorf("cannot find connector %q with protocol %q", provider, protocol)
	}

	c, ok := connectors[protocol]
	if !ok {
		return nil, fmt.Errorf("cannot find connector %q with protocol %q", provider, protocol)
	}

	return c, nil
}

// Lookup returns the connector registered for provider and protocol. OAuth2
// connectors live in the default Register map; other protocols use
// RegisterProtocol.
func (r *Registry) Lookup(provider string, protocol ProtocolType) (Connector, error) {
	if protocol == ProtocolOAuth2 || protocol == "" {
		return r.Get(provider)
	}

	return r.GetProtocol(provider, protocol)
}

// ConfigureConnection injects deployment-held runtime configuration into a
// loaded connection. Credentials injected here are never persisted in the
// connector row, so rotation takes effect without reconnecting every tenant.
func (r *Registry) ConfigureConnection(provider string, conn Connection) error {
	if conn == nil {
		return nil
	}

	if _, ok := conn.(RuntimeConfigurationRequired); !ok {
		return nil
	}

	registered, err := r.Lookup(provider, conn.Type())
	if err != nil {
		return fmt.Errorf("cannot configure connector connection: %w", err)
	}

	configurer, ok := registered.(ConnectionConfigurer)
	if !ok {
		return nil
	}

	if err := configurer.ConfigureConnection(conn); err != nil {
		return fmt.Errorf("cannot configure connector connection: %w", err)
	}

	return nil
}

// ConfiguredProtocols returns the connector protocols that are registered for
// provider in this deployment. OAuth2 connectors registered via Register appear
// as ProtocolOAuth2; protocol-specific connectors registered via RegisterProtocol
// appear as their ProtocolType. The result is sorted for stable GraphQL output.
func (r *Registry) ConfiguredProtocols(provider string) []ProtocolType {
	r.RLock()
	defer r.RUnlock()

	protocols := make([]ProtocolType, 0, 1+len(r.protocolConnectors[provider]))

	if _, ok := r.connectors[provider]; ok {
		protocols = append(protocols, ProtocolOAuth2)
	}

	for protocol := range r.protocolConnectors[provider] {
		protocols = append(protocols, protocol)
	}

	slices.SortFunc(protocols, func(a, b ProtocolType) int {
		return strings.Compare(string(a), string(b))
	})

	return protocols
}

func (r *Registry) Initiate(
	ctx context.Context,
	provider string,
	organizationID gid.GID,
	opts InitiateOptions,
	req *http.Request,
) (string, error) {
	return r.InitiateForProtocol(ctx, provider, ProtocolOAuth2, organizationID, opts, req)
}

// InitiateForProtocol starts the install/auth flow for the connector registered
// under provider and protocol.
func (r *Registry) InitiateForProtocol(
	ctx context.Context,
	provider string,
	protocol ProtocolType,
	organizationID gid.GID,
	opts InitiateOptions,
	req *http.Request,
) (string, error) {
	c, err := r.Lookup(provider, protocol)
	if err != nil {
		return "", fmt.Errorf("cannot initiate connector: %w", err)
	}

	return c.Initiate(ctx, provider, organizationID, opts, req)
}

func (r *Registry) InitiateProtocol(
	ctx context.Context,
	provider string,
	protocol ProtocolType,
	organizationID gid.GID,
	opts InitiateOptions,
	req *http.Request,
) (string, error) {
	return r.InitiateForProtocol(ctx, provider, protocol, organizationID, opts, req)
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

	if payload.Type != OAuth2TokenType {
		return "", fmt.Errorf("cannot extract provider from state token: unexpected token type")
	}

	if payload.Data.Provider == "" {
		return "", fmt.Errorf("cannot extract provider from state token: missing provider field")
	}

	return payload.Data.Provider, nil
}

func (r *Registry) Complete(ctx context.Context, provider string, req *http.Request) (Connection, *gid.GID, string, error) {
	c, err := r.Get(provider)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot complete connector: %w", err)
	}

	return c.Complete(ctx, req)
}

// CompleteWithState completes the OAuth2 flow and returns the full state
// including any reconnection context (ConnectorID).
func (r *Registry) CompleteWithState(ctx context.Context, provider string, req *http.Request) (Connection, *OAuth2State, error) {
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

func (r *Registry) InitiateGitHubApp(
	ctx context.Context,
	organizationID gid.GID,
	opts InitiateOptions,
	req *http.Request,
) (string, error) {
	c, err := r.GetProtocol(GitHubProvider, ProtocolGitHubApp)
	if err != nil {
		return "", fmt.Errorf("cannot initiate github app connector: %w", err)
	}

	return c.Initiate(ctx, GitHubProvider, organizationID, opts, req)
}

func (r *Registry) CompleteGitHubApp(
	ctx context.Context,
	req *http.Request,
) (*GitHubAppConnection, *GitHubAppState, error) {
	c, err := r.GetProtocol(GitHubProvider, ProtocolGitHubApp)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot complete github app connector: %w", err)
	}

	gitHubAppConnector, ok := c.(*GitHubAppConnector)
	if !ok {
		return nil, nil, fmt.Errorf("cannot complete github app connector: invalid connector type")
	}

	return gitHubAppConnector.CompleteWithState(ctx, req)
}

// GitHubAppInstallationURL builds the GitHub App install page for a state
// token issued by Initiate. Callers use this when authorize succeeded but
// the user has no organization installation to bind.
func (r *Registry) GitHubAppInstallationURL(stateToken string) (string, error) {
	c, err := r.GetProtocol(GitHubProvider, ProtocolGitHubApp)
	if err != nil {
		return "", fmt.Errorf("cannot build github app installation URL: %w", err)
	}

	gitHubAppConnector, ok := c.(*GitHubAppConnector)
	if !ok {
		return "", fmt.Errorf("cannot build github app installation URL: invalid connector type")
	}

	return gitHubAppConnector.InstallationURL(stateToken)
}

// GetOAuth2RefreshConfig returns the OAuth2 refresh configuration for a provider.
// Returns nil if the provider is not found or is not an OAuth2 connector.
func (r *Registry) GetOAuth2RefreshConfig(provider string) *OAuth2RefreshConfig {
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
