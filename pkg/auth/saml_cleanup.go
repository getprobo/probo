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

package auth

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
)

const (
	DefaultCleanupInterval = 1 * time.Hour
)

type (
	Cleaner struct {
		pg       *pg.Client
		interval time.Duration
		logger   *log.Logger
	}
)

func NewCleaner(
	pg *pg.Client,
	interval time.Duration,
	logger *log.Logger,
) *Cleaner {
	if interval == 0 {
		interval = DefaultCleanupInterval
	}

	return &Cleaner{
		pg:       pg,
		interval: interval,
		logger:   logger.Named("saml.cleaner"),
	}
}

func (c *Cleaner) Run(ctx context.Context) error {
	c.logger.InfoCtx(ctx, "SAML cleaner starting", log.Duration("interval", c.interval))

	if err := c.cleanup(ctx); err != nil {
		c.logger.ErrorCtx(ctx, "initial cleanup failed", log.Error(err))
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.InfoCtx(ctx, "SAML cleaner shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := c.cleanup(ctx); err != nil {
				c.logger.ErrorCtx(ctx, "periodic cleanup failed", log.Error(err))
			}
		}
	}
}

func (c *Cleaner) cleanup(ctx context.Context) error {
	var assertionsDeleted, requestsDeleted, relayStatesDeleted int64

	err := c.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			count, err := CleanupExpiredAssertions(ctx, conn)
			if err != nil {
				return err
			}
			assertionsDeleted = count

			count, err = CleanupExpiredRequests(ctx, conn)
			if err != nil {
				return err
			}
			requestsDeleted = count

			count, err = CleanupExpiredRelayStates(ctx, conn)
			if err != nil {
				return err
			}
			relayStatesDeleted = count

			return nil
		},
	)

	if err != nil {
		return err
	}

	if assertionsDeleted > 0 || requestsDeleted > 0 || relayStatesDeleted > 0 {
		c.logger.InfoCtx(ctx, "cleaned up expired SAML data",
			log.Int64("assertions", assertionsDeleted),
			log.Int64("requests", requestsDeleted),
			log.Int64("relay_states", relayStatesDeleted))
	}

	return nil
}
