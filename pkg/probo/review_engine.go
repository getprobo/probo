// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

	err := e.pg.WithTx(ctx, func(conn pg.Conn) error {
		previousByAccountKey := map[string]struct{}{}
		lastCompletedCampaign := &coredata.AccessReviewCampaign{}
		if err := lastCompletedCampaign.LoadLastCompletedByAccessReviewID(ctx, conn, e.scope, campaign.AccessReviewID); err != nil {
			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load last completed campaign: %w", err)
			}
		} else {
			rows, err := conn.Query(ctx, `
SELECT account_key
FROM access_entries
WHERE access_review_campaign_id = $1
  AND access_source_id = $2
`, lastCompletedCampaign.ID, sourceID)
			if err != nil {
				return fmt.Errorf("cannot load baseline entries: %w", err)
			}
			defer rows.Close()
			for rows.Next() {
				var accountKey string
				if err := rows.Scan(&accountKey); err != nil {
					return fmt.Errorf("cannot scan baseline entry: %w", err)
				}
				previousByAccountKey[accountKey] = struct{}{}
			}
			if err := rows.Err(); err != nil {
				return fmt.Errorf("cannot iterate baseline entries: %w", err)
			}
		}

		source := &coredata.AccessSource{}
		if err := source.LoadByID(ctx, conn, e.scope, sourceID); err != nil {
			return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
		}
		if source.AccessReviewID != campaign.AccessReviewID {
			return fmt.Errorf("access source %s does not belong to campaign access review", sourceID)
		}

		driver, err := e.resolveDriver(ctx, conn, source)
		if err != nil {
			return fmt.Errorf("cannot resolve driver for source %s: %w", source.Name, err)
		}

		sourceCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		accounts, err := driver.ListAccounts(sourceCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("cannot list accounts from source %s: %w", source.Name, err)
		}
		fetchedCount = len(accounts)

		now := time.Now()
		for _, account := range accounts {
			accountKey := normalizeAccountKey(account.Email, account.ExternalID)
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

			_, err = conn.Exec(ctx, `
INSERT INTO access_entries (
    id,
    tenant_id,
    access_review_campaign_id,
    access_source_id,
    identity_id,
    email,
    full_name,
    role,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flag,
    flag_reason,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at,
    updated_at
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25
)
ON CONFLICT (access_review_campaign_id, access_source_id, account_key) DO UPDATE SET
    email = EXCLUDED.email,
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    job_title = EXCLUDED.job_title,
    is_admin = EXCLUDED.is_admin,
    mfa_status = EXCLUDED.mfa_status,
    auth_method = EXCLUDED.auth_method,
    last_login = EXCLUDED.last_login,
    account_created_at = EXCLUDED.account_created_at,
    external_id = EXCLUDED.external_id,
    incremental_tag = EXCLUDED.incremental_tag,
    updated_at = EXCLUDED.updated_at
`,
				entry.ID,
				e.scope.GetTenantID(),
				entry.AccessReviewCampaignID,
				entry.AccessSourceID,
				entry.IdentityID,
				entry.Email,
				entry.FullName,
				entry.Role,
				entry.JobTitle,
				entry.IsAdmin,
				entry.MFAStatus,
				entry.AuthMethod,
				entry.LastLogin,
				entry.AccountCreatedAt,
				entry.ExternalID,
				entry.AccountKey,
				entry.IncrementalTag,
				entry.Flag,
				entry.FlagReason,
				entry.Decision,
				entry.DecisionNote,
				entry.DecidedBy,
				entry.DecidedAt,
				entry.CreatedAt,
				entry.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("cannot upsert access entry: %w", err)
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
		review := &coredata.AccessReview{}
		if err := review.LoadByID(ctx, conn, e.scope, source.AccessReviewID); err != nil {
			return nil, fmt.Errorf("cannot load access review: %w", err)
		}

		return accesssource.NewProboMembershipsDriver(e.pg, e.scope, review.OrganizationID), nil
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
		baseURL, _ := dbConnector.Settings["scim_bridge_url"].(string)
		if baseURL == "" {
			return nil, fmt.Errorf("1password connector requires scim_bridge_url in settings")
		}
		return accesssource.NewOnePasswordDriver(httpClient, baseURL), nil
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
		organizationID, _ := dbConnector.Settings["organization_id"].(string)
		if organizationID == "" {
			return nil, fmt.Errorf("tally connector requires organization_id in settings")
		}
		return accesssource.NewTallyDriver(httpClient, organizationID), nil
	default:
		return nil, fmt.Errorf("unsupported connector provider %q for access source driver", dbConnector.Provider)
	}
}
