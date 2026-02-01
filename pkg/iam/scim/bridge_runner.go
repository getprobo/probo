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
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type (
	// BridgeRunnerConfig holds the configuration for the SCIM bridge runner.
	BridgeRunnerConfig struct {
		// Interval is the time between sync attempts for each bridge.
		Interval time.Duration
		// PollInterval is the time between polling for bridges to sync.
		PollInterval time.Duration
		// SyncTimeout is the maximum time allowed for a single sync operation.
		SyncTimeout time.Duration
		// BaseURL is the base URL of the API server (used to construct SCIM endpoint).
		BaseURL *baseurl.BaseURL
		// MaxBackoff is the maximum backoff duration between retries for failed bridges.
		MaxBackoff time.Duration
		// MaxConsecutiveFailures is the maximum number of consecutive failures
		// before a bridge is automatically disabled.
		MaxConsecutiveFailures int
		// StaleSyncThreshold is the time after which a SYNCING bridge is considered
		// stale and can be recovered by another runner (handles crashed runners).
		StaleSyncThreshold time.Duration
	}

	// BridgeRunner is the SCIM bridge background runner.
	BridgeRunner struct {
		pg                *pg.Client
		logger            *log.Logger
		tp                trace.TracerProvider
		tracer            trace.Tracer
		registerer        prometheus.Registerer
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connector.ConnectorRegistry
		cfg               BridgeRunnerConfig
	}
)

// NewBridgeRunner creates a new SCIM bridge runner.
func NewBridgeRunner(
	pgClient *pg.Client,
	logger *log.Logger,
	tp trace.TracerProvider,
	registerer prometheus.Registerer,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	cfg BridgeRunnerConfig,
) *BridgeRunner {
	if cfg.Interval == 0 {
		cfg.Interval = 15 * time.Minute
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 30 * time.Second
	}
	if cfg.SyncTimeout == 0 {
		cfg.SyncTimeout = 5 * time.Minute
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = DefaultMaxBackoff
	}
	if cfg.MaxConsecutiveFailures == 0 {
		cfg.MaxConsecutiveFailures = DefaultMaxConsecutiveFailures
	}
	if cfg.StaleSyncThreshold == 0 {
		cfg.StaleSyncThreshold = DefaultStaleSyncThreshold
	}

	return &BridgeRunner{
		pg:                pgClient,
		logger:            logger,
		tp:                tp,
		tracer:            tp.Tracer("scim-bridge-runner"),
		registerer:        registerer,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		cfg:               cfg,
	}
}

// Run starts the runner loop that processes SCIM bridges.
func (r *BridgeRunner) Run(ctx context.Context) error {
	r.logger.InfoCtx(ctx, "starting SCIM bridge runner",
		log.Duration("poll_interval", r.cfg.PollInterval),
		log.Duration("sync_interval", r.cfg.Interval),
		log.Duration("sync_timeout", r.cfg.SyncTimeout),
		log.Duration("max_backoff", r.cfg.MaxBackoff),
		log.Int("max_consecutive_failures", r.cfg.MaxConsecutiveFailures),
		log.Duration("stale_sync_threshold", r.cfg.StaleSyncThreshold),
	)

	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	for {
		if err := r.processBridge(ctx); err != nil {
			if !errors.Is(err, coredata.ErrNoSCIMBridgeAvailable) {
				r.logger.ErrorCtx(ctx, "cannot process SCIM bridge", log.Error(err))
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *BridgeRunner) processBridge(ctx context.Context) error {
	bridge, scope, err := r.acquireNextBridge(ctx)
	if err != nil {
		return err
	}

	ctx, span := r.tracer.Start(ctx, "scim-bridge-runner.processBridge")
	defer span.End()

	logger := r.logger.Named("bridge-sync").With(
		log.String("bridge_id", bridge.ID.String()),
		log.String("scim_configuration_id", bridge.ScimConfigurationID.String()),
		log.String("bridge_type", string(bridge.Type)),
		log.Int("consecutive_failures", bridge.ConsecutiveFailures),
	)

	logger.InfoCtx(ctx, "starting sync")

	syncCtx, cancel := context.WithTimeout(ctx, r.cfg.SyncTimeout)
	defer cancel()

	stats, duration, connector, err := r.executeSync(syncCtx, bridge, scope, logger)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "sync failed")
		return r.transitionToFailed(ctx, bridge, scope, err, duration, logger)
	}

	return r.transitionToSuccess(ctx, bridge, scope, stats, duration, connector, logger)
}
