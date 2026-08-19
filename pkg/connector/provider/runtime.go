// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/identityfederation"
)

type (
	// Runtime turns a stored connector row into a Handle. It is the registry
	// plus the deployment-held material an open needs and a Registration cannot
	// carry: the OAuth app credentials a token refresh authenticates with, and
	// the issuer a workload identity connector exchanges tokens through.
	//
	// Open is the only place the protocol column is switched on. What a caller
	// then does with the connector comes from the provider's Registration
	// factories, not from the Handle.
	Runtime struct {
		providers *Registry
		oauth     *connector.ConnectorRegistry
		issuer    *identityfederation.Issuer
	}

	// Handle is one connector opened for use, holding exactly one live
	// credential: it probes and persists that credential, and carries nothing
	// domain-specific. Callers pass it to a Registration factory instead.
	//
	// The credential stays unexported: an *http.Client and a cloud.Session
	// share no useful interface, so the split is resolved by the HTTP/Cloud
	// factory adapters rather than at every call site.
	Handle struct {
		Connector *coredata.Connector

		reg          *Registration
		httpClient   *http.Client
		cloudSession cloud.Session
		// accessTokenBefore and refreshTokenBefore are the OAuth2 tokens as
		// loaded, so PersistIfDirty can tell whether opening the handle
		// refreshed either of them.
		accessTokenBefore  string
		refreshTokenBefore string
	}

	// httpCredential is the stored connection of a connector whose credential
	// an HTTP transport can carry. Open is the only consumer, which is why the
	// method is not on connector.Connection: that interface describes what is
	// serialized into the encrypted blob, not how it is used.
	httpCredential interface {
		Client(ctx context.Context) (*http.Client, error)
	}
)

// NewRuntime returns a Runtime over the given registry. oauth may be nil, in
// which case an OAuth2 connector keeps its stored access token instead of
// refreshing. issuer is nil when the deployment serves no federation issuer,
// which makes opening a workload identity connector fail with a clear error
// rather than an opaque rejection from the customer's cloud.
func NewRuntime(
	providers *Registry,
	oauth *connector.ConnectorRegistry,
	issuer *identityfederation.Issuer,
) *Runtime {
	return &Runtime{
		providers: providers,
		oauth:     oauth,
		issuer:    issuer,
	}
}

// Providers returns the registry behind the runtime, for the catalog lookups
// (display names, scopes, settings writers) that need no credential.
func (rt *Runtime) Providers() *Registry {
	return rt.providers
}

// Open readies a loaded, decrypted connector for use. It is the only place the
// protocol column is switched on.
func (rt *Runtime) Open(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (*Handle, error) {
	reg, ok := rt.providers.Get(dbConnector.Provider)
	if !ok {
		return nil, fmt.Errorf("cannot open connector: unsupported provider %q", dbConnector.Provider)
	}

	h := &Handle{Connector: dbConnector, reg: reg}

	switch dbConnector.Protocol {
	case coredata.ConnectorProtocolWorkloadIdentity:
		if err := rt.openCloud(ctx, h); err != nil {
			return nil, err
		}

	case coredata.ConnectorProtocolOAuth2, coredata.ConnectorProtocolAPIKey:
		if err := rt.openHTTP(ctx, h); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("cannot open %s connector: unsupported protocol %q", dbConnector.Provider, dbConnector.Protocol)
	}

	return h, nil
}

// openCloud performs the credential exchange a workload identity connector
// federates through. Nothing is persisted afterwards: the SDK renews the
// session's temporary credentials in memory for as long as the caller holds it.
func (rt *Runtime) openCloud(ctx context.Context, h *Handle) error {
	if h.reg.NewCloudSession == nil {
		return fmt.Errorf("cannot open connector: provider %q does not federate into a cloud", h.Connector.Provider)
	}

	// A deployment serving no issuer cannot federate at all: the customer's
	// cloud has nothing to validate our tokens against. Say so plainly rather
	// than surface the cloud's opaque AccessDenied later.
	if rt.issuer == nil {
		return fmt.Errorf("cannot open %s connector: identity federation issuer is disabled", h.Connector.Provider)
	}

	session, err := h.reg.NewCloudSession(ctx, rt.issuer, h.Connector)
	if err != nil {
		return fmt.Errorf("cannot create cloud session for provider %q: %w", h.Connector.Provider, err)
	}

	h.cloudSession = session

	return nil
}

// openHTTP builds the authenticated client an OAuth2 or API-key connector
// speaks through, refreshing an expired OAuth2 token on the way.
func (rt *Runtime) openHTTP(ctx context.Context, h *Handle) error {
	if h.Connector.Connection == nil {
		return fmt.Errorf("cannot open %s connector: no connection configured", h.Connector.Provider)
	}

	if oauth2Conn, ok := h.Connector.Connection.(*connector.OAuth2Connection); ok {
		h.accessTokenBefore = oauth2Conn.AccessToken
		h.refreshTokenBefore = oauth2Conn.RefreshToken

		if rt.oauth != nil {
			if refreshCfg := rt.oauth.GetOAuth2RefreshConfig(h.Connector.Provider.String()); refreshCfg != nil {
				httpClient, err := oauth2Conn.RefreshableClient(ctx, *refreshCfg)
				if err != nil {
					return fmt.Errorf("cannot create refreshable HTTP client for %s connector: %w", h.Connector.Provider, err)
				}

				h.httpClient = httpClient

				return nil
			}
		}
	}

	// Inject the Probo-held key for ManagedAPIKey providers (no-op otherwise),
	// resolving it fresh at use time rather than from the connection row.
	if err := rt.providers.ApplyManagedAPIKey(h.Connector); err != nil {
		return fmt.Errorf("cannot open %s connector: %w", h.Connector.Provider, err)
	}

	cred, ok := h.Connector.Connection.(httpCredential)
	if !ok {
		return fmt.Errorf("cannot open %s connector: %s connection carries no HTTP credential", h.Connector.Provider, h.Connector.Protocol)
	}

	httpClient, err := cred.Client(ctx)
	if err != nil {
		return fmt.Errorf("cannot create HTTP client for %s connector: %w", h.Connector.Provider, err)
	}

	h.httpClient = httpClient

	return nil
}

// Probe verifies that the provider still accepts the opened credential. It
// dispatches to a provider-specific Probe closure when registered, otherwise
// issues a lightweight GET against Endpoints.Probe or BuildProbeURL. An empty
// probe URL means the check is skipped.
//
// A cloud connector needs no default: obtaining its session already performed
// the credential exchange, which is most of the check.
func (h *Handle) Probe(ctx context.Context) error {
	if h.reg.Probe != nil {
		return h.reg.Probe(ctx, h)
	}

	if h.httpClient == nil {
		return nil
	}

	probeURL := h.reg.Endpoints.Probe
	if h.reg.BuildProbeURL != nil {
		built, err := h.reg.BuildProbeURL(h.Connector, h.reg.Endpoints)
		if err != nil {
			return fmt.Errorf("cannot build probe URL: %w", err)
		}

		probeURL = built
	}

	return probeGET(ctx, h.httpClient, probeURL)
}

// CredentialRotated reports whether opening the handle refreshed either stored
// OAuth2 token, so a caller can skip opening a transaction it would not use.
func (h *Handle) CredentialRotated() bool {
	oauth2Conn, ok := h.Connector.Connection.(*connector.OAuth2Connection)

	return ok &&
		(oauth2Conn.AccessToken != h.accessTokenBefore ||
			oauth2Conn.RefreshToken != h.refreshTokenBefore)
}

// PersistIfDirty writes the connector back when opening it rotated an OAuth2
// token, and is a no-op otherwise. Providers that rotate refresh tokens
// (HubSpot, DocuSign) fail on the next poll if the old one is reused.
func (h *Handle) PersistIfDirty(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	if !h.CredentialRotated() {
		return nil
	}

	h.Connector.UpdatedAt = time.Now()

	if err := h.Connector.Update(ctx, tx, scope, encryptionKey); err != nil {
		return fmt.Errorf("cannot persist refreshed token for connector %s: %w", h.Connector.ID, err)
	}

	return nil
}

// http returns the handle's HTTP client, or an error naming the mismatch when
// the connector was opened with a cloud credential instead. Reaching it means a
// provider registered an HTTP factory for a workload identity protocol, which
// Register cannot see; failing here beats handing a nil client to a driver.
func (h *Handle) http() (*http.Client, error) {
	if h.httpClient == nil {
		return nil, fmt.Errorf("cannot use %s connector over HTTP: its %s credential is not an HTTP one", h.Connector.Provider, h.Connector.Protocol)
	}

	return h.httpClient, nil
}

// cloud is the counterpart of http for a factory written against a cloud SDK.
func (h *Handle) cloud() (cloud.Session, error) {
	if h.cloudSession == nil {
		return nil, fmt.Errorf("cannot use %s connector against a cloud: its %s credential is not a cloud session", h.Connector.Provider, h.Connector.Protocol)
	}

	return h.cloudSession, nil
}

// NewHTTPHandleForTest returns a Handle backed by the given HTTP client, so a
// provider's factories can be exercised without a database row, an OAuth app,
// or a credential exchange.
func NewHTTPHandleForTest(
	reg *Registration,
	dbConnector *coredata.Connector,
	httpClient *http.Client,
) *Handle {
	return &Handle{Connector: dbConnector, reg: reg, httpClient: httpClient}
}

// NewCloudHandleForTest is NewHTTPHandleForTest for a workload identity
// provider, whose factories consume a cloud.Session instead.
func NewCloudHandleForTest(
	reg *Registration,
	dbConnector *coredata.Connector,
	session cloud.Session,
) *Handle {
	return &Handle{Connector: dbConnector, reg: reg, cloudSession: session}
}
