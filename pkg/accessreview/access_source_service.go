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

package accessreview

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
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	NameMaxLength = 1000
)

type (
	AccessSourceService struct {
		pg                *pg.Client
		scope             coredata.Scoper
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connector.ConnectorRegistry
	}

	CreateAccessSourceRequest struct {
		OrganizationID gid.GID
		ConnectorID    *gid.GID
		Name           string
		Category       coredata.AccessSourceCategory
		CsvData        *string
	}

	UpdateAccessSourceRequest struct {
		AccessSourceID gid.GID
		Name           *string
		Category       *coredata.AccessSourceCategory
		ConnectorID    **gid.GID
		CsvData        **string
	}

	ConfigureAccessSourceRequest struct {
		AccessSourceID   gid.GID
		OrganizationSlug string
	}
)

func (r *CreateAccessSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.Category, "category", validator.OneOfSlice(coredata.AccessSourceCategories()))

	return v.Error()
}

func (r *ConfigureAccessSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessSourceID, "access_source_id", validator.Required(), validator.GID(coredata.AccessSourceEntityType))
	v.Check(r.OrganizationSlug, "organization_slug", validator.Required())

	return v.Error()
}

func (r *UpdateAccessSourceRequest) Validate() error {
	v := validator.New()

	v.Check(r.AccessSourceID, "access_source_id", validator.Required(), validator.GID(coredata.AccessSourceEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.Category, "category", validator.OneOfSlice(coredata.AccessSourceCategories()))

	return v.Error()
}

func (s AccessSourceService) Create(
	ctx context.Context,
	req CreateAccessSourceRequest,
) (*coredata.AccessSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	source := &coredata.AccessSource{
		ID:             gid.New(s.scope.GetTenantID(), coredata.AccessSourceEntityType),
		OrganizationID: req.OrganizationID,
		ConnectorID:    req.ConnectorID,
		Name:           req.Name,
		Category:       req.Category,
		CsvData:        req.CsvData,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			// Validate connector exists if provided
			if req.ConnectorID != nil {
				connector := &coredata.Connector{}
				if err := connector.LoadMetadataByID(ctx, conn, s.scope, *req.ConnectorID); err != nil {
					return fmt.Errorf("cannot load connector: %w", err)
				}
			}

			if err := source.Insert(ctx, conn, s.scope); err != nil {
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

func (s AccessSourceService) Get(
	ctx context.Context,
	accessSourceID gid.GID,
) (*coredata.AccessSource, error) {
	source := &coredata.AccessSource{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return source.LoadByID(ctx, conn, s.scope, accessSourceID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get access source: %w", err)
	}

	return source, nil
}

func (s AccessSourceService) Update(
	ctx context.Context,
	req UpdateAccessSourceRequest,
) (*coredata.AccessSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	source := &coredata.AccessSource{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := source.LoadByID(ctx, conn, s.scope, req.AccessSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			if req.Name != nil {
				source.Name = *req.Name
			}

			if req.Category != nil {
				source.Category = *req.Category
			}

			if req.ConnectorID != nil {
				if *req.ConnectorID != nil {
					connector := &coredata.Connector{}
					if err := connector.LoadMetadataByID(ctx, conn, s.scope, **req.ConnectorID); err != nil {
						return fmt.Errorf("cannot load connector: %w", err)
					}
				}
				source.ConnectorID = *req.ConnectorID
			}

			if req.CsvData != nil {
				source.CsvData = *req.CsvData
			}

			source.UpdatedAt = time.Now()

			if err := source.Update(ctx, conn, s.scope); err != nil {
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

func (s AccessSourceService) Delete(
	ctx context.Context,
	accessSourceID gid.GID,
) error {
	source := &coredata.AccessSource{ID: accessSourceID}

	return s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return source.Delete(ctx, conn, s.scope)
		},
	)
}

func (s AccessSourceService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessSourceOrderField],
) (*page.Page[*coredata.AccessSource, coredata.AccessSourceOrderField], error) {
	var sources coredata.AccessSources

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return sources.LoadByOrganizationID(ctx, conn, s.scope, organizationID, cursor)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list access sources: %w", err)
	}

	return page.NewPage(sources, cursor), nil
}

func (s AccessSourceService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			sources := coredata.AccessSources{}
			count, err = sources.CountByOrganizationID(ctx, conn, s.scope, organizationID)
			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count access sources: %w", err)
	}

	return count, nil
}

func (s AccessSourceService) ListScopeSourcesForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
) ([]*coredata.AccessSource, error) {
	var sources coredata.AccessSources

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return sources.LoadScopeSourcesByCampaignID(ctx, conn, s.scope, campaignID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list scope sources: %w", err)
	}

	return sources, nil
}

// ConnectorHTTPClient loads a connector by ID with decrypted credentials
// and returns an HTTP client with token refresh support. If the token was
// refreshed during client creation, the updated credentials are persisted.
func (s AccessSourceService) ConnectorHTTPClient(
	ctx context.Context,
	connectorID gid.GID,
) (*http.Client, *coredata.Connector, error) {
	var dbConnector coredata.Connector

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := dbConnector.LoadByID(ctx, conn, s.scope, connectorID, s.encryptionKey); err != nil {
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
		if err := s.pg.WithConn(
			ctx,
			func(conn pg.Conn) error {
				return dbConnector.Update(ctx, conn, s.scope, s.encryptionKey)
			},
		); err != nil {
			return nil, nil, fmt.Errorf("cannot persist refreshed token: %w", err)
		}
	}

	return httpClient, &dbConnector, nil
}

func (s AccessSourceService) ConfigureAccessSource(
	ctx context.Context,
	req ConfigureAccessSourceRequest,
) (*coredata.AccessSource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	source := &coredata.AccessSource{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := source.LoadByID(ctx, conn, s.scope, req.AccessSourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			if source.ConnectorID == nil {
				return fmt.Errorf("cannot configure access source: no connector attached")
			}

			dbConnector := &coredata.Connector{}
			if err := dbConnector.LoadByID(ctx, conn, s.scope, *source.ConnectorID, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			switch dbConnector.Provider {
			case coredata.ConnectorProviderGitHub:
				if err := dbConnector.SetSettings(&coredata.GitHubConnectorSettings{
					Organization: req.OrganizationSlug,
				}); err != nil {
					return fmt.Errorf("cannot set github settings: %w", err)
				}
			case coredata.ConnectorProviderSentry:
				if err := dbConnector.SetSettings(&coredata.SentryConnectorSettings{
					OrganizationSlug: req.OrganizationSlug,
				}); err != nil {
					return fmt.Errorf("cannot set sentry settings: %w", err)
				}
			default:
				return fmt.Errorf("cannot configure access source: provider %s does not support organization configuration", dbConnector.Provider)
			}

			dbConnector.UpdatedAt = time.Now()

			if err := dbConnector.Update(ctx, conn, s.scope, s.encryptionKey); err != nil {
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
