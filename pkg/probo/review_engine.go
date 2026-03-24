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

package probo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo/accesssource"
)

// ReviewEngine contains the stateless core logic for access review campaigns:
// snapshot and source data collection.
type ReviewEngine struct {
	pg            *pg.Client
	scope         coredata.Scoper
	encryptionKey cipher.EncryptionKey
}

func NewReviewEngine(pgClient *pg.Client, scope coredata.Scoper, encryptionKey cipher.EncryptionKey) *ReviewEngine {
	return &ReviewEngine{
		pg:            pgClient,
		scope:         scope,
		encryptionKey: encryptionKey,
	}
}

// SnapshotSource pulls accounts from a single source and upserts access entries.
func (e *ReviewEngine) SnapshotSource(
	ctx context.Context,
	campaign *coredata.AccessReviewCampaign,
	sourceID gid.GID,
) (int, error) {
	fetchedCount := 0

	// Resolve the driver and load baseline data outside the write transaction
	// so that external HTTP calls do not hold a database connection.
	var (
		source   *coredata.AccessSource
		driver   accesssource.Driver
		baseline []coredata.BaselineAccountEntry
	)

	err := e.pg.WithConn(ctx, func(conn pg.Conn) error {
		source = &coredata.AccessSource{}
		if err := source.LoadByID(ctx, conn, e.scope, sourceID); err != nil {
			return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
		}
		if source.OrganizationID != campaign.OrganizationID {
			return fmt.Errorf("cannot process access source: %s does not belong to campaign organization", sourceID)
		}

		var err error
		driver, err = e.resolveDriver(ctx, conn, source)
		if err != nil {
			return fmt.Errorf("cannot resolve driver for source %s: %w", source.Name, err)
		}

		lastCompletedCampaign := &coredata.AccessReviewCampaign{}
		if err := lastCompletedCampaign.LoadLastCompletedByOrganizationID(ctx, conn, e.scope, campaign.OrganizationID); err != nil {
			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load last completed campaign: %w", err)
			}
		} else {
			entries := &coredata.AccessEntries{}
			baseline, err = entries.LoadBaselineBySourceID(ctx, conn, e.scope, lastCompletedCampaign.ID, sourceID)
			if err != nil {
				return err
			}
		}

		return nil
	})
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

	err = e.pg.WithTx(ctx, func(conn pg.Conn) error {
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
				ID:                     gid.New(e.scope.GetTenantID(), coredata.AccessEntryEntityType),
				AccessReviewCampaignID: campaign.ID,
				AccessSourceID:         sourceID,
				Email:                  account.Email,
				FullName:               account.FullName,
				Role:                   account.Role,
				JobTitle:               account.JobTitle,
				IsAdmin:                account.IsAdmin,
				MFAStatus:              account.MFAStatus,
				AuthMethod:             account.AuthMethod,
				LastLogin:              account.LastLogin,
				AccountCreatedAt:       account.CreatedAt,
				ExternalID:             account.ExternalID,
				AccountKey:             accountKey,
				IncrementalTag:         incrementalTag,
				Flag:                   coredata.AccessEntryFlagNone,
				Decision:               coredata.AccessEntryDecisionPending,
				CreatedAt:              now,
				UpdatedAt:              now,
			}

			if err := entry.Upsert(ctx, conn, e.scope); err != nil {
				return err
			}
		}

		// Create REMOVED entries for accounts that existed in the previous
		// campaign but are no longer present in the current fetch.
		for accountKey, prev := range previousByAccountKey {
			if _, seen := seenAccountKeys[accountKey]; seen {
				continue
			}

			entry := &coredata.AccessEntry{
				ID:                     gid.New(e.scope.GetTenantID(), coredata.AccessEntryEntityType),
				AccessReviewCampaignID: campaign.ID,
				AccessSourceID:         sourceID,
				Email:                  prev.Email,
				FullName:               prev.FullName,
				AccountKey:             accountKey,
				IncrementalTag:         coredata.AccessEntryIncrementalTagRemoved,
				Flag:                   coredata.AccessEntryFlagNone,
				Decision:               coredata.AccessEntryDecisionPending,
				MFAStatus:              coredata.MFAStatusUnknown,
				AuthMethod:             coredata.AccessEntryAuthMethodUnknown,
				CreatedAt:              now,
				UpdatedAt:              now,
			}

			if err := entry.Upsert(ctx, conn, e.scope); err != nil {
				return err
			}
		}

		return nil
	})
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

// resolveDriver creates a Driver for the given AccessSource based on
// connector_id (null = built-in, set = connector-backed).
func (e *ReviewEngine) resolveDriver(
	ctx context.Context,
	conn pg.Conn,
	source *coredata.AccessSource,
) (accesssource.Driver, error) {
	if source.ConnectorID == nil {
		// CSV-backed source: use CSVDriver when csv_data is present
		if source.CsvData != nil && *source.CsvData != "" {
			return accesssource.NewCSVDriver(strings.NewReader(*source.CsvData)), nil
		}

		// Built-in driver: default to ProboMemberships
		return accesssource.NewProboMembershipsDriver(e.pg, e.scope, source.OrganizationID), nil
	}

	// Connector-backed: look up the connector and resolve driver by provider
	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByID(ctx, conn, e.scope, *source.ConnectorID, e.encryptionKey); err != nil {
		return nil, fmt.Errorf("cannot load connector %s: %w", *source.ConnectorID, err)
	}

	switch dbConnector.Provider {
	case coredata.ConnectorProviderGoogleWorkspace:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for google workspace connector: %w", err)
		}
		return accesssource.NewGoogleWorkspaceDriver(httpClient), nil
	case coredata.ConnectorProviderLinear:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for linear connector: %w", err)
		}
		return accesssource.NewLinearDriver(httpClient), nil
	case coredata.ConnectorProviderSlack:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for slack connector: %w", err)
		}
		return accesssource.NewSlackDriver(httpClient), nil
	case coredata.ConnectorProviderFigma:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for figma connector: %w", err)
		}
		return accesssource.NewFigmaDriver(httpClient), nil
	case coredata.ConnectorProviderOnePassword:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for 1password connector: %w", err)
		}
		onePasswordSettings, err := dbConnector.OnePasswordSettings()
		if err != nil {
			return nil, fmt.Errorf("cannot read 1password connector settings: %w", err)
		}
		if onePasswordSettings.SCIMBridgeURL == "" {
			return nil, fmt.Errorf("1password connector requires scim_bridge_url in settings")
		}
		return accesssource.NewOnePasswordDriver(httpClient, onePasswordSettings.SCIMBridgeURL), nil
	case coredata.ConnectorProviderHubSpot:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for hubspot connector: %w", err)
		}
		return accesssource.NewHubSpotDriver(httpClient), nil
	case coredata.ConnectorProviderDocuSign:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for docusign connector: %w", err)
		}
		return accesssource.NewDocuSignDriver(httpClient), nil
	case coredata.ConnectorProviderNotion:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for notion connector: %w", err)
		}
		return accesssource.NewNotionDriver(httpClient), nil
	case coredata.ConnectorProviderBrex:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for brex connector: %w", err)
		}
		return accesssource.NewBrexDriver(httpClient), nil
	case coredata.ConnectorProviderTally:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for tally connector: %w", err)
		}
		tallySettings, err := dbConnector.TallySettings()
		if err != nil {
			return nil, fmt.Errorf("cannot read tally connector settings: %w", err)
		}
		if tallySettings.OrganizationID == "" {
			return nil, fmt.Errorf("tally connector requires organization_id in settings")
		}
		return accesssource.NewTallyDriver(httpClient, tallySettings.OrganizationID), nil
	case coredata.ConnectorProviderCloudflare:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for cloudflare connector: %w", err)
		}
		return accesssource.NewCloudflareDriver(httpClient), nil
	case coredata.ConnectorProviderOpenAI:
		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client for openai connector: %w", err)
		}
		return accesssource.NewOpenAIDriver(httpClient), nil
	default:
		return nil, fmt.Errorf("unsupported connector provider %q for access source driver", dbConnector.Provider)
	}
}
