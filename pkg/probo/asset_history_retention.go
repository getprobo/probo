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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// DefaultAssetHistoryRetention is how long an asset snapshot and a
	// change event are kept. A snapshot name does not extend that.
	DefaultAssetHistoryRetention = 90 * 24 * time.Hour

	// DefaultAssetHistoryRetentionInterval is how often the sweep runs.
	DefaultAssetHistoryRetentionInterval = 24 * time.Hour

	assetHistoryRetentionBatchSize = 1000
)

type assetHistoryRetentionHandler struct {
	pg        *pg.Client
	logger    *log.Logger
	retention time.Duration
	batchSize int
	now       func() time.Time
}

// NewAssetHistoryRetentionWorker deletes asset snapshots and change events
// older than retention. A non-positive retention uses
// DefaultAssetHistoryRetention. The schedule defaults to
// DefaultAssetHistoryRetentionInterval; worker.WithInterval overrides it.
func NewAssetHistoryRetentionWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	retention time.Duration,
	opts ...worker.Option,
) *worker.PeriodicWorker {
	if retention <= 0 {
		retention = DefaultAssetHistoryRetention
	}

	workerOpts := append(
		[]worker.Option{worker.WithInterval(DefaultAssetHistoryRetentionInterval)},
		opts...,
	)

	return worker.NewPeriodic(
		"asset-history-retention-worker",
		&assetHistoryRetentionHandler{
			pg:        pgClient,
			logger:    logger,
			retention: retention,
			batchSize: assetHistoryRetentionBatchSize,
			now:       time.Now,
		},
		logger,
		workerOpts...,
	)
}

func (h *assetHistoryRetentionHandler) Run(ctx context.Context) error {
	before := h.now().Add(-h.retention)

	snapshots, err := h.deleteSnapshotsBefore(ctx, before)
	if err != nil {
		return err
	}

	events, err := h.deleteEventsBefore(ctx, before)
	if err != nil {
		return err
	}

	if snapshots > 0 || events > 0 {
		h.logger.InfoCtx(
			ctx,
			"deleted expired asset history",
			log.Int64("snapshots", snapshots),
			log.Int64("change_events", events),
		)
	}

	return nil
}

func (h *assetHistoryRetentionHandler) deleteSnapshotsBefore(
	ctx context.Context,
	before time.Time,
) (int64, error) {
	return h.deleteBefore(
		ctx,
		"asset snapshots",
		func(ctx context.Context, conn pg.Querier) (int64, error) {
			return coredata.DeleteAssetSnapshotsBeforeBatch(ctx, conn, before, h.batchSize)
		},
	)
}

func (h *assetHistoryRetentionHandler) deleteEventsBefore(
	ctx context.Context,
	before time.Time,
) (int64, error) {
	return h.deleteBefore(
		ctx,
		"asset change events",
		func(ctx context.Context, conn pg.Querier) (int64, error) {
			return coredata.DeleteAssetChangeEventsBeforeBatch(ctx, conn, before, h.batchSize)
		},
	)
}

func (h *assetHistoryRetentionHandler) deleteBefore(
	ctx context.Context,
	kind string,
	deleteBatch func(context.Context, pg.Querier) (int64, error),
) (int64, error) {
	var total int64

	for {
		if err := ctx.Err(); err != nil {
			return total, fmt.Errorf("cannot delete expired %s: %w", kind, err)
		}

		var deleted int64

		err := h.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				count, err := deleteBatch(ctx, conn)
				if err != nil {
					return fmt.Errorf("cannot delete expired %s batch: %w", kind, err)
				}

				deleted = count

				return nil
			},
		)
		if err != nil {
			return total, err
		}

		total += deleted

		// A short batch can mean another replica deleted part of these
		// rows. Stop only when a batch finds nothing left to delete.
		if deleted == 0 {
			return total, nil
		}
	}
}
