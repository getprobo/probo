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

package cookiebanner

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
	analyticsRetentionBatch  = 5000
	analyticsRetentionPeriod = time.Hour
)

type (
	analyticsRetentionHandler struct {
		pg        *pg.Client
		logger    *log.Logger
		batchSize int
		now       func() time.Time
	}

	analyticsCleanup struct {
		table       string
		deleteBatch func(ctx context.Context, conn pg.Querier, before time.Time, limit int) (int64, error)
	}
)

var analyticsCleanups = []analyticsCleanup{
	{
		table:       "cookie_banner_daily_page_loads",
		deleteBatch: coredata.DeleteDailyPageLoadsBeforeBatch,
	},
	{
		table:       "tracker_pattern_daily_hits",
		deleteBatch: coredata.DeleteDailyPatternHitsBeforeBatch,
	},
}

// NewAnalyticsRetentionWorker deletes the daily analytics buckets that fall
// outside the longest window the API serves, MaxAnalyticsDays.
func NewAnalyticsRetentionWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.PeriodicWorker {
	workerOpts := append(
		[]worker.Option{worker.WithInterval(analyticsRetentionPeriod)},
		opts...,
	)

	return worker.NewPeriodic(
		"cookie-banner-analytics-retention-worker",
		&analyticsRetentionHandler{
			pg:        pgClient,
			logger:    logger,
			batchSize: analyticsRetentionBatch,
			now:       time.Now,
		},
		logger,
		workerOpts...,
	)
}

func (h *analyticsRetentionHandler) Run(ctx context.Context) error {
	before := analyticsSince(h.now(), MaxAnalyticsDays)

	for _, cleanup := range analyticsCleanups {
		deleted, err := h.deleteBefore(ctx, cleanup, before)
		if err != nil {
			return fmt.Errorf("cannot clean %s: %w", cleanup.table, err)
		}

		if deleted > 0 {
			h.logger.InfoCtx(
				ctx,
				"deleted expired cookie banner analytics",
				log.String("table", cleanup.table),
				log.Int64("count", deleted),
			)
		}
	}

	return nil
}

func (h *analyticsRetentionHandler) deleteBefore(
	ctx context.Context,
	cleanup analyticsCleanup,
	before time.Time,
) (int64, error) {
	var total int64

	for {
		var deleted int64

		err := h.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				count, err := cleanup.deleteBatch(ctx, conn, before, h.batchSize)
				if err != nil {
					return err
				}

				deleted = count

				return nil
			},
		)
		if err != nil {
			return total, err
		}

		total += deleted

		if deleted < int64(h.batchSize) {
			return total, nil
		}
	}
}
