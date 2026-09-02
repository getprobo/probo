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

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	NameMaxLength = 1000
)

type (
	CreateAccessReviewSourceRequest struct {
		OrganizationID gid.GID
		ConnectorID    *gid.GID
		Name           string
		CsvData        *string
	}

	UpdateAccessReviewSourceRequest struct {
		AccessReviewSourceID gid.GID
		Name                 **string
		ConnectorID          **gid.GID
		CsvData              **string
	}

	ConfigureAccessReviewSourceRequest struct {
		AccessReviewSourceID gid.GID
		OrganizationSlug     string

		// OnlyIfUnset makes the configure a no-op when the connector already
		// has an org selected. AutoSelectDefaultOrganization sets it so a
		// concurrent user pick made while ListOrgs was in flight is not
		// silently overwritten by the first listed org.
		OnlyIfUnset bool
	}
)

func (r *CreateAccessReviewSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))

	return v.Error()
}

func (r *ConfigureAccessReviewSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessReviewSourceID, "access_review_source_id", validator.Required(), validator.GID(coredata.AccessReviewSourceEntityType))
	v.Check(r.OrganizationSlug, "organization_slug", validator.Required())

	return v.Error()
}

func (r *UpdateAccessReviewSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessReviewSourceID, "access_review_source_id", validator.Required(), validator.GID(coredata.AccessReviewSourceEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))

	return v.Error()
}

// EnsureSource returns the access source for req.ConnectorID,
// creating it when absent; an existing source is returned untouched
// with created=false. The partial unique index on connector_id
// arbitrates concurrent callers, so exactly one inserts and the
// others load the winner. CSV sources (no connector) are always
// created.
func (s *Service) EnsureSource(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateAccessReviewSourceRequest,
) (*coredata.AccessReviewSource, bool, error) {
	if err := req.Validate(); err != nil {
		return nil, false, err
	}

	now := time.Now()
	source := &coredata.AccessReviewSource{
		ID:             gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: req.OrganizationID,
		ConnectorID:    req.ConnectorID,
		Name:           req.Name,
		CsvData:        req.CsvData,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	created := false

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if req.ConnectorID != nil {
				connector := &coredata.Connector{}
				if err := connector.LoadMetadataByID(ctx, conn, scope, *req.ConnectorID); err != nil {
					return fmt.Errorf("cannot load connector: %w", err)
				}

				// Unlocked read: a concurrent bridge bind could in theory
				// race this check. Organic flows only ever bind their own
				// freshly created connector, so the race is accepted
				// rather than serialized.
				bridges := &coredata.SCIMBridges{}

				bridgeCount, err := bridges.CountByConnectorID(ctx, conn, scope, *req.ConnectorID)
				if err != nil {
					return fmt.Errorf("cannot count scim bridges for connector: %w", err)
				}

				if bridgeCount > 0 {
					return fmt.Errorf("cannot create access source: connector is used by a SCIM bridge: %w", coredata.ErrResourceInUse)
				}
			}

			inserted, err := source.Insert(ctx, conn, scope)
			if err != nil {
				return fmt.Errorf("cannot insert access source: %w", err)
			}

			if inserted {
				created = true
				return nil
			}

			existing := &coredata.AccessReviewSource{}
			if err := existing.LoadByConnectorID(ctx, conn, scope, *req.ConnectorID); err != nil {
				return fmt.Errorf("cannot load access source by connector: %w", err)
			}

			*source = *existing

			return nil
		},
	)
	if err != nil {
		return nil, false, fmt.Errorf("cannot create access source: %w", err)
	}

	return source, created, nil
}

func (s *Service) GetSource(
	ctx context.Context,
	scope coredata.Scoper,
	accessSourceID gid.GID,
) (*coredata.AccessReviewSource, error) {
	source := &coredata.AccessReviewSource{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return source.LoadByID(ctx, conn, scope, accessSourceID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get access source: %w", err)
	}

	return source, nil
}

func (s *Service) UpdateSource(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateAccessReviewSourceRequest,
) (*coredata.AccessReviewSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	source := &coredata.AccessReviewSource{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			// The row lock keeps the connector handoff below stable
			// against a concurrent relink or delete.
			if err := source.LoadByIDForUpdate(ctx, conn, scope, req.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			previousConnectorID := source.ConnectorID

			if req.Name != nil {
				if *req.Name != nil {
					source.Name = **req.Name
				}
			}

			if req.ConnectorID != nil {
				if *req.ConnectorID != nil {
					connector := &coredata.Connector{}
					if err := connector.LoadMetadataByID(ctx, conn, scope, **req.ConnectorID); err != nil {
						return fmt.Errorf("cannot load connector: %w", err)
					}

					bridges := &coredata.SCIMBridges{}

					bridgeCount, err := bridges.CountByConnectorID(ctx, conn, scope, **req.ConnectorID)
					if err != nil {
						return fmt.Errorf("cannot count scim bridges for connector: %w", err)
					}

					if bridgeCount > 0 {
						return fmt.Errorf("cannot update access source: connector is used by a SCIM bridge: %w", coredata.ErrResourceInUse)
					}

					// The partial unique index on connector_id is the
					// guard; this pre-check only produces a clearer
					// error than its 23505.
					other := &coredata.AccessReviewSource{}

					err = other.LoadByConnectorID(ctx, conn, scope, **req.ConnectorID)
					if err == nil && other.ID != source.ID {
						return fmt.Errorf("cannot update access source: connector already referenced by another source")
					}

					if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
						return fmt.Errorf("cannot load access source by connector: %w", err)
					}
				}

				source.ConnectorID = *req.ConnectorID

				// A (re)linked connector may resolve to a different instance
				// name; clear the synced flag so the source-name worker picks
				// the row up and re-resolves it.
				source.NameSyncedAt = nil
			}

			if req.CsvData != nil {
				source.CsvData = *req.CsvData
			}

			source.UpdatedAt = time.Now()

			if err := source.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update access source: %w", err)
			}

			// A relink took the previous connector's only owner with it;
			// delete the credential in the same transaction.
			if req.ConnectorID != nil && previousConnectorID != nil &&
				(source.ConnectorID == nil || *source.ConnectorID != *previousConnectorID) {
				abandoned := &coredata.Connector{ID: *previousConnectorID}
				if err := abandoned.Delete(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot delete abandoned connector: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot update access source: %w", err)
	}

	return source, nil
}

func (s *Service) DeleteSource(
	ctx context.Context,
	scope coredata.Scoper,
	accessSourceID gid.GID,
) error {
	source := &coredata.AccessReviewSource{ID: accessSourceID}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			// RETURNING reads the connector under the DELETE's own row
			// lock, so a concurrent relink cannot swap it unobserved.
			connectorID, err := source.DeleteReturningConnectorID(ctx, conn, scope)
			if err != nil {
				return fmt.Errorf("cannot delete access source: %w", err)
			}

			// The source was the connector's only owner; the credential
			// dies with it in the same transaction.
			if connectorID == nil {
				return nil
			}

			cnnctr := &coredata.Connector{ID: *connectorID}
			if err := cnnctr.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete connector: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListSourcesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewSourceOrderField],
) (*page.Page[*coredata.AccessReviewSource, coredata.AccessReviewSourceOrderField], error) {
	var sources coredata.AccessReviewSources

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return sources.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list access sources: %w", err)
	}

	return page.NewPage(sources, cursor), nil
}

func (s *Service) CountSourcesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			sources := coredata.AccessReviewSources{}
			count, err = sources.CountByOrganizationID(ctx, conn, scope, organizationID)

			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count access sources: %w", err)
	}

	return count, nil
}

// loadConfiguredConnector loads a connector by ID with its connection decrypted
// and the deployment's runtime configuration injected. The raw
// ErrResourceNotFound is propagated so callers can decide how to treat a missing
// connector.
func (s *Service) loadConfiguredConnector(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*coredata.Connector, error) {
	dbConnector := &coredata.Connector{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return dbConnector.LoadByID(ctx, conn, scope, connectorID, s.encryptionKey)
		},
	)
	if err != nil {
		return nil, err
	}

	if err := s.connectorRegistry.ConfigureConnection(
		string(dbConnector.Provider),
		dbConnector.Connection,
	); err != nil {
		return nil, fmt.Errorf("cannot configure connector connection: %w", err)
	}

	return dbConnector, nil
}

// BuildHTTPClient loads a connector by ID with decrypted credentials
// and returns an HTTP client with token refresh support. If the token was
// refreshed during client creation, the updated credentials are persisted.
//
// It fails for a connector whose credential does not ride on HTTP (workload
// identity). Callers that must handle both kinds should branch on the protocol
// from loadConnectorMetadata first, as ProbeConnector does.
func (s *Service) BuildHTTPClient(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*http.Client, *coredata.Connector, error) {
	dbConnector, err := s.loadConfiguredConnector(ctx, scope, connectorID)
	if err != nil {
		return nil, nil, err
	}

	conn, ok := dbConnector.Connection.(connector.HTTPConnection)
	if !ok {
		return nil, nil, fmt.Errorf(
			"cannot create HTTP client for %s connector: credential does not ride on HTTP",
			dbConnector.Provider,
		)
	}

	httpClient, err := s.httpClientFor(ctx, scope, dbConnector, conn)
	if err != nil {
		return nil, nil, err
	}

	return httpClient, dbConnector, nil
}

// httpClientFor builds the client for an already-loaded HTTP connection,
// persisting an OAuth2 access token the refresh replaced so later calls and
// other workers use it.
func (s *Service) httpClientFor(
	ctx context.Context,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
	conn connector.HTTPConnection,
) (*http.Client, error) {
	var tokenBefore string

	oauth2Conn, isOAuth2 := conn.(*connector.OAuth2Connection)
	if isOAuth2 {
		tokenBefore = oauth2Conn.AccessToken
	}

	httpClient, err := buildHTTPClient(
		ctx,
		s.connectorRegistry,
		s.providerRegistry,
		dbConnector.Provider,
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP client: %w", err)
	}

	if isOAuth2 && oauth2Conn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()

		if err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return dbConnector.Update(ctx, tx, scope, s.encryptionKey)
			},
		); err != nil {
			return nil, fmt.Errorf("cannot persist refreshed token: %w", err)
		}
	}

	return httpClient, nil
}

func (s *Service) ConfigureAccessReviewSource(
	ctx context.Context,
	scope coredata.Scoper,
	req ConfigureAccessReviewSourceRequest,
) (*coredata.AccessReviewSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	source := &coredata.AccessReviewSource{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := source.LoadByID(ctx, conn, scope, req.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			if source.ConnectorID == nil {
				return fmt.Errorf("cannot configure access source: no connector attached")
			}

			dbConnector := &coredata.Connector{}
			if err := dbConnector.LoadByID(ctx, conn, scope, *source.ConnectorID, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			// TOCTOU guard for the auto-default path: if the org was set (e.g.
			// by a concurrent user pick) after the caller observed it as unset,
			// leave the existing selection untouched.
			if req.OnlyIfUnset {
				if cfg, ok := providerOrgConfigs[dbConnector.Provider]; ok && cfg.SelectedSlug(dbConnector) != "" {
					return nil
				}
			}

			reg, ok := s.providerRegistry.Get(dbConnector.Provider)
			if !ok || reg.SetOrganizationSettings == nil {
				return fmt.Errorf("cannot configure access source: provider %s does not support organization configuration", dbConnector.Provider)
			}

			if err := reg.SetOrganizationSettings(dbConnector, req.OrganizationSlug); err != nil {
				return fmt.Errorf("cannot set %s settings: %w", dbConnector.Provider, err)
			}

			dbConnector.UpdatedAt = time.Now()

			if err := dbConnector.Update(ctx, conn, scope, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot update connector: %w", err)
			}

			// The selected org changed, so the resolvable instance name may
			// have too; clear the synced flag so the source-name worker
			// re-resolves the display name.
			source.NameSyncedAt = nil
			source.UpdatedAt = time.Now()

			if err := source.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot reset access source name sync: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return source, nil
}

// loadConnectorMetadata loads a connector's metadata (provider, settings)
// without decrypting the connection. The raw ErrResourceNotFound is
// propagated so callers can decide how to treat a missing connector.
func (s *Service) loadConnectorMetadata(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*coredata.Connector, error) {
	dbConnector := &coredata.Connector{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return dbConnector.LoadMetadataByID(ctx, conn, scope, connectorID)
		},
	)
	if err != nil {
		return nil, err
	}

	return dbConnector, nil
}

// ProbeConnector verifies the connector's credential is still accepted by the
// provider, taking the HTTP or the cloud path according to the connection's own
// credential model. A nil return means connected (or that the provider
// registers no probe); coredata.ErrResourceNotFound means the connector is
// gone. Keeping the branch here rather than in the resolver is what stops a
// healthy workload identity connector from reporting itself disconnected.
func (s *Service) ProbeConnector(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) error {
	dbConnector, err := s.loadConfiguredConnector(ctx, scope, connectorID)
	if err != nil {
		return err
	}

	// Only a ProbeError means the credential or the provider is at fault;
	// everything else returned here is Probo's own.
	switch conn := dbConnector.Connection.(type) {
	case *connector.WorkloadIdentityConnection:
		session, err := s.buildCloudSession(ctx, dbConnector)
		if err != nil {
			return err
		}

		if err := s.providerRegistry.ProbeCloudConnection(ctx, session, dbConnector); err != nil {
			if !IsProviderVerdict(err) {
				return err
			}

			return NewProbeError(dbConnector.Provider, err)
		}

		return nil

	case connector.HTTPConnection:
		// The eager token refresh runs here, so a revoked grant fails the
		// probe before any request is made.
		httpClient, err := s.httpClientFor(ctx, scope, dbConnector, conn)
		if err != nil {
			if !IsProviderVerdict(err) {
				return err
			}

			return NewProbeError(dbConnector.Provider, err)
		}

		if err := s.providerRegistry.ProbeConnection(ctx, httpClient, dbConnector); err != nil {
			if !IsProviderVerdict(err) {
				return err
			}

			return NewProbeError(dbConnector.Provider, err)
		}

		return nil

	default:
		return fmt.Errorf(
			"cannot probe %s connector: credential is of no known kind",
			dbConnector.Provider,
		)
	}
}

// buildCloudSession opens authenticated access to the cloud account a workload
// identity connector points at, delegating to the provider that knows which
// role and region its settings name.
func (s *Service) buildCloudSession(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (cloud.Session, error) {
	if s.federation == nil {
		return nil, fmt.Errorf(
			"cannot reach %s connector: identity federation is not configured in this deployment",
			dbConnector.Provider,
		)
	}

	reg, ok := s.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil {
		return nil, fmt.Errorf(
			"cannot reach %s connector: provider offers no workload identity path",
			dbConnector.Provider,
		)
	}

	session, err := reg.WorkloadIdentity.NewSession(ctx, s.federation, dbConnector)
	if err != nil {
		return nil, fmt.Errorf("cannot open cloud session for %s connector: %w", dbConnector.Provider, err)
	}

	return session, nil
}

// ProviderOrganizations lists the orgs/workspaces the connector backing the
// source can be scoped to, for the picker UI. Returns an empty list when the
// connector is gone or the provider has no picker.
func (s *Service) ProviderOrganizations(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) ([]drivers.Organization, error) {
	// Resolve the picker from cheap metadata first. Only a provider that has one
	// should pay for the connector decrypt, token refresh, and HTTP-client build
	// below — and a workload identity connector, which has no picker and no HTTP
	// credential, must not reach them at all.
	dbMeta, err := s.loadConnectorMetadata(ctx, scope, connectorID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load connector metadata: %w", err)
	}

	cfg, ok := providerOrgConfigs[dbMeta.Provider]
	if !ok || !ProviderSupportsOrganizationPicker(dbMeta.Provider, dbMeta.Protocol) {
		return nil, nil
	}

	httpClient, dbConnector, err := s.BuildHTTPClient(ctx, scope, connectorID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot get connector HTTP client: %w", err)
	}

	orgs, err := cfg.ListOrgs(ctx, httpClient, s.providerListBaseURL(dbConnector.Provider))
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

// providerListBaseURL returns the base URL a picker's ListOrgs call should
// target: the provider registration's static API root (Endpoints.APIBase),
// falling back to Endpoints.Identity when the provider has no static data
// root of its own. DocuSign is that case — it declares no APIBase, but its
// Identity host is exactly the host ListDocuSignOrganizations needs, so the
// fallback lets an Identity override reach the picker the same way it
// reaches the driver and name resolver. Returns "" when the provider is not
// registered or declares neither; listers treat "" as "no override" and
// fall back to their production base.
func (s *Service) providerListBaseURL(connectorProvider coredata.ConnectorProvider) string {
	reg, ok := s.providerRegistry.Get(connectorProvider)
	if !ok {
		return ""
	}

	if reg.Endpoints.APIBase != "" {
		return reg.Endpoints.APIBase
	}

	return reg.Endpoints.Identity
}

// SelectedOrganizationSlug returns the org identifier currently configured on
// the connector backing the source, or "" when none is set or the provider
// has no picker. ErrResourceNotFound is propagated for a missing connector.
func (s *Service) SelectedOrganizationSlug(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (string, error) {
	dbConnector, err := s.loadConnectorMetadata(ctx, scope, connectorID)
	if err != nil {
		return "", err
	}

	cfg, ok := providerOrgConfigs[dbConnector.Provider]
	if !ok {
		return "", nil
	}

	return cfg.SelectedSlug(dbConnector), nil
}

// SourceNeedsConfiguration reports whether the connector backing the source
// has a picker UI and no org selected yet. ErrResourceNotFound is propagated
// for a missing connector.
func (s *Service) SourceNeedsConfiguration(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (bool, error) {
	dbConnector, err := s.loadConnectorMetadata(ctx, scope, connectorID)
	if err != nil {
		return false, err
	}

	cfg, ok := providerOrgConfigs[dbConnector.Provider]
	if !ok || !ProviderSupportsOrganizationPicker(dbConnector.Provider, dbConnector.Protocol) {
		return false, nil
	}

	return cfg.SelectedSlug(dbConnector) == "", nil
}

// SourceMissingOAuthScopes returns the OAuth scopes required by the current
// provider registration that are absent from the connector's stored grant.
// Only OAuth2 connectors are checked: API-key (and other non-OAuth)
// credentials have no grant scopes and return an empty slice, even when the
// provider also advertises OAuth2Scopes for its dual-auth path.
// ErrResourceNotFound is propagated for a missing connector.
func (s *Service) SourceMissingOAuthScopes(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) ([]string, error) {
	var dbConnector coredata.Connector

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := dbConnector.LoadByID(ctx, conn, scope, connectorID, s.encryptionKey); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	required := s.providerRegistry.ProviderOAuth2Scopes(dbConnector.Provider)

	return missingOAuthScopesForConnector(dbConnector, required), nil
}

// SourceNeedsReconnect reports whether the connector is missing OAuth scopes
// required by the current provider registration. ErrResourceNotFound is
// propagated for a missing connector.
func (s *Service) SourceNeedsReconnect(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (bool, error) {
	missing, err := s.SourceMissingOAuthScopes(ctx, scope, connectorID)
	if err != nil {
		return false, err
	}

	return len(missing) > 0, nil
}

// missingOAuthScopesForConnector returns scopes in required that are absent
// from the connector's stored OAuth grant. Connections that do not support
// scope-grant checks (API key, install-scoped apps, …) and empty required
// lists yield an empty result. A nil Connection falls back to the protocol
// capability probe.
func missingOAuthScopesForConnector(
	dbConnector coredata.Connector,
	required []string,
) []string {
	if !connector.SupportsScopeGrantCheckFor(
		dbConnector.Connection,
		connector.ProtocolType(dbConnector.Protocol),
	) {
		return []string{}
	}

	if len(required) == 0 {
		return []string{}
	}

	var granted []string
	if dbConnector.Connection != nil {
		granted = dbConnector.Connection.Scopes()
	}

	// Microsoft (and similar OIDC providers) omit offline_access from the
	// token scope echo even when a refresh token was issued. Treat a
	// refresh token as proof of that grant for missing-scope checks only —
	// never synthesize it into Connection.Scopes(), which reconnect uses
	// to build the next authorize request (Google rejects offline_access).
	if connectionHasRefreshToken(dbConnector.Connection) {
		granted = connector.UnionScopes(granted, []string{"offline_access"})
	}

	return connector.MissingScopes(required, granted)
}

func connectionHasRefreshToken(c connector.Connection) bool {
	switch conn := c.(type) {
	case *connector.OAuth2Connection:
		return conn.RefreshToken != ""
	case *connector.SlackConnection:
		return conn.RefreshToken != ""
	default:
		return false
	}
}

// AutoSelectDefaultOrganization picks the first workspace/org a freshly linked
// picker-provider source can see when none is selected yet, so the source is
// usable immediately instead of failing its first campaign fetch. The picker
// stays available to switch when several are listed.
//
// Best-effort: any failure leaves the source in its "needs configuration"
// state (the picker is the fallback); it never errors and must not fail the
// create/update that triggered it.
func (s *Service) AutoSelectDefaultOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) {
	if source == nil || source.ConnectorID == nil {
		return
	}

	// Resolve the provider from cheap metadata first: only picker providers
	// that still need defaulting should pay for the connector decrypt, token
	// refresh, and HTTP-client build below (all ~50 other providers skip it).
	dbMeta, err := s.loadConnectorMetadata(ctx, scope, *source.ConnectorID)
	if err != nil {
		// A missing connector is not worth logging: the picker simply never
		// surfaces a default.
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			s.logger.WarnCtx(ctx, "cannot load connector metadata for default organization", log.Error(err))
		}

		return
	}

	cfg, ok := providerOrgConfigs[dbMeta.Provider]
	if !ok || !ProviderSupportsOrganizationPicker(dbMeta.Provider, dbMeta.Protocol) {
		return
	}

	// Never override an org the user (or an earlier default) already picked.
	if cfg.SelectedSlug(dbMeta) != "" {
		return
	}

	httpClient, dbConnector, err := s.BuildHTTPClient(ctx, scope, *source.ConnectorID)
	if err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			s.logger.WarnCtx(ctx, "cannot load connector for default organization", log.Error(err))
		}

		return
	}

	// Bound the outbound provider call so a hung provider cannot stall the
	// create/update mutation that triggered the defaulting.
	listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	orgs, err := cfg.ListOrgs(listCtx, httpClient, s.providerListBaseURL(dbMeta.Provider))
	if err != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot list provider organizations for default selection",
			log.String("provider", dbConnector.Provider.String()),
			log.Error(err),
		)

		return
	}

	if len(orgs) == 0 {
		return
	}

	// OnlyIfUnset guards against a user picking an org while ListOrgs was in
	// flight: the configure re-checks inside its tx and does not overwrite.
	if _, err := s.ConfigureAccessReviewSource(
		ctx,
		scope,
		ConfigureAccessReviewSourceRequest{
			AccessReviewSourceID: source.ID,
			OrganizationSlug:     orgs[0].Slug,
			OnlyIfUnset:          true,
		},
	); err != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot apply default provider organization",
			log.String("provider", dbConnector.Provider.String()),
			log.Error(err),
		)
	}
}

// ResetSourceNameSyncForConnector clears the synced-name flag on every access
// source backed by connectorID so the source-name worker re-resolves the
// display name. Called after a connector is reconnected — the new grant may
// scope a different org/workspace, changing the resolvable name.
func (s *Service) ResetSourceNameSyncForConnector(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			sources := &coredata.AccessReviewSources{}

			return sources.ClearNameSyncedAtByConnectorID(ctx, conn, scope, connectorID)
		},
	)
}
