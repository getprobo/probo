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

package iam

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
	// DefaultAuditLogRetention is how long an audit log entry is kept.
	DefaultAuditLogRetention = 90 * 24 * time.Hour

	// DefaultAuditLogRetentionInterval is how often the sweep runs.
	DefaultAuditLogRetentionInterval = 24 * time.Hour

	auditLogRetentionBatchSize = 1000
)

type auditLogRetentionHandler struct {
	pg        *pg.Client
	logger    *log.Logger
	retention time.Duration
	batchSize int
	now       func() time.Time
}

// NewAuditLogRetentionWorker deletes audit log entries older than
// retention. A non-positive retention uses DefaultAuditLogRetention.
// The schedule defaults to DefaultAuditLogRetentionInterval;
// worker.WithInterval overrides it.
func NewAuditLogRetentionWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	retention time.Duration,
	opts ...worker.Option,
) *worker.PeriodicWorker {
	if retention <= 0 {
		retention = DefaultAuditLogRetention
	}

	workerOpts := append(
		[]worker.Option{worker.WithInterval(DefaultAuditLogRetentionInterval)},
		opts...,
	)

	return worker.NewPeriodic(
		"audit-log-retention-worker",
		&auditLogRetentionHandler{
			pg:        pgClient,
			logger:    logger,
			retention: retention,
			batchSize: auditLogRetentionBatchSize,
			now:       time.Now,
		},
		logger,
		workerOpts...,
	)
}

func (h *auditLogRetentionHandler) Run(ctx context.Context) error {
	before := h.now().Add(-h.retention)

	deleted, err := h.deleteBefore(ctx, before)
	if err != nil {
		return err
	}

	if deleted > 0 {
		h.logger.InfoCtx(
			ctx,
			"deleted expired audit logs",
			log.Int64("count", deleted),
		)
	}

	return nil
}

func (h *auditLogRetentionHandler) deleteBefore(
	ctx context.Context,
	before time.Time,
) (int64, error) {
	var total int64

	for {
		if err := ctx.Err(); err != nil {
			return total, fmt.Errorf("cannot delete expired audit logs: %w", err)
		}

		var deleted int64

		err := h.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				count, err := coredata.DeleteAuditLogEntriesBeforeBatch(
					ctx,
					conn,
					before,
					h.batchSize,
				)
				if err != nil {
					return err
				}

				deleted = count

				return nil
			},
		)
		if err != nil {
			return total, fmt.Errorf("cannot delete expired audit log batch: %w", err)
		}

		total += deleted

		// A short batch can mean another replica deleted part of these
		// rows. Stop only when a batch finds nothing left to delete.
		if deleted == 0 {
			return total, nil
		}
	}
}
