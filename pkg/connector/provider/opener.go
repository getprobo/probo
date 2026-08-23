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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

type (
	// Opener is the single seam between a stored connector row and a live
	// credential. It owns the whole credential lifecycle — decrypt, mint,
	// write back a rotated token — so no consuming domain holds an encryption
	// key or knows a protocol.
	//
	// It holds the registry plus the deployment material an open needs and a
	// Registration cannot carry: the OAuth app credentials a token refresh
	// authenticates with, and the issuer a workload identity connector
	// exchanges tokens through.
	Opener struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		providers     *Registry
		oauth         *connector.ConnectorRegistry
		issuer        *identityfederation.Issuer
	}

	// Handle is one connector opened for use. It carries the live credential,
	// the row it came from, and the provider's resolved endpoints — everything
	// a capability factory needs and nothing domain-specific.
	Handle struct {
		Connector  *coredata.Connector
		Credential connector.Credential
		Endpoints  Endpoints

		reg *Registration
		// accessTokenBefore and refreshTokenBefore are the OAuth2 tokens as
		// loaded, so PersistIfDirty can tell whether opening the handle
		// refreshed either of them.
		accessTokenBefore  string
		refreshTokenBefore string
	}

	// httpCredential is the stored connection of a connector whose credential
	// an HTTP transport can carry. mintHTTP is the only consumer, which is why
	// the method is not on connector.Connection: that interface describes what
	// is serialized into the encrypted blob, not how it is used.
	httpCredential interface {
		Client(ctx context.Context) (*http.Client, error)
	}
)

// NewOpener returns an Opener over the given registry. oauth may be nil, in
// which case an OAuth2 connector keeps its stored access token instead of
// refreshing. issuer is nil when the deployment serves no federation issuer,
// which makes opening a workload identity connector fail with a clear error
// rather than an opaque rejection from the customer's cloud.
func NewOpener(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	providers *Registry,
	oauth *connector.ConnectorRegistry,
	issuer *identityfederation.Issuer,
) *Opener {
	return &Opener{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		providers:     providers,
		oauth:         oauth,
		issuer:        issuer,
	}
}

// Providers returns the registry behind the opener, for the catalog lookups
// (display names, scopes, settings writers) that need no credential.
func (rt *Opener) Providers() *Registry {
	return rt.providers
}

// Use loads the connector, mints its credential, and hands both to fn as a
// Handle. It writes back an OAuth2 token that minting rotated once fn returns.
//
// No transaction is held while fn runs: a capability built on the handle talks
// to a third-party API, which must never happen with a database connection
// checked out. The credential's lifetime is the callback, which is why callers
// get one rather than a handle they could stash.
func (rt *Opener) Use(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	fn func(context.Context, *Handle) error,
) error {
	handle, err := rt.open(ctx, scope, connectorID)
	if err != nil {
		return err
	}

	fnErr := fn(ctx, handle)

	// The rotated token is persisted even when fn failed: minting already
	// spent the old refresh token, so dropping the new one would strand the
	// connector on a credential the provider no longer honours.
	if err := rt.persist(ctx, scope, handle); err != nil {
		if fnErr != nil {
			return fnErr
		}

		return err
	}

	return fnErr
}

// Modify applies mutate to a connector's decrypted row and writes it back
// inside the caller's transaction, so a connector change commits atomically
// with whatever domain rows accompany it. mutate returns false to leave the row
// untouched, which is how a caller guards against a concurrent write it must
// not clobber.
//
// No credential is minted: this reconfigures the connector rather than using
// it. The encryption key stays here either way.
func (rt *Opener) Modify(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	connectorID gid.GID,
	mutate func(*Registration, *coredata.Connector) (bool, error),
) error {
	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByID(ctx, tx, scope, connectorID, rt.encryptionKey); err != nil {
		return fmt.Errorf("cannot load connector %s: %w", connectorID, err)
	}

	reg, ok := rt.providers.Get(dbConnector.Provider)
	if !ok {
		return fmt.Errorf("cannot modify connector %s: unsupported provider %q", connectorID, dbConnector.Provider)
	}

	changed, err := mutate(reg, dbConnector)
	if err != nil {
		return err
	}

	if !changed {
		return nil
	}

	dbConnector.UpdatedAt = time.Now()

	if err := dbConnector.Update(ctx, tx, scope, rt.encryptionKey); err != nil {
		return fmt.Errorf("cannot update connector %s: %w", connectorID, err)
	}

	return nil
}

// GrantedScopes returns the scopes the connector's stored OAuth2 grant carries,
// and nil for a protocol that holds no grant at all. It decrypts the row
// without minting a credential, so it reaches no third party.
func (rt *Opener) GrantedScopes(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) ([]string, error) {
	dbConnector := &coredata.Connector{}

	err := rt.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return dbConnector.LoadByID(ctx, conn, scope, connectorID, rt.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load connector %s: %w", connectorID, err)
	}

	if dbConnector.Protocol != coredata.ConnectorProtocolOAuth2 {
		return nil, nil
	}

	return connector.GrantedScopes(dbConnector.Connection), nil
}

// open loads and decrypts a connector row, then mints its credential.
func (rt *Opener) open(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*Handle, error) {
	dbConnector := &coredata.Connector{}

	err := rt.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return dbConnector.LoadByID(ctx, conn, scope, connectorID, rt.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load connector %s: %w", connectorID, err)
	}

	return rt.Open(ctx, dbConnector)
}

// persist writes back a credential that minting rotated, and is a no-op
// otherwise — which is the common case, so the transaction is opened only when
// there is something to write.
func (rt *Opener) persist(
	ctx context.Context,
	scope coredata.Scoper,
	handle *Handle,
) error {
	if !handle.credentialRotated() {
		return nil
	}

	return rt.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			handle.Connector.UpdatedAt = time.Now()

			if err := handle.Connector.Update(ctx, tx, scope, rt.encryptionKey); err != nil {
				return fmt.Errorf("cannot persist refreshed token for connector %s: %w", handle.Connector.ID, err)
			}

			return nil
		},
	)
}

// Open readies an already loaded, decrypted connector for use. Prefer Use,
// which owns the load and the write-back too; this is for a caller that
// already holds the row.
func (rt *Opener) Open(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (*Handle, error) {
	reg, ok := rt.providers.Get(dbConnector.Provider)
	if !ok {
		return nil, fmt.Errorf("cannot open connector: unsupported provider %q", dbConnector.Provider)
	}

	h := &Handle{Connector: dbConnector, Endpoints: reg.Endpoints, reg: reg}

	credential, err := rt.mint(ctx, h)
	if err != nil {
		return nil, err
	}

	h.Credential = credential

	return h, nil
}

// mint turns the connector's stored connection into a live credential. It is
// the only place the protocol column is switched on.
func (rt *Opener) mint(ctx context.Context, h *Handle) (connector.Credential, error) {
	switch h.Connector.Protocol {
	case coredata.ConnectorProtocolWorkloadIdentity:
		return rt.mintCloud(ctx, h)

	case coredata.ConnectorProtocolOAuth2, coredata.ConnectorProtocolAPIKey:
		return rt.mintHTTP(ctx, h)
	}

	return nil, fmt.Errorf("cannot open %s connector: unsupported protocol %q", h.Connector.Provider, h.Connector.Protocol)
}

// mintCloud performs the credential exchange a workload identity connector
// federates through. Nothing is persisted afterwards: the SDK renews the
// session's temporary credentials in memory for as long as the caller holds it.
func (rt *Opener) mintCloud(ctx context.Context, h *Handle) (connector.Credential, error) {
	if h.reg.WorkloadIdentity == nil {
		return nil, fmt.Errorf("cannot open connector: provider %q does not federate into a cloud", h.Connector.Provider)
	}

	// A deployment serving no issuer cannot federate at all: the customer's
	// cloud has nothing to validate our tokens against. Say so plainly rather
	// than surface the cloud's opaque AccessDenied later.
	if rt.issuer == nil {
		return nil, fmt.Errorf("cannot open %s connector: identity federation issuer is disabled", h.Connector.Provider)
	}

	session, err := h.reg.WorkloadIdentity.NewSession(ctx, rt.issuer, h.Connector)
	if err != nil {
		return nil, fmt.Errorf("cannot create cloud session for provider %q: %w", h.Connector.Provider, err)
	}

	return connector.CloudCredential{Session: session}, nil
}

// mintHTTP builds the authenticated client an OAuth2 or API-key connector
// speaks through, refreshing an expired OAuth2 token on the way.
func (rt *Opener) mintHTTP(ctx context.Context, h *Handle) (connector.Credential, error) {
	if h.Connector.Connection == nil {
		return nil, fmt.Errorf("cannot open %s connector: no connection configured", h.Connector.Provider)
	}

	if oauth2Conn, ok := h.Connector.Connection.(*connector.OAuth2Connection); ok {
		h.accessTokenBefore = oauth2Conn.AccessToken
		h.refreshTokenBefore = oauth2Conn.RefreshToken

		if rt.oauth != nil {
			if refreshCfg := rt.oauth.GetOAuth2RefreshConfig(h.Connector.Provider.String()); refreshCfg != nil {
				httpClient, err := oauth2Conn.RefreshableClient(ctx, *refreshCfg)
				if err != nil {
					return nil, fmt.Errorf("cannot create refreshable HTTP client for %s connector: %w", h.Connector.Provider, err)
				}

				return connector.HTTPCredential{Client: httpClient}, nil
			}
		}
	}

	// Inject the Probo-held key for ManagedAPIKey providers (no-op otherwise),
	// resolving it fresh at use time rather than from the connection row.
	if err := rt.providers.ApplyManagedAPIKey(h.Connector); err != nil {
		return nil, fmt.Errorf("cannot open %s connector: %w", h.Connector.Provider, err)
	}

	cred, ok := h.Connector.Connection.(httpCredential)
	if !ok {
		return nil, fmt.Errorf("cannot open %s connector: %s connection carries no HTTP credential", h.Connector.Provider, h.Connector.Protocol)
	}

	httpClient, err := cred.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP client for %s connector: %w", h.Connector.Provider, err)
	}

	return connector.HTTPCredential{Client: httpClient}, nil
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

	credential, ok := h.Credential.(connector.HTTPCredential)
	if !ok {
		return nil
	}

	probeURL := h.Endpoints.Probe
	if h.reg.BuildProbeURL != nil {
		built, err := h.reg.BuildProbeURL(h.Connector, h.Endpoints)
		if err != nil {
			return fmt.Errorf("cannot build probe URL: %w", err)
		}

		probeURL = built
	}

	return probeGET(ctx, credential.Client, probeURL)
}

// credentialRotated reports whether minting refreshed either stored OAuth2
// token. Providers that rotate refresh tokens (HubSpot, DocuSign) fail on the
// next poll if the old one is reused, so Opener.persist consults this to decide
// whether the row needs writing back.
func (h *Handle) credentialRotated() bool {
	oauth2Conn, ok := h.Connector.Connection.(*connector.OAuth2Connection)

	return ok &&
		(oauth2Conn.AccessToken != h.accessTokenBefore ||
			oauth2Conn.RefreshToken != h.refreshTokenBefore)
}

// NewHTTPHandleForTest returns a Handle backed by the given HTTP client, so a
// provider's capabilities can be exercised without a database row, an OAuth
// app, or a credential exchange.
//
// There is no cloud counterpart: a federating provider mints its credential
// during the open, so a test exercises it through Opener.Open with a fake
// session factory instead of assembling a Handle by hand.
func NewHTTPHandleForTest(
	reg *Registration,
	dbConnector *coredata.Connector,
	httpClient *http.Client,
) *Handle {
	return &Handle{
		Connector:  dbConnector,
		Credential: connector.HTTPCredential{Client: httpClient},
		Endpoints:  reg.Endpoints,
		reg:        reg,
	}
}
