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

package scim

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

// SyncStats holds the statistics from a sync operation.
type SyncStats struct {
	Created     int
	Updated     int
	Deleted     int
	Deactivated int
	Skipped     int
}

func (r *BridgeRunner) acquireNextBridge(ctx context.Context) (*coredata.SCIMBridge, coredata.Scoper, error) {
	var bridge *coredata.SCIMBridge
	var scope coredata.Scoper

	err := r.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			bridge = &coredata.SCIMBridge{}
			if err := bridge.LoadNextForSyncSkipLocked(ctx, tx, r.cfg.StaleSyncThreshold); err != nil {
				return err
			}

			scope = coredata.NewScope(bridge.ID.TenantID())

			now := time.Now()
			bridge.State = coredata.SCIMBridgeStateSyncing
			bridge.UpdatedAt = now

			return bridge.Update(ctx, tx, scope)
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return bridge, scope, nil
}

func (r *BridgeRunner) transitionToSuccess(
	ctx context.Context,
	bridge *coredata.SCIMBridge,
	scope coredata.Scoper,
	stats SyncStats,
	duration time.Duration,
	connector *coredata.Connector,
	logger *log.Logger,
) error {
	return r.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			now := time.Now()
			nextSync := now.Add(r.cfg.Interval)

			bridge.State = coredata.SCIMBridgeStateActive
			bridge.LastSyncedAt = &now
			bridge.NextSyncAt = &nextSync
			bridge.SyncError = nil
			bridge.ConsecutiveFailures = 0
			bridge.TotalSyncCount++
			bridge.UpdatedAt = now

			if err := bridge.Update(ctx, conn, scope); err != nil {
				logger.ErrorCtx(ctx, "cannot update bridge after successful sync",
					log.Error(err),
				)
				return err
			}

			if connector != nil {
				connector.UpdatedAt = now
				if err := connector.Update(ctx, conn, scope, r.encryptionKey); err != nil {
					logger.WarnCtx(ctx, "cannot persist refreshed OAuth2 token",
						log.String("connector_id", connector.ID.String()),
						log.Error(err),
					)
				}
			}

			logger.InfoCtx(ctx, "sync completed successfully",
				log.Duration("sync_duration", duration),
				log.Int("users_created", stats.Created),
				log.Int("users_updated", stats.Updated),
				log.Int("users_deleted", stats.Deleted),
				log.Int("users_deactivated", stats.Deactivated),
				log.Int("users_skipped", stats.Skipped),
			)

			return nil
		},
	)
}

func (r *BridgeRunner) transitionToFailed(
	ctx context.Context,
	bridge *coredata.SCIMBridge,
	scope coredata.Scoper,
	syncErr error,
	duration time.Duration,
	logger *log.Logger,
) error {
	return r.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			now := time.Now()

			bridge.ConsecutiveFailures++
			bridge.TotalFailureCount++
			bridge.TotalSyncCount++
			bridge.LastSyncedAt = &now
			bridge.UpdatedAt = now

			errStr := syncErr.Error()
			bridge.SyncError = &errStr

			if r.shouldDisable(bridge.ConsecutiveFailures) {
				bridge.State = coredata.SCIMBridgeStateDisabled
				bridge.NextSyncAt = nil

				logger.ErrorCtx(ctx, "bridge disabled due to max consecutive failures",
					log.Duration("sync_duration", duration),
					log.Int("consecutive_failures", bridge.ConsecutiveFailures),
					log.Int("max_consecutive_failures", r.cfg.MaxConsecutiveFailures),
					log.Error(syncErr),
				)
			} else {
				bridge.State = coredata.SCIMBridgeStateFailed
				backoff := r.calculateBackoff(bridge.ConsecutiveFailures)
				nextSync := now.Add(backoff)
				bridge.NextSyncAt = &nextSync

				logger.ErrorCtx(ctx, "sync failed, will retry with backoff",
					log.Duration("sync_duration", duration),
					log.Int("consecutive_failures", bridge.ConsecutiveFailures),
					log.Duration("next_retry_in", backoff),
					log.Error(syncErr),
				)
			}

			if err := bridge.Update(ctx, conn, scope); err != nil {
				logger.ErrorCtx(ctx, "cannot update bridge after failed sync",
					log.String("new_state", string(bridge.State)),
					log.Error(err),
				)
				return err
			}

			return nil
		},
	)
}
