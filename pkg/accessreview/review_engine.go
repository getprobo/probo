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
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// sourceFetchTimeout bounds one source's account listing, however many
// paginated calls the driver makes to complete it.
const sourceFetchTimeout = 30 * time.Second

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

	// Load baseline data first and fetch accounts afterwards, so the
	// third-party call never holds a database connection.
	var (
		source   *coredata.AccessReviewSource
		baseline []coredata.BaselineAccountEntry
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source = &coredata.AccessReviewSource{}
			if err := source.LoadByID(ctx, tx, scope, sourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
			}

			lastCompletedCampaign := &coredata.AccessReviewCampaign{}
			if err := lastCompletedCampaign.LoadLastCompletedByOrganizationID(ctx, tx, scope, campaign.OrganizationID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load last completed campaign: %w", err)
				}

				return nil
			}

			entries := &coredata.AccessReviewEntries{}

			loaded, err := entries.LoadBaselineBySourceID(ctx, tx, scope, lastCompletedCampaign.ID, sourceID)
			if err != nil {
				return fmt.Errorf("cannot load baseline entries by source: %w", err)
			}

			baseline = loaded

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

	accounts, err := s.listAccounts(ctx, scope, source)
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

// listAccounts pulls the accounts an access source exposes. A connector-backed
// source runs inside Opener.Use, so its credential lives no longer than the
// fetch; a built-in source needs no credential at all.
func (s *Service) listAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) ([]drivers.AccountRecord, error) {
	if source.ConnectorID == nil {
		return listAccountsWithTimeout(ctx, s.builtinDriver(scope, source))
	}

	var accounts []drivers.AccountRecord

	err := s.connectors.Use(
		ctx,
		scope,
		*source.ConnectorID,
		func(ctx context.Context, handle *provider.Handle) error {
			driver, err := s.sources.New(ctx, handle, s.logger)
			if err != nil {
				return err
			}

			accounts, err = listAccountsWithTimeout(ctx, driver)

			return err
		},
	)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// builtinDriver serves a source with no connector behind it: pasted CSV, or the
// organization's own Probo memberships.
func (s *Service) builtinDriver(
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) drivers.Driver {
	if source.CsvData != nil && *source.CsvData != "" {
		return drivers.NewCSVDriver(strings.NewReader(*source.CsvData))
	}

	return drivers.NewProboMembershipsDriver(s.pg, scope, source.OrganizationID)
}

func listAccountsWithTimeout(
	ctx context.Context,
	driver drivers.Driver,
) ([]drivers.AccountRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, sourceFetchTimeout)
	defer cancel()

	return driver.ListAccounts(ctx)
}
