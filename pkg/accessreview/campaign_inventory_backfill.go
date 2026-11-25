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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// BackfillInventoryFromLatestCampaignEntries maps the latest successful
// campaign entries for each live source in the organization into live
// inventory principals and coarse grants on the default resource.
//
// Sources that already have any SUCCESS inventory sync run are skipped —
// dual-write or continuous sync already seeded them (a later FAILED run must
// not re-seed stale campaign grants). Existing inventory rows are left
// untouched (insert-only). No change events or sync runs are created.
func (s *Service) BackfillInventoryFromLatestCampaignEntries(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	upserted := 0

	err := page.WalkAll(
		ctx,
		page.OrderBy[coredata.AccessReviewSourceOrderField]{
			Field:     coredata.AccessReviewSourceOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(
			ctx context.Context,
			cursor *page.Cursor[coredata.AccessReviewSourceOrderField],
		) ([]*coredata.AccessReviewSource, error) {
			var batch coredata.AccessReviewSources
			err := s.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return batch.LoadByOrganizationID(
						ctx,
						conn,
						scope,
						organizationID,
						cursor,
					)
				},
			)
			if err != nil {
				return nil, err
			}

			return batch, nil
		},
		func(sources []*coredata.AccessReviewSource) error {
			for _, source := range sources {
				count, err := s.backfillInventoryForSourceFromCampaign(
					ctx,
					scope,
					source,
				)
				if err != nil {
					// One sticky source must not block the rest of the org on
					// every Service.Run — match per-org continue in BackfillAll.
					s.logger.ErrorCtx(
						ctx,
						"cannot backfill inventory for access source",
						log.String("organization_id", organizationID.String()),
						log.String("access_review_source_id", source.ID.String()),
						log.Error(err),
					)

					continue
				}

				upserted += count
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot backfill inventory from campaign entries: %w", err)
	}

	return upserted, nil
}

func (s *Service) backfillInventoryForSourceFromCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (int, error) {
	var (
		skip     bool
		campaign *coredata.AccessReviewCampaignSource
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			successful := &coredata.AccessSourceSyncRun{}
			err := successful.LoadLatestSuccessfulBySourceID(
				ctx,
				conn,
				scope,
				source.ID,
			)
			switch {
			case err == nil:
				skip = true

				return nil
			case errors.Is(err, coredata.ErrResourceNotFound):
				// No successful sync yet — campaign entries may still seed
				// inventory.
			default:
				return fmt.Errorf("cannot load successful sync run: %w", err)
			}

			campaign = &coredata.AccessReviewCampaignSource{}
			err = campaign.LoadLatestSuccessfulByAccessReviewSourceID(
				ctx,
				conn,
				scope,
				source.ID,
			)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				skip = true

				return nil
			}
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	if skip {
		return 0, nil
	}

	observedAt := time.Now()
	upserted := 0

	err = page.WalkAll(
		ctx,
		page.OrderBy[coredata.AccessReviewEntryOrderField]{
			Field:     coredata.AccessReviewEntryOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(
			ctx context.Context,
			cursor *page.Cursor[coredata.AccessReviewEntryOrderField],
		) ([]*coredata.AccessReviewEntry, error) {
			var batch coredata.AccessReviewEntries
			err := s.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return batch.LoadByCampaignIDAndSourceID(
						ctx,
						conn,
						scope,
						campaign.AccessReviewCampaignID,
						campaign.ID,
						cursor,
						&coredata.AccessReviewEntryFilter{},
					)
				},
			)
			if err != nil {
				return nil, err
			}

			return batch, nil
		},
		func(entries []*coredata.AccessReviewEntry) error {
			return s.pg.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					locked := &coredata.AccessReviewSource{}
					if err := locked.LoadByIDForUpdate(ctx, tx, scope, source.ID); err != nil {
						return fmt.Errorf("cannot lock access source: %w", err)
					}

					successful := &coredata.AccessSourceSyncRun{}
					err := successful.LoadLatestSuccessfulBySourceID(
						ctx,
						tx,
						scope,
						source.ID,
					)
					switch {
					case err == nil:
						return nil
					case errors.Is(err, coredata.ErrResourceNotFound):
						// Continue backfill.
					default:
						return fmt.Errorf("cannot load successful sync run: %w", err)
					}

					count, err := s.applyCampaignEntryInventoryBackfill(
						ctx,
						tx,
						scope,
						locked,
						entries,
						observedAt,
					)
					if err != nil {
						return err
					}

					upserted += count

					return nil
				},
			)
		},
	)
	if err != nil {
		return 0, fmt.Errorf(
			"cannot backfill inventory for source %s: %w",
			source.ID,
			err,
		)
	}

	return upserted, nil
}

// BackfillAllInventoryFromLatestCampaignEntries runs the org-scoped backfill
// for every organization that has a successful live-linked campaign fetch.
// Safe to re-run: insert-only and skips sources with any SUCCESS sync. Per-org
// failures are logged and skipped so one bad tenant cannot abort the rest.
func (s *Service) BackfillAllInventoryFromLatestCampaignEntries(
	ctx context.Context,
) (int, error) {
	var organizationIDs []gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var sources coredata.AccessReviewCampaignSources
			ids, err := sources.LoadOrganizationIDsWithSuccessfulLiveFetches(
				ctx,
				conn,
				coredata.NewNoScope(),
			)
			if err != nil {
				return err
			}

			organizationIDs = ids

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot list organizations for campaign backfill: %w", err)
	}

	upserted := 0
	for _, organizationID := range organizationIDs {
		scope := coredata.NewScopeFromObjectID(organizationID)
		count, err := s.BackfillInventoryFromLatestCampaignEntries(
			ctx,
			scope,
			organizationID,
		)
		if err != nil {
			s.logger.ErrorCtx(
				ctx,
				"cannot backfill inventory from campaign entries",
				log.String("organization_id", organizationID.String()),
				log.Error(err),
			)

			continue
		}

		upserted += count
	}

	return upserted, nil
}

// applyCampaignEntryInventoryBackfill inserts missing principals / the default
// resource / grants from campaign entries. It never emits change events and
// never updates rows that already exist in live inventory.
func (s *Service) applyCampaignEntryInventoryBackfill(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
	entries []*coredata.AccessReviewEntry,
	observedAt time.Time,
) (int, error) {
	tenantID := scope.GetTenantID()
	resourceByKey := make(map[string]*coredata.AccessInventoryResource)
	principalByKey := make(map[string]*coredata.AccessInventoryPrincipal)
	upserted := 0

	if _, err := ensureBackfillInventoryResource(
		ctx,
		tx,
		scope,
		tenantID,
		source,
		GrantTriple{
			ResourceKey:  DefaultResourceKey,
			ResourceKind: DefaultResourceKind,
		},
		observedAt,
		resourceByKey,
	); err != nil {
		return 0, err
	}

	for _, entry := range entries {
		account := accountRecordFromCampaignEntry(entry)

		triples, err := GrantTriplesFromAccount(account)
		if err != nil {
			if errors.Is(err, ErrMissingPrincipalKey) {
				continue
			}

			return 0, fmt.Errorf("cannot normalise campaign entry grants: %w", err)
		}

		principal, err := ensureBackfillInventoryPrincipal(
			ctx,
			tx,
			scope,
			tenantID,
			source,
			account,
			observedAt,
			principalByKey,
		)
		if err != nil {
			return 0, err
		}

		for _, triple := range triples {
			resource, err := ensureBackfillInventoryResource(
				ctx,
				tx,
				scope,
				tenantID,
				source,
				triple,
				observedAt,
				resourceByKey,
			)
			if err != nil {
				return 0, err
			}

			inserted, err := insertBackfillInventoryGrantIfAbsent(
				ctx,
				tx,
				scope,
				tenantID,
				source,
				principal,
				resource,
				triple,
				observedAt,
			)
			if err != nil {
				return 0, err
			}

			if inserted {
				upserted++
			}
		}
	}

	return upserted, nil
}

func accountRecordFromCampaignEntry(entry *coredata.AccessReviewEntry) drivers.AccountRecord {
	return drivers.AccountRecord{
		Email:       entry.Email,
		FullName:    entry.FullName,
		Roles:       entry.Roles,
		JobTitle:    entry.JobTitle,
		Active:      entry.Active,
		IsAdmin:     entry.IsAdmin,
		MFAStatus:   entry.MFAStatus,
		AuthMethod:  entry.AuthMethod,
		AccountType: entry.AccountType,
		LastLogin:   entry.LastLogin,
		CreatedAt:   entry.AccountCreatedAt,
		ExternalID:  entry.ExternalID,
	}
}

func ensureBackfillInventoryPrincipal(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	account drivers.AccountRecord,
	observedAt time.Time,
	cache map[string]*coredata.AccessInventoryPrincipal,
) (*coredata.AccessInventoryPrincipal, error) {
	principalKey := normalizeAccountKey(account.Email, account.ExternalID)
	if cached, ok := cache[principalKey]; ok {
		return cached, nil
	}

	existing := &coredata.AccessInventoryPrincipal{}
	err := existing.LoadBySourceIDAndPrincipalKey(
		ctx,
		tx,
		scope,
		source.ID,
		principalKey,
	)
	switch {
	case err == nil:
		cache[principalKey] = existing

		return existing, nil
	case errors.Is(err, coredata.ErrResourceNotFound):
		// Insert below.
	default:
		return nil, fmt.Errorf("cannot load inventory principal: %w", err)
	}

	principal := &coredata.AccessInventoryPrincipal{
		ID:                   gid.New(tenantID, coredata.AccessInventoryPrincipalEntityType),
		OrganizationID:       source.OrganizationID,
		AccessReviewSourceID: source.ID,
		PrincipalKey:         principalKey,
		Email:                account.Email,
		FullName:             account.FullName,
		ExternalID:           account.ExternalID,
		JobTitle:             account.JobTitle,
		IsAdmin:              account.IsAdmin,
		MFAStatus:            account.MFAStatus,
		AuthMethod:           account.AuthMethod,
		AccountType:          account.AccountType,
		Active:               account.Active,
		LastLogin:            account.LastLogin,
		AccountCreatedAt:     account.CreatedAt,
		FirstSeenAt:          observedAt,
		LastSeenAt:           observedAt,
		CreatedAt:            observedAt,
		UpdatedAt:            observedAt,
	}

	if _, err := principal.Upsert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert backfill inventory principal: %w", err)
	}

	cache[principalKey] = principal

	return principal, nil
}

func ensureBackfillInventoryResource(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	triple GrantTriple,
	observedAt time.Time,
	cache map[string]*coredata.AccessInventoryResource,
) (*coredata.AccessInventoryResource, error) {
	if cached, ok := cache[triple.ResourceKey]; ok {
		return cached, nil
	}

	existing := &coredata.AccessInventoryResource{}
	err := existing.LoadBySourceIDAndResourceKey(
		ctx,
		tx,
		scope,
		source.ID,
		triple.ResourceKey,
	)
	switch {
	case err == nil:
		cache[triple.ResourceKey] = existing

		return existing, nil
	case errors.Is(err, coredata.ErrResourceNotFound):
		// Insert below.
	default:
		return nil, fmt.Errorf("cannot load inventory resource: %w", err)
	}

	resource := &coredata.AccessInventoryResource{
		ID:                   gid.New(tenantID, coredata.AccessInventoryResourceEntityType),
		OrganizationID:       source.OrganizationID,
		AccessReviewSourceID: source.ID,
		ResourceKey:          triple.ResourceKey,
		DisplayName:          source.Name,
		Kind:                 triple.ResourceKind,
		FirstSeenAt:          observedAt,
		LastSeenAt:           observedAt,
		CreatedAt:            observedAt,
		UpdatedAt:            observedAt,
	}

	if _, err := resource.Upsert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert backfill inventory resource: %w", err)
	}

	cache[triple.ResourceKey] = resource

	return resource, nil
}

func insertBackfillInventoryGrantIfAbsent(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	principal *coredata.AccessInventoryPrincipal,
	resource *coredata.AccessInventoryResource,
	triple GrantTriple,
	observedAt time.Time,
) (bool, error) {
	existing := &coredata.AccessInventoryGrant{}
	err := existing.LoadByNaturalKey(
		ctx,
		tx,
		scope,
		source.ID,
		triple.PrincipalKey,
		triple.ResourceKey,
		triple.EntitlementKey,
	)
	switch {
	case err == nil:
		return false, nil
	case errors.Is(err, coredata.ErrResourceNotFound):
		// Insert below.
	default:
		return false, fmt.Errorf("cannot load inventory grant: %w", err)
	}

	grant := &coredata.AccessInventoryGrant{
		ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
		OrganizationID:          source.OrganizationID,
		AccessReviewSourceID:    source.ID,
		PrincipalID:             principal.ID,
		ResourceID:              resource.ID,
		PrincipalKey:            triple.PrincipalKey,
		ResourceKey:             triple.ResourceKey,
		EntitlementKey:          triple.EntitlementKey,
		EntitlementDisplayName:  triple.EntitlementDisplayName,
		Presence:                coredata.AccessInventoryGrantPresencePresent,
		ConsecutiveMissingSyncs: 0,
		FirstSeenAt:             observedAt,
		LastSeenAt:              observedAt,
		CreatedAt:               observedAt,
		UpdatedAt:               observedAt,
	}

	if _, err := grant.Upsert(ctx, tx, scope); err != nil {
		return false, fmt.Errorf("cannot insert backfill inventory grant: %w", err)
	}

	return true, nil
}
