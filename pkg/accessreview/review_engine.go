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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

// ReviewEngine contains the stateless core logic for access review campaigns:
// snapshot and source data collection.
type ReviewEngine struct {
	pg                *pg.Client
	predicate         coredata.Predicater
	encryptionKey     cipher.EncryptionKey
	connectorRegistry *connector.ConnectorRegistry
	providerRegistry  *provider.Registry
	logger            *log.Logger
}

func NewReviewEngine(
	pgClient *pg.Client,
	predicate coredata.Predicater,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	providerRegistry *provider.Registry,
	logger *log.Logger,
) *ReviewEngine {
	return &ReviewEngine{
		pg:                pgClient,
		predicate:         predicate,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		providerRegistry:  providerRegistry,
		logger:            logger,
	}
}

// FetchSource pulls accounts from a single source and upserts access entries.
func (e *ReviewEngine) FetchSource(
	ctx context.Context,
	campaign *coredata.AccessReviewCampaign,
	sourceID gid.GID,
) (int, error) {
	fetchedCount := 0

	// Resolve the driver and load baseline data outside the write transaction
	// so that external HTTP calls do not hold a database connection.
	var (
		source   *coredata.AccessSource
		driver   drivers.Driver
		baseline []coredata.BaselineAccountEntry
	)

	err := e.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source = &coredata.AccessSource{}
			if err := source.LoadByID(ctx, tx, e.predicate, sourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
			}

			if source.OrganizationID != campaign.OrganizationID {
				return fmt.Errorf("cannot process access source: %s does not belong to campaign organization", sourceID)
			}

			var err error

			driver, err = e.resolveDriver(ctx, tx, source)
			if err != nil {
				return fmt.Errorf("cannot resolve driver for source %s: %w", source.Name, err)
			}

			lastCompletedCampaign := &coredata.AccessReviewCampaign{}
			if err := lastCompletedCampaign.LoadLastCompletedByOrganizationID(ctx, tx, e.predicate, campaign.OrganizationID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load last completed campaign: %w", err)
				}
			} else {
				entries := &coredata.AccessEntries{}

				baseline, err = entries.LoadBaselineBySourceID(ctx, tx, e.predicate, lastCompletedCampaign.ID, sourceID)
				if err != nil {
					return fmt.Errorf("cannot load baseline entries by source: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	previousByAccountKey := make(map[string]coredata.BaselineAccountEntry, len(baseline))
	for _, entry := range baseline {
		previousByAccountKey[entry.AccountKey] = entry
	}

	sourceCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	accounts, err := driver.ListAccounts(sourceCtx)

	cancel()

	if err != nil {
		return 0, fmt.Errorf("cannot list accounts from source %s: %w", source.Name, err)
	}

	fetchedCount = len(accounts)

	err = e.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			now := time.Now()
			seenAccountKeys := make(map[string]struct{}, len(accounts))

			for _, account := range accounts {
				accountKey := normalizeAccountKey(account.Email, account.ExternalID)
				seenAccountKeys[accountKey] = struct{}{}

				incrementalTag := coredata.AccessEntryIncrementalTagNew
				if _, ok := previousByAccountKey[accountKey]; ok {
					incrementalTag = coredata.AccessEntryIncrementalTagUnchanged
				}

				entry := &coredata.AccessEntry{
					ID:                     gid.New(e.predicate.GetTenantID(), coredata.AccessEntryEntityType),
					OrganizationID:         campaign.OrganizationID,
					AccessReviewCampaignID: campaign.ID,
					AccessSourceID:         sourceID,
					Email:                  account.Email,
					FullName:               account.FullName,
					Role:                   account.Role,
					JobTitle:               account.JobTitle,
					IsAdmin:                account.IsAdmin,
					MFAStatus:              account.MFAStatus,
					AuthMethod:             account.AuthMethod,
					AccountType:            account.AccountType,
					Active:                 account.Active,
					LastLogin:              account.LastLogin,
					AccountCreatedAt:       account.CreatedAt,
					ExternalID:             account.ExternalID,
					AccountKey:             accountKey,
					IncrementalTag:         incrementalTag,
					Flags:                  []coredata.AccessEntryFlag{},
					FlagReasons:            []string{},
					Decision:               coredata.AccessEntryDecisionPending,
					CreatedAt:              now,
					UpdatedAt:              now,
				}

				if err := entry.Upsert(ctx, conn, e.predicate); err != nil {
					return fmt.Errorf("cannot upsert access entry: %w", err)
				}
			}

			// Create REMOVED entries for accounts that existed in the previous
			// campaign but are no longer present in the current fetch.
			for accountKey, prev := range previousByAccountKey {
				if _, seen := seenAccountKeys[accountKey]; seen {
					continue
				}

				entry := &coredata.AccessEntry{
					ID:                     gid.New(e.predicate.GetTenantID(), coredata.AccessEntryEntityType),
					OrganizationID:         campaign.OrganizationID,
					AccessReviewCampaignID: campaign.ID,
					AccessSourceID:         sourceID,
					Email:                  prev.Email,
					FullName:               prev.FullName,
					AccountKey:             accountKey,
					IncrementalTag:         coredata.AccessEntryIncrementalTagRemoved,
					Flags:                  []coredata.AccessEntryFlag{},
					FlagReasons:            []string{},
					Decision:               coredata.AccessEntryDecisionPending,
					MFAStatus:              coredata.MFAStatusUnknown,
					AuthMethod:             coredata.AccessEntryAuthMethodUnknown,
					AccountType:            coredata.AccessEntryAccountTypeUser,
					CreatedAt:              now,
					UpdatedAt:              now,
				}

				if err := entry.Upsert(ctx, conn, e.predicate); err != nil {
					return fmt.Errorf("cannot upsert removed access entry: %w", err)
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
func (e *ReviewEngine) oauthClient(
	ctx context.Context,
	conn *connector.OAuth2Connection,
	provider coredata.ConnectorProvider,
) (*http.Client, error) {
	if e.connectorRegistry != nil {
		refreshCfg := e.connectorRegistry.GetOAuth2RefreshConfig(string(provider))
		if refreshCfg != nil {
			return conn.RefreshableClient(ctx, *refreshCfg)
		}
	}

	return conn.Client(ctx)
}

// connectorHTTPClient returns an HTTP client for the given connector.
// For OAuth2 connections it delegates to oauthClient so that token refresh
// is handled transparently. For other connection types it falls back to
// the standard Client method.
func (e *ReviewEngine) connectorHTTPClient(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (*http.Client, error) {
	if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
		return e.oauthClient(ctx, oauth2Conn, dbConnector.Provider)
	}

	return dbConnector.Connection.Client(ctx)
}

// resolveDriver creates a Driver for the given AccessSource based on
// connector_id (null = built-in, set = connector-backed).
func (e *ReviewEngine) resolveDriver(
	ctx context.Context,
	tx pg.Tx,
	source *coredata.AccessSource,
) (drivers.Driver, error) {
	if source.ConnectorID == nil {
		// CSV-backed source: use CSVDriver when csv_data is present
		if source.CsvData != nil && *source.CsvData != "" {
			return drivers.NewCSVDriver(strings.NewReader(*source.CsvData)), nil
		}

		// Built-in driver: default to ProboMemberships
		return drivers.NewProboMembershipsDriver(e.pg, e.predicate, source.OrganizationID), nil
	}

	// Connector-backed: look up the connector and resolve driver by provider
	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByID(ctx, tx, e.predicate, *source.ConnectorID, e.encryptionKey); err != nil {
		return nil, fmt.Errorf("cannot load connector %s: %w", *source.ConnectorID, err)
	}

	// Capture token before refresh to detect changes.
	var tokenBefore string
	if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
		tokenBefore = oauth2Conn.AccessToken
	}

	// Build an HTTP client. For OAuth2 connections, use RefreshableClient
	// so that short-lived tokens are transparently refreshed.
	httpClient, err := e.connectorHTTPClient(ctx, dbConnector)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP client for %s connector: %w", dbConnector.Provider, err)
	}

	// Persist the refreshed token back to the database so subsequent
	// calls (and other workers) use the updated credentials. Providers
	// that rotate refresh tokens (HubSpot, DocuSign) will fail on the
	// next poll if the old refresh token is reused.
	if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
		if oauth2Conn.AccessToken != tokenBefore {
			dbConnector.UpdatedAt = time.Now()
			if err := dbConnector.Update(ctx, tx, e.predicate, e.encryptionKey); err != nil {
				return nil, fmt.Errorf("cannot persist refreshed token for connector %s: %w", *source.ConnectorID, err)
			}
		}
	}

	reg, ok := e.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.NewDriver == nil {
		return nil, fmt.Errorf("cannot resolve driver: unsupported provider %q", dbConnector.Provider)
	}

	return reg.NewDriver(ctx, httpClient, dbConnector, e.logger)
}
