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
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type grantSnapshot struct {
	EntitlementDisplayName string `json:"entitlement_display_name"`
	Presence               string `json:"presence"`
}

// SyncInventorySource pulls the live roster for one AccessReviewSource and
// upserts principals, resources, and grants. Real grant diffs emit
// AccessInventoryChangeEvent rows (APPEARED / DISAPPEARED / ATTRIBUTE_CHANGED).
// Soft disappear runs only after a successful full roster pull; failed syncs
// leave presence and consecutive_missing_syncs unchanged.
func (s *Service) SyncInventorySource(
	ctx context.Context,
	scope coredata.Scoper,
	run *coredata.AccessSourceSyncRun,
) error {
	var (
		source *coredata.AccessReviewSource
		driver drivers.Driver
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source = &coredata.AccessReviewSource{}
			if err := source.LoadByID(ctx, tx, scope, run.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", run.AccessReviewSourceID, err)
			}

			var err error

			driver, err = s.resolveDriver(ctx, tx, scope, source)
			if err != nil {
				// Commit a refreshed token anyway: rotated refresh tokens are single-use.
				return pg.NoRollback(fmt.Errorf("cannot resolve driver for source %s: %w", source.Name, err))
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	sourceCtx, cancel := context.WithTimeout(ctx, sourceFetchTimeout)
	accounts, err := driver.ListAccounts(sourceCtx)

	cancel()

	if err != nil {
		return fmt.Errorf("cannot list accounts from source %s: %w", source.Name, err)
	}

	// Refresh the lease after ListAccounts so a long provider call does not
	// look stale before the roster write begins.
	if err := s.heartbeatInventorySyncRun(ctx, scope, run); err != nil {
		return fmt.Errorf("cannot heartbeat sync run: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			observedAt := time.Now()

			upserted, err := s.applyInventoryRoster(
				ctx,
				tx,
				scope,
				source,
				run,
				accounts,
				observedAt,
			)
			if err != nil {
				return err
			}

			run.Status = coredata.AccessSourceSyncRunStatusSuccess
			run.GrantsUpsertedCount = upserted
			run.Error = nil
			run.CompletedAt = &observedAt
			run.UpdatedAt = observedAt

			if err := run.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot mark sync run success: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) heartbeatInventorySyncRun(
	ctx context.Context,
	scope coredata.Scoper,
	run *coredata.AccessSourceSyncRun,
) error {
	now := time.Now()
	run.UpdatedAt = now

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return run.Update(ctx, tx, scope)
		},
	)
}

func (s *Service) applyInventoryRoster(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
	run *coredata.AccessSourceSyncRun,
	accounts []drivers.AccountRecord,
	observedAt time.Time,
) (int, error) {
	tenantID := scope.GetTenantID()
	resourceByKey := make(map[string]*coredata.AccessInventoryResource)
	principalByKey := make(map[string]*coredata.AccessInventoryPrincipal)
	seenGrants := make(map[string]struct{})
	upserted := 0

	if _, err := upsertInventoryResource(
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

	for _, account := range accounts {
		triples, err := GrantTriplesFromAccount(account)
		if err != nil {
			if errors.Is(err, ErrMissingPrincipalKey) {
				continue
			}

			return 0, fmt.Errorf("cannot normalise account grants: %w", err)
		}

		principal, err := upsertInventoryPrincipal(
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
			resource, err := upsertInventoryResource(
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

			if err := upsertInventoryGrant(
				ctx,
				tx,
				scope,
				tenantID,
				source,
				run,
				principal,
				resource,
				triple,
				observedAt,
			); err != nil {
				return 0, err
			}

			seenGrants[inventoryGrantNaturalKey(
				triple.PrincipalKey,
				triple.ResourceKey,
				triple.EntitlementKey,
			)] = struct{}{}
			upserted++
		}
	}

	if err := markMissingInventoryGrants(
		ctx,
		tx,
		scope,
		tenantID,
		source,
		run,
		seenGrants,
		observedAt,
	); err != nil {
		return 0, err
	}

	return upserted, nil
}

func inventoryGrantNaturalKey(principalKey, resourceKey, entitlementKey string) string {
	return principalKey + "\x00" + resourceKey + "\x00" + entitlementKey
}

func markMissingInventoryGrants(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	run *coredata.AccessSourceSyncRun,
	seen map[string]struct{},
	observedAt time.Time,
) error {
	return page.WalkAll(
		ctx,
		page.OrderBy[coredata.AccessInventoryGrantOrderField]{
			Field:     coredata.AccessInventoryGrantOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(
			ctx context.Context,
			cursor *page.Cursor[coredata.AccessInventoryGrantOrderField],
		) ([]*coredata.AccessInventoryGrant, error) {
			var batch coredata.AccessInventoryGrants
			if err := batch.LoadBySourceID(
				ctx,
				tx,
				scope,
				source.ID,
				cursor,
			); err != nil {
				return nil, err
			}

			return batch, nil
		},
		func(grants []*coredata.AccessInventoryGrant) error {
			for _, grant := range grants {
				key := inventoryGrantNaturalKey(
					grant.PrincipalKey,
					grant.ResourceKey,
					grant.EntitlementKey,
				)
				if _, ok := seen[key]; ok {
					continue
				}

				if err := disappearInventoryGrant(
					ctx,
					tx,
					scope,
					tenantID,
					source,
					run,
					grant,
					observedAt,
				); err != nil {
					return err
				}
			}

			return nil
		},
	)
}

func disappearInventoryGrant(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	run *coredata.AccessSourceSyncRun,
	grant *coredata.AccessInventoryGrant,
	observedAt time.Time,
) error {
	wasPresent := grant.Presence == coredata.AccessInventoryGrantPresencePresent
	priorDisplayName := grant.EntitlementDisplayName
	priorPresence := grant.Presence

	grant.Presence = coredata.AccessInventoryGrantPresenceDisappeared
	grant.ConsecutiveMissingSyncs++
	grant.UpdatedAt = observedAt

	if err := grant.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot mark inventory grant disappeared: %w", err)
	}

	if !wasPresent {
		return nil
	}

	before, err := marshalGrantSnapshot(
		&coredata.AccessInventoryGrant{
			EntitlementDisplayName: priorDisplayName,
			Presence:               priorPresence,
		},
	)
	if err != nil {
		return err
	}

	return insertInventoryChangeEvent(
		ctx,
		tx,
		scope,
		tenantID,
		source,
		run,
		grant,
		coredata.AccessInventoryChangeTypeDisappeared,
		observedAt,
		before,
		nil,
		"disappeared",
	)
}

func upsertInventoryPrincipal(
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
		return nil, fmt.Errorf("cannot upsert inventory principal: %w", err)
	}

	cache[principalKey] = principal

	return principal, nil
}

func upsertInventoryResource(
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
		return nil, fmt.Errorf("cannot upsert inventory resource: %w", err)
	}

	cache[triple.ResourceKey] = resource

	return resource, nil
}

func upsertInventoryGrant(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	run *coredata.AccessSourceSyncRun,
	principal *coredata.AccessInventoryPrincipal,
	resource *coredata.AccessInventoryResource,
	triple GrantTriple,
	observedAt time.Time,
) error {
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

	var (
		prior            *coredata.AccessInventoryGrant
		appeared         bool
		attributeChanged bool
		priorDisplayName string
		priorPresence    coredata.AccessInventoryGrantPresence
	)

	switch {
	case err == nil:
		prior = existing
		priorDisplayName = existing.EntitlementDisplayName
		priorPresence = existing.Presence
		appeared = existing.Presence == coredata.AccessInventoryGrantPresenceDisappeared
		attributeChanged = existing.EntitlementDisplayName != triple.EntitlementDisplayName
	case errors.Is(err, coredata.ErrResourceNotFound):
		appeared = true
	default:
		return fmt.Errorf("cannot load inventory grant: %w", err)
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

	if prior != nil {
		grant.ID = prior.ID
		grant.FirstSeenAt = prior.FirstSeenAt
		grant.CreatedAt = prior.CreatedAt
		grant.ProviderChangedAt = prior.ProviderChangedAt
	}

	if _, err := grant.Upsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot upsert inventory grant: %w", err)
	}

	if appeared {
		after, err := marshalGrantSnapshot(grant)
		if err != nil {
			return err
		}

		return insertInventoryChangeEvent(
			ctx,
			tx,
			scope,
			tenantID,
			source,
			run,
			grant,
			coredata.AccessInventoryChangeTypeAppeared,
			observedAt,
			nil,
			after,
			"appeared",
		)
	}

	if attributeChanged {
		beforePayload, err := marshalGrantSnapshot(
			&coredata.AccessInventoryGrant{
				EntitlementDisplayName: priorDisplayName,
				Presence:               priorPresence,
			},
		)
		if err != nil {
			return err
		}

		after, err := marshalGrantSnapshot(grant)
		if err != nil {
			return err
		}

		fieldHash := fmt.Sprintf(
			"entitlement_display_name:%s→%s",
			priorDisplayName,
			triple.EntitlementDisplayName,
		)

		return insertInventoryChangeEvent(
			ctx,
			tx,
			scope,
			tenantID,
			source,
			run,
			grant,
			coredata.AccessInventoryChangeTypeAttributeChanged,
			observedAt,
			beforePayload,
			after,
			fieldHash,
		)
	}

	return nil
}

func marshalGrantSnapshot(grant *coredata.AccessInventoryGrant) (*json.RawMessage, error) {
	payload, err := json.Marshal(
		grantSnapshot{
			EntitlementDisplayName: grant.EntitlementDisplayName,
			Presence:               grant.Presence.String(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal grant snapshot: %w", err)
	}

	raw := json.RawMessage(payload)

	return &raw, nil
}

func insertInventoryChangeEvent(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tenantID gid.TenantID,
	source *coredata.AccessReviewSource,
	run *coredata.AccessSourceSyncRun,
	grant *coredata.AccessInventoryGrant,
	changeType coredata.AccessInventoryChangeType,
	observedAt time.Time,
	before *json.RawMessage,
	after *json.RawMessage,
	fieldHash string,
) error {
	syncRunID := run.ID
	grantID := grant.ID
	idempotencyKey := inventoryChangeIdempotencyKey(
		run.ID.String(),
		grant.PrincipalKey,
		grant.ResourceKey,
		grant.EntitlementKey,
		changeType.String(),
		fieldHash,
	)

	event := &coredata.AccessInventoryChangeEvent{
		ID:                   gid.New(tenantID, coredata.AccessInventoryChangeEventEntityType),
		OrganizationID:       source.OrganizationID,
		AccessReviewSourceID: source.ID,
		GrantID:              &grantID,
		PrincipalKey:         grant.PrincipalKey,
		ResourceKey:          grant.ResourceKey,
		EntitlementKey:       grant.EntitlementKey,
		ChangeType:           changeType,
		ObservedAt:           observedAt,
		SyncRunID:            &syncRunID,
		Before:               before,
		After:                after,
		IdempotencyKey:       idempotencyKey,
		CreatedAt:            observedAt,
	}

	if _, err := event.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert inventory change event: %w", err)
	}

	return nil
}

// inventoryChangeIdempotencyKey hashes a length-prefixed tuple so provider
// keys that contain delimiters cannot collide with another encoding.
func inventoryChangeIdempotencyKey(parts ...string) string {
	h := sha256.New()

	for _, part := range parts {
		var length [4]byte

		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = h.Write(length[:])
		_, _ = h.Write([]byte(part))
	}

	return hex.EncodeToString(h.Sum(nil))
}
