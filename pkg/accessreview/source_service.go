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

func (s *Service) CreateSource(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateAccessReviewSourceRequest,
) (*coredata.AccessReviewSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
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

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			// Validate connector exists if provided
			if req.ConnectorID != nil {
				connector := &coredata.Connector{}
				if err := connector.LoadMetadataByID(ctx, conn, scope, *req.ConnectorID); err != nil {
					return fmt.Errorf("cannot load connector: %w", err)
				}
			}

			if err := source.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert access source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create access source: %w", err)
	}

	return source, nil
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
			if err := source.LoadByID(ctx, conn, scope, req.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

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
	source := &coredata.AccessReviewSource{}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := source.LoadByID(ctx, conn, scope, accessSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			if err := source.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete access source: %w", err)
			}

			// Garbage-collect the underlying connector once nothing else
			// references it. The connectors table is unique per
			// (organization_id, provider), so leaving an orphaned connector
			// behind would block re-adding a source for the same provider.
			if source.ConnectorID == nil {
				return nil
			}

			accessSources := &coredata.AccessReviewSources{}

			sourceCount, err := accessSources.CountByConnectorID(ctx, conn, scope, *source.ConnectorID)
			if err != nil {
				return fmt.Errorf("cannot count access sources for connector: %w", err)
			}

			if sourceCount > 0 {
				return nil
			}

			bridges := &coredata.SCIMBridges{}

			bridgeCount, err := bridges.CountByConnectorID(ctx, conn, scope, *source.ConnectorID)
			if err != nil {
				return fmt.Errorf("cannot count scim bridges for connector: %w", err)
			}

			if bridgeCount > 0 {
				return nil
			}

			// Garbage-collecting the connector is best-effort. A
			// concurrent transaction may insert a new access source or
			// SCIM bridge referencing this connector between the counts
			// above and the DELETE, producing a foreign-key violation.
			// Run the delete inside a savepoint so such a failure rolls
			// back only the GC attempt and still commits the access
			// source deletion instead of aborting the whole transaction.
			if err := conn.Savepoint(
				ctx,
				func(ctx context.Context, conn pg.Tx) error {
					cnnctr := &coredata.Connector{ID: *source.ConnectorID}
					if err := cnnctr.Delete(ctx, conn, scope); err != nil {
						return fmt.Errorf("cannot delete connector: %w", err)
					}

					return nil
				},
			); err != nil {
				return err
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

// ConnectorHTTPClient loads a connector by ID with decrypted credentials
// and returns an HTTP client with token refresh support. If the token was
// refreshed during client creation, the updated credentials are persisted.
func (s *Service) ConnectorHTTPClient(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*http.Client, *coredata.Connector, error) {
	var dbConnector coredata.Connector

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := dbConnector.LoadByID(ctx, conn, scope, connectorID, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	var tokenBefore string

	oauth2Conn, isOAuth2 := dbConnector.Connection.(*connector.OAuth2Connection)
	if isOAuth2 {
		tokenBefore = oauth2Conn.AccessToken
	}

	var httpClient *http.Client

	if isOAuth2 && s.connectorRegistry != nil {
		refreshCfg := s.connectorRegistry.GetOAuth2RefreshConfig(string(dbConnector.Provider))
		if refreshCfg != nil {
			var err error

			httpClient, err = oauth2Conn.RefreshableClient(ctx, *refreshCfg)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot create refreshable HTTP client: %w", err)
			}
		}
	}

	if httpClient == nil {
		// Inject the Probo-held key for ManagedAPIKey providers (no-op
		// otherwise), resolving it fresh at use time rather than from the
		// connection row.
		if err := s.providerRegistry.ApplyManagedAPIKey(&dbConnector); err != nil {
			return nil, nil, err
		}

		var err error

		httpClient, err = dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot create HTTP client: %w", err)
		}
	}

	// Persist refreshed token if it changed.
	if isOAuth2 && oauth2Conn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()

		if err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return dbConnector.Update(ctx, tx, scope, s.encryptionKey)
			},
		); err != nil {
			return nil, nil, fmt.Errorf("cannot persist refreshed token: %w", err)
		}
	}

	return httpClient, &dbConnector, nil
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

// ProviderOrganizations lists the orgs/workspaces the connector backing the
// source can be scoped to, for the picker UI. Returns an empty list when the
// connector is gone or the provider has no picker.
func (s *Service) ProviderOrganizations(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) ([]drivers.Organization, error) {
	httpClient, dbConnector, err := s.ConnectorHTTPClient(ctx, scope, connectorID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot get connector HTTP client: %w", err)
	}

	cfg, ok := providerOrgConfigs[dbConnector.Provider]
	if !ok || cfg.ListOrgs == nil {
		return nil, nil
	}

	orgs, err := cfg.ListOrgs(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return orgs, nil
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
	if !ok || !cfg.NeedsPicker {
		return false, nil
	}

	return cfg.SelectedSlug(dbConnector) == "", nil
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
	if !ok || !cfg.NeedsPicker || cfg.ListOrgs == nil {
		return
	}

	// Never override an org the user (or an earlier default) already picked.
	if cfg.SelectedSlug(dbMeta) != "" {
		return
	}

	httpClient, dbConnector, err := s.ConnectorHTTPClient(ctx, scope, *source.ConnectorID)
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

	orgs, err := cfg.ListOrgs(listCtx, httpClient)
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
