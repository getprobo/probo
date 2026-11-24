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
	"math/rand/v2"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

const (
	inventorySyncFailureBaseBackoff = time.Minute
	inventorySyncFailureMaxBackoff  = time.Hour

	// ±20% spread so many orgs on the same cadence do not hit providers together.
	inventorySyncJitterFraction = 5
)

func (s *Service) GetAccessSourceSyncRun(
	ctx context.Context,
	scope coredata.Scoper,
	runID gid.GID,
) (*coredata.AccessSourceSyncRun, error) {
	run := &coredata.AccessSourceSyncRun{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadByID(ctx, conn, scope, runID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get access source sync run: %w", err)
	}

	return run, nil
}

func (s *Service) GetLatestAccessSourceSyncRunForSourceID(
	ctx context.Context,
	scope coredata.Scoper,
	sourceID gid.GID,
) (*coredata.AccessSourceSyncRun, error) {
	run := &coredata.AccessSourceSyncRun{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadLatestBySourceID(ctx, conn, scope, sourceID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get latest access source sync run: %w", err)
	}

	return run, nil
}

// SyncNow enqueues an AccessSourceSyncRun for the source. If a QUEUED or
// RUNNING run already exists, that run is returned unchanged.
func (s *Service) SyncNow(
	ctx context.Context,
	scope coredata.Scoper,
	sourceID gid.GID,
) (*coredata.AccessSourceSyncRun, error) {
	v := validator.New()
	v.Check(sourceID, "access_review_source_id", validator.Required(), validator.GID(coredata.AccessReviewSourceEntityType))

	if err := v.Error(); err != nil {
		return nil, err
	}

	var run *coredata.AccessSourceSyncRun

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source := &coredata.AccessReviewSource{}
			if err := source.LoadByIDForUpdate(ctx, tx, scope, sourceID); err != nil {
				return fmt.Errorf("cannot load access source: %w", err)
			}

			enqueued, err := s.enqueueInventorySyncRun(ctx, tx, scope, source)
			if err != nil {
				return err
			}

			run = enqueued

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot sync now: %w", err)
	}

	return run, nil
}

func (s *Service) enqueueInventorySyncRun(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (*coredata.AccessSourceSyncRun, error) {
	existing := &coredata.AccessSourceSyncRun{}

	err := existing.LoadInFlightBySourceID(ctx, tx, scope, source.ID)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load in-flight sync run: %w", err)
	}

	return s.insertQueuedInventorySyncRun(ctx, tx, scope, source)
}

// enqueueInventorySyncRunAfterCSVUpdate always inserts a QUEUED run so a
// re-upload during RUNNING is not lost (driver may already hold stale CSV).
func (s *Service) enqueueInventorySyncRunAfterCSVUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (*coredata.AccessSourceSyncRun, error) {
	return s.insertQueuedInventorySyncRun(ctx, tx, scope, source)
}

func (s *Service) insertQueuedInventorySyncRun(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (*coredata.AccessSourceSyncRun, error) {
	now := time.Now()
	run := &coredata.AccessSourceSyncRun{
		ID:                   gid.New(scope.GetTenantID(), coredata.AccessSourceSyncRunEntityType),
		OrganizationID:       source.OrganizationID,
		AccessReviewSourceID: source.ID,
		Status:               coredata.AccessSourceSyncRunStatusQueued,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := run.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert sync run: %w", err)
	}

	return run, nil
}

func (s *Service) resolveSourceSyncCadence(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (time.Duration, error) {
	providerKey, err := s.resolveInventoryProviderKey(ctx, conn, scope, source)
	if err != nil {
		return 0, err
	}

	config := &coredata.AccessInventoryProviderSyncConfig{}
	err = config.LoadByOrganizationIDAndProviderKey(
		ctx,
		conn,
		scope,
		source.OrganizationID,
		providerKey,
	)

	switch {
	case err == nil:
		duration, ok := coredata.ResolveSyncCadence(providerKey, &config.SyncCadence).ClockDuration()
		if !ok {
			return 0, fmt.Errorf(
				"cannot use sync cadence for provider %s: interval is not representable as a fixed clock duration",
				providerKey,
			)
		}

		if duration <= 0 {
			duration, ok = coredata.DefaultSyncCadence(providerKey).ClockDuration()
			if !ok || duration <= 0 {
				return time.Hour, nil
			}
		}

		return duration, nil
	case errors.Is(err, coredata.ErrResourceNotFound):
		duration, ok := coredata.ResolveSyncCadence(providerKey, nil).ClockDuration()
		if !ok || duration <= 0 {
			return time.Hour, nil
		}

		return duration, nil
	default:
		return 0, fmt.Errorf("cannot load provider sync config: %w", err)
	}
}

func (s *Service) resolveInventoryProviderKey(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (coredata.AccessInventoryProviderKey, error) {
	if source.ConnectorAccountID == nil {
		hasCSV := source.CsvData != nil && *source.CsvData != ""

		return coredata.AccessInventoryProviderKeyForSource(nil, hasCSV), nil
	}

	account := &coredata.ConnectorAccount{}
	if err := account.LoadByID(ctx, conn, scope, *source.ConnectorAccountID); err != nil {
		return "", fmt.Errorf("cannot load connector account: %w", err)
	}

	connector := &coredata.Connector{}
	if err := connector.LoadMetadataByID(ctx, conn, scope, account.ConnectorID); err != nil {
		return "", fmt.Errorf("cannot load connector: %w", err)
	}

	provider := connector.Provider

	return coredata.AccessInventoryProviderKeyForSource(&provider, false), nil
}

func (s *Service) scheduleInventorySyncAfterSuccess(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
	completedAt time.Time,
) error {
	cadence, err := s.resolveSourceSyncCadence(ctx, tx, scope, source)
	if err != nil {
		return err
	}

	next := completedAt.Add(applyInventorySyncJitter(cadence, rand.Int64N))
	source.LastSuccessfulSyncAt = &completedAt
	source.SyncConsecutiveFailures = 0
	source.NextSyncAt = &next
	source.UpdatedAt = completedAt

	if err := source.UpdateInventorySyncSchedule(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot schedule next inventory sync: %w", err)
	}

	return nil
}

func (s *Service) scheduleInventorySyncAfterFailure(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
	failedAt time.Time,
) error {
	source.SyncConsecutiveFailures++
	backoff := inventorySyncFailureBackoff(source.SyncConsecutiveFailures)
	delay := min(applyInventorySyncJitter(backoff, rand.Int64N), inventorySyncFailureMaxBackoff)

	next := failedAt.Add(delay)
	source.NextSyncAt = &next
	source.UpdatedAt = failedAt

	if err := source.UpdateInventorySyncSchedule(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot schedule inventory sync backoff: %w", err)
	}

	return nil
}

func (s *Service) advanceInventorySyncCadence(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
	now time.Time,
) error {
	cadence, err := s.resolveSourceSyncCadence(ctx, tx, scope, source)
	if err != nil {
		return err
	}

	next := now.Add(applyInventorySyncJitter(cadence, rand.Int64N))
	source.NextSyncAt = &next
	source.UpdatedAt = now

	if err := source.UpdateInventorySyncSchedule(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot advance inventory sync cadence: %w", err)
	}

	return nil
}

func inventorySyncFailureBackoff(consecutiveFailures int) time.Duration {
	if consecutiveFailures <= 0 {
		return inventorySyncFailureBaseBackoff
	}

	backoff := inventorySyncFailureBaseBackoff << min(consecutiveFailures-1, 20)
	if backoff <= 0 || backoff > inventorySyncFailureMaxBackoff {
		return inventorySyncFailureMaxBackoff
	}

	return backoff
}

// applyInventorySyncJitter spreads base by ±20%. randN must return a value in
// [0, n) for n > 0 (same contract as rand.Int64N).
func applyInventorySyncJitter(base time.Duration, randN func(int64) int64) time.Duration {
	if base <= 0 {
		return base
	}

	jitterRange := base / inventorySyncJitterFraction
	if jitterRange <= 0 {
		return base
	}

	span := int64(jitterRange)*2 + 1
	offset := time.Duration(randN(span)) - jitterRange

	return base + offset
}
