// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package accessreview

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/pg"
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

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return source, nil
}
