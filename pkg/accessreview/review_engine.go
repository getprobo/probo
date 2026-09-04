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
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// sourceFetchTimeout is the budget for one driver's ListAccounts call.
// GitHub paginates members, then several best-effort APIs (audit log,
// apps, tokens, deploy keys); 30s was enough to list members but not to
// also walk a long audit log.
const sourceFetchTimeout = 2 * time.Minute

// FetchSource pulls accounts from a single campaign source snapshot and upserts
// access entries against that snapshot.
func (s *Service) FetchSource(
	ctx context.Context,
	scope coredata.Scoper,
	campaign *coredata.AccessReviewCampaign,
	campaignSource *coredata.AccessReviewCampaignSource,
) (int, error) {
	fetchedCount := 0

	if campaignSource.AccessReviewSourceID == nil {
		return 0, fmt.Errorf("cannot fetch source %s: the access source no longer exists", campaignSource.ID)
	}

	sourceID := *campaignSource.AccessReviewSourceID

	// Resolve the driver outside the write transaction so that external HTTP
	// calls do not hold a database connection.
	var (
		source *coredata.AccessReviewSource
		driver drivers.Driver
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source = &coredata.AccessReviewSource{}
			if err := source.LoadByID(ctx, tx, scope, sourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
			}

			var err error

			driver, err = s.resolveDriver(ctx, tx, scope, source)
			if err != nil {
				return fmt.Errorf("cannot resolve driver for source %s: %w", source.Name, err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	sourceCtx, cancel := context.WithTimeout(ctx, sourceFetchTimeout)
	accounts, err := driver.ListAccounts(sourceCtx)

	cancel()

	if err != nil {
		return 0, fmt.Errorf("cannot list accounts from source %s: %w", source.Name, err)
	}

	fetchedCount = len(accounts)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			now := time.Now()

			for _, account := range accounts {
				accountKey := normalizeAccountKey(account.Email, account.ExternalID)

				entry := &coredata.AccessReviewEntry{
					ID:                           gid.New(scope.GetTenantID(), coredata.AccessReviewEntryEntityType),
					OrganizationID:               campaign.OrganizationID,
					AccessReviewCampaignID:       campaign.ID,
					AccessReviewCampaignSourceID: campaignSource.ID,
					Email:                        account.Email,
					FullName:                     account.FullName,
					Roles:                        account.Roles,
					JobTitle:                     account.JobTitle,
					IsAdmin:                      account.IsAdmin,
					MFAStatus:                    account.MFAStatus,
					AuthMethod:                   account.AuthMethod,
					AccountType:                  account.AccountType,
					Active:                       account.Active,
					LastLogin:                    account.LastLogin,
					AccountCreatedAt:             account.CreatedAt,
					ExternalID:                   account.ExternalID,
					AccountKey:                   accountKey,
					Flags:                        []coredata.AccessReviewEntryFlag{},
					FlagReasons:                  []string{},
					Decision:                     coredata.AccessReviewEntryDecisionPending,
					CreatedAt:                    now,
					UpdatedAt:                    now,
				}

				if err := entry.Upsert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot upsert access entry: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return fetchedCount, nil
}

func normalizeAccountKey(email, externalID string) string {
	emailKey := strings.ToLower(strings.TrimSpace(email))

	externalID = strings.TrimSpace(externalID)
	if externalID != "" {
		return emailKey + "|" + externalID
	}

	return emailKey
}

// oauthClient returns an HTTP client for an OAuth2 connection, using
// RefreshableClient when a refresh config is available for the provider.
func oauthClient(
	ctx context.Context,
	connectorRegistry *connector.Registry,
	conn *connector.OAuth2Connection,
	provider coredata.ConnectorProvider,
) (*http.Client, error) {
	refreshCfg := connectorRegistry.GetOAuth2RefreshConfig(string(provider))
	if refreshCfg != nil {
		return conn.RefreshableClient(ctx, *refreshCfg)
	}

	return conn.Client(ctx)
}

// buildHTTPClient returns an HTTP client for the given connection.
// For OAuth2 connections it delegates to oauthClient so that token refresh
// is handled transparently. For API-key connections it overlays a
// Probo-held key onto a copy when the provider is managed, so the loaded
// row is never mutated. Other connection types use the standard Client
// method.
func buildHTTPClient(
	ctx context.Context,
	connectorRegistry *connector.Registry,
	providerRegistry *provider.Registry,
	provider coredata.ConnectorProvider,
	conn connector.HTTPConnection,
) (*http.Client, error) {
	if err := connectorRegistry.ConfigureConnection(string(provider), conn); err != nil {
		return nil, err
	}

	if oauth2Conn, ok := conn.(*connector.OAuth2Connection); ok {
		return oauthClient(ctx, connectorRegistry, oauth2Conn, provider)
	}

	apiKeyConn, ok := conn.(*connector.APIKeyConnection)
	if !ok {
		return conn.Client(ctx)
	}

	key, err := providerRegistry.APIKeyFor(provider, apiKeyConn)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve API key for %s connector: %w", provider, err)
	}

	prepared := *apiKeyConn
	prepared.APIKey = key

	return prepared.Client(ctx)
}

// resolveDriver creates a Driver for the given AccessReviewSource based on
// connector_id (null = built-in, set = connector-backed).
func (s *Service) resolveDriver(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (drivers.Driver, error) {
	if source.ConnectorID == nil {
		// CSV-backed source: use CSVDriver when csv_data is present
		if source.CsvData != nil && *source.CsvData != "" {
			return drivers.NewCSVDriver(strings.NewReader(*source.CsvData)), nil
		}

		// Built-in driver: default to ProboMemberships
		return drivers.NewProboMembershipsDriver(s.pg, scope, source.OrganizationID), nil
	}

	// Connector-backed: look up the connector and resolve driver by provider
	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByID(ctx, tx, scope, *source.ConnectorID, s.encryptionKey); err != nil {
		return nil, fmt.Errorf("cannot load connector %s: %w", *source.ConnectorID, err)
	}

	reg, ok := s.providerRegistry.Get(dbConnector.Provider)
	if !ok {
		return nil, fmt.Errorf("cannot resolve driver: provider %q is not registered", dbConnector.Provider)
	}

	// The connection decides which credential the driver can be built from, so
	// it decides which factory runs. A protocol added later without a factory
	// lands in the default arm and fails loudly rather than fetching nothing.
	switch conn := dbConnector.Connection.(type) {
	case *connector.WorkloadIdentityConnection:
		return s.newCloudDriver(ctx, reg, dbConnector)

	case connector.HTTPConnection:
		return s.newHTTPDriver(ctx, tx, scope, reg, dbConnector, conn)

	default:
		return nil, fmt.Errorf(
			"cannot resolve driver: %s connector has an unsupported credential",
			dbConnector.Provider,
		)
	}
}

// newCloudDriver builds the driver for a workload identity connector from a
// cloud session rather than an HTTP client.
func (s *Service) newCloudDriver(
	ctx context.Context,
	reg *provider.Registration,
	dbConnector *coredata.Connector,
) (drivers.Driver, error) {
	session, err := s.buildCloudSession(ctx, dbConnector)
	if err != nil {
		return nil, err
	}

	return reg.WorkloadIdentity.NewDriver(ctx, session, dbConnector, s.logger)
}

// newHTTPDriver builds the driver for a connector whose credential rides on an
// HTTP client, refreshing and persisting an OAuth2 token on the way.
func (s *Service) newHTTPDriver(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	reg *provider.Registration,
	dbConnector *coredata.Connector,
	conn connector.HTTPConnection,
) (drivers.Driver, error) {
	if reg.NewDriver == nil {
		return nil, fmt.Errorf("cannot resolve driver: provider %q registers no driver", dbConnector.Provider)
	}

	var tokenBefore string
	if oauth2Conn, ok := conn.(*connector.OAuth2Connection); ok {
		tokenBefore = oauth2Conn.AccessToken
	}

	httpClient, err := buildHTTPClient(
		ctx,
		s.connectorRegistry,
		s.providerRegistry,
		reg.Provider,
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP client for %s connector: %w", reg.Provider, err)
	}

	// Persist the refreshed token back to the database so subsequent
	// calls (and other workers) use the updated credentials. Providers
	// that rotate refresh tokens (HubSpot, DocuSign) will fail on the
	// next poll if the old refresh token is reused.
	if oauth2Conn, ok := conn.(*connector.OAuth2Connection); ok {
		if oauth2Conn.AccessToken != tokenBefore {
			dbConnector.UpdatedAt = time.Now()
			if err := dbConnector.Update(ctx, tx, scope, s.encryptionKey); err != nil {
				return nil, fmt.Errorf("cannot persist refreshed token for connector %s: %w", dbConnector.ID, err)
			}
		}
	}

	return reg.NewDriver(ctx, httpClient, dbConnector, s.logger, reg.Endpoints)
}
