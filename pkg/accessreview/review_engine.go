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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

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

	// Resolve the driver and load baseline data outside the write transaction
	// so that external HTTP calls do not hold a database connection.
	var (
		source   *coredata.AccessReviewSource
		driver   drivers.Driver
		baseline []coredata.BaselineAccountEntry
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source = &coredata.AccessReviewSource{}
			if err := source.LoadByID(ctx, tx, scope, sourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
			}

			if source.OrganizationID != campaign.OrganizationID {
				return fmt.Errorf("cannot process access source: %s does not belong to campaign organization", sourceID)
			}

			var err error

			driver, err = s.resolveDriver(ctx, tx, scope, source)
			if err != nil {
				return fmt.Errorf("cannot resolve driver for source %s: %w", source.Name, err)
			}

			lastCompletedCampaign := &coredata.AccessReviewCampaign{}
			if err := lastCompletedCampaign.LoadLastCompletedByOrganizationID(ctx, tx, scope, campaign.OrganizationID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load last completed campaign: %w", err)
				}
			} else {
				entries := &coredata.AccessReviewEntries{}

				baseline, err = entries.LoadBaselineBySourceID(ctx, tx, scope, lastCompletedCampaign.ID, sourceID)
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

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			now := time.Now()
			seenAccountKeys := make(map[string]struct{}, len(accounts))

			for _, account := range accounts {
				accountKey := normalizeAccountKey(account.Email, account.ExternalID)
				seenAccountKeys[accountKey] = struct{}{}

				incrementalTag := coredata.AccessReviewEntryIncrementalTagNew
				if _, ok := previousByAccountKey[accountKey]; ok {
					incrementalTag = coredata.AccessReviewEntryIncrementalTagUnchanged
				}

			entry := &coredata.AccessReviewEntry{
				ID:                           gid.New(scope.GetTenantID(), coredata.AccessReviewEntryEntityType),
				OrganizationID:               campaign.OrganizationID,
				AccessReviewCampaignID:       campaign.ID,
				AccessReviewCampaignSourceID: campaignSource.ID,
				Email:                        account.Email,
				FullName:                     account.FullName,
				Role:                         account.Role,
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
				IncrementalTag:               incrementalTag,
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

			// Create REMOVED entries for accounts that existed in the previous
			// campaign but are no longer present in the current fetch.
			for accountKey, prev := range previousByAccountKey {
				if _, seen := seenAccountKeys[accountKey]; seen {
					continue
				}

				entry := &coredata.AccessReviewEntry{
					ID:                           gid.New(scope.GetTenantID(), coredata.AccessReviewEntryEntityType),
					OrganizationID:               campaign.OrganizationID,
					AccessReviewCampaignID:       campaign.ID,
					AccessReviewCampaignSourceID: campaignSource.ID,
					Email:                        prev.Email,
					FullName:                     prev.FullName,
					AccountKey:                   accountKey,
					IncrementalTag:               coredata.AccessReviewEntryIncrementalTagRemoved,
					Flags:                        []coredata.AccessReviewEntryFlag{},
					FlagReasons:                  []string{},
					Decision:                     coredata.AccessReviewEntryDecisionPending,
					MFAStatus:                    coredata.MFAStatusUnknown,
					AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
					AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
					CreatedAt:                    now,
					UpdatedAt:                    now,
				}

				if err := entry.Upsert(ctx, conn, scope); err != nil {
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
func (s *Service) oauthClient(
	ctx context.Context,
	conn *connector.OAuth2Connection,
	provider coredata.ConnectorProvider,
) (*http.Client, error) {
	if s.connectorRegistry != nil {
		refreshCfg := s.connectorRegistry.GetOAuth2RefreshConfig(string(provider))
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
func (s *Service) connectorHTTPClient(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (*http.Client, error) {
	if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
		return s.oauthClient(ctx, oauth2Conn, dbConnector.Provider)
	}

	return dbConnector.Connection.Client(ctx)
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

	// Capture token before refresh to detect changes.
	var tokenBefore string
	if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
		tokenBefore = oauth2Conn.AccessToken
	}

	// Build an HTTP client. For OAuth2 connections, use RefreshableClient
	// so that short-lived tokens are transparently refreshed.
	httpClient, err := s.connectorHTTPClient(ctx, dbConnector)
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
			if err := dbConnector.Update(ctx, tx, scope, s.encryptionKey); err != nil {
				return nil, fmt.Errorf("cannot persist refreshed token for connector %s: %w", *source.ConnectorID, err)
			}
		}
	}

	reg, ok := s.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.NewDriver == nil {
		return nil, fmt.Errorf("cannot resolve driver: unsupported provider %q", dbConnector.Provider)
	}

	return reg.NewDriver(ctx, httpClient, dbConnector, s.logger)
}
