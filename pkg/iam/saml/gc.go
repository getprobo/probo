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

package saml

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	DefaultGarbageCollectionInterval = 1 * time.Hour
)

type (
	GarbageCollector struct {
		pg       *pg.Client
		interval time.Duration
		logger   *log.Logger
	}

	GarbageCollectorOption func(*GarbageCollector)
)

func WithGarbageCollectionInterval(interval time.Duration) GarbageCollectorOption {
	return func(gc *GarbageCollector) {
		gc.interval = interval
	}
}

func NewGarbageCollector(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...GarbageCollectorOption,
) *GarbageCollector {
	gc := &GarbageCollector{
		pg:       pgClient,
		interval: DefaultGarbageCollectionInterval,
		logger:   logger.Named("saml.garbage_collector"),
	}

	for _, opt := range opts {
		opt(gc)
	}

	gc.logger = gc.logger.With(log.Duration("interval", gc.interval))

	return gc
}

func (gc *GarbageCollector) Run(ctx context.Context) error {
	gc.logger.InfoCtx(ctx, "saml garbage collector starting")

	if err := gc.cleanup(ctx); err != nil {
		gc.logger.ErrorCtx(ctx, "cannot run initial cleanup", log.Error(err))
	}

	if gc.interval <= 0 {
		return fmt.Errorf("cannot run SAML garbage collector: interval must be greater than zero")
	}

	ticker := time.NewTicker(gc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			gc.logger.InfoCtx(ctx, "saml garbage collector shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := gc.cleanup(ctx); err != nil {
				gc.logger.ErrorCtx(ctx, "cannot run periodic cleanup", log.Error(err))
			}
		}
	}
}

func (gc *GarbageCollector) cleanup(ctx context.Context) error {
	now := time.Now()

	return gc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			assertionsDeleted, err := coredata.DeleteExpiredSAMLAssertions(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired saml assertions: %w", err)
			}

			requestsDeleted, err := coredata.DeleteExpiredSAMLRequests(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired saml requests: %w", err)
			}

			gc.logger.InfoCtx(
				ctx,
				"saml garbage collector cleaned up expired assertions and requests",
				log.Int64("assertions", assertionsDeleted),
				log.Int64("requests", requestsDeleted),
			)

			return nil
		},
	)
}
