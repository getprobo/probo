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

package itam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/itam/mdm"
)

const (
	// DefaultDeviceSyncInterval is how often a healthy connector's devices
	// are synced.
	DefaultDeviceSyncInterval = time.Hour

	// deviceSyncBaseBackoff is the wait after a first failure; each further
	// consecutive failure doubles it, up to the sync interval.
	deviceSyncBaseBackoff = 5 * time.Minute

	// deviceSyncFetchTimeout bounds one provider listing. A large fleet is a
	// few hundred paged requests.
	deviceSyncFetchTimeout = 10 * time.Minute

	// deviceSyncTimeout bounds a whole sync: the listing, then applying it.
	deviceSyncTimeout = deviceSyncFetchTimeout + 3*time.Minute

	// deviceSyncLease is how long a claimed sync stays out of every other
	// worker's reach. It outlasts deviceSyncTimeout, so a sync still running
	// is never claimed a second time; a worker that crashed mid-sync leaves
	// the row due again once it runs out.
	deviceSyncLease = deviceSyncTimeout + 2*time.Minute
)

type (
	DeviceSyncWorker = worker.Worker[coredata.DeviceSync]

	deviceSyncHandler struct {
		pg                *pg.Client
		itam              *Service
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connectorprotocol.Registry
		providerRegistry  *provider.Registry
		providers         []coredata.ConnectorProvider
		interval          time.Duration
		logger            *log.Logger
	}
)

// NewDeviceSyncWorker syncs the devices of every connector whose provider
// registers a device driver, one connector at a time.
func NewDeviceSyncWorker(
	pgClient *pg.Client,
	itamSvc *Service,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connectorprotocol.Registry,
	providerRegistry *provider.Registry,
	logger *log.Logger,
	opts ...worker.Option,
) *DeviceSyncWorker {
	var providers []coredata.ConnectorProvider

	for _, reg := range providerRegistry.All() {
		if reg != nil && reg.SyncsDevices() {
			providers = append(providers, reg.Provider)
		}
	}

	h := &deviceSyncHandler{
		providers:         providers,
		pg:                pgClient,
		itam:              itamSvc,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		providerRegistry:  providerRegistry,
		interval:          DefaultDeviceSyncInterval,
		logger:            logger.Named("itam.device_sync"),
	}

	return worker.New(
		"itam.device_sync",
		h,
		logger,
		append(
			[]worker.Option{
				worker.WithInterval(30 * time.Second),
				worker.WithMaxConcurrency(1),
			},
			opts...,
		)...,
	)
}

// deviceSyncBackoff is how long a sync waits after its attempt-th consecutive
// failure, doubling from deviceSyncBaseBackoff up to the regular interval.
func deviceSyncBackoff(attempt int, interval time.Duration) time.Duration {
	backoff := deviceSyncBaseBackoff << max(attempt-1, 0)
	if backoff > interval || backoff <= 0 {
		return interval
	}

	return backoff
}

func (h *deviceSyncHandler) Claim(ctx context.Context) (coredata.DeviceSync, error) {
	var sync coredata.DeviceSync

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			if _, err := coredata.ScheduleMissingDeviceSyncs(ctx, tx, h.providers, now); err != nil {
				return err
			}

			if err := sync.LoadNextDueForUpdateSkipLocked(ctx, tx, now); err != nil {
				return err
			}

			// Charged before the provider is called, so a crash or a hung
			// request cannot leave the row due on the very next poll.
			return sync.RecordAttempt(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(sync.ConnectorID),
				now,
				now.Add(deviceSyncLease),
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.DeviceSync{}, worker.ErrNoTask
		}

		return coredata.DeviceSync{}, fmt.Errorf("cannot claim device sync: %w", err)
	}

	return sync, nil
}

func (h *deviceSyncHandler) Process(ctx context.Context, sync coredata.DeviceSync) error {
	logger := h.logger.With(log.String("connector_id", sync.ConnectorID.String()))
	scope := coredata.NewScopeFromObjectID(sync.ConnectorID)

	syncCtx, cancel := context.WithTimeout(ctx, deviceSyncTimeout)
	defer cancel()

	result, err := h.sync(syncCtx, &sync)

	switch {
	case errors.Is(err, coredata.ErrDeviceSyncSuperseded):
		// The connector was saved while this sync ran, so what it fetched
		// used the previous credentials and was not applied. The save made
		// the sync due again; giving up the lease lets it run now.
		releaseErr := h.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return sync.Release(ctx, conn, scope, time.Now())
			},
		)
		if releaseErr != nil && !errors.Is(releaseErr, coredata.ErrResourceNotFound) {
			logger.ErrorCtx(ctx, "cannot release device sync", log.Error(releaseErr))
		}

		logger.InfoCtx(ctx, "device sync superseded by a connector update, not applied")

		return nil
	case errors.Is(err, coredata.ErrResourceNotFound):
		// The connector was deleted, or the lease ran out and another worker
		// claimed the sync. Nothing was applied, and there is either no sync
		// left or a newer claim that owns its outcome.
		logger.InfoCtx(ctx, "device sync claim lost, not applied")

		return nil
	}

	if err != nil {
		failErr := h.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				now := time.Now()

				return sync.RecordFailure(
					ctx,
					conn,
					scope,
					now,
					now.Add(deviceSyncBackoff(sync.Attempts, h.interval)),
				)
			},
		)

		switch {
		case errors.Is(failErr, coredata.ErrResourceNotFound):
			logger.InfoCtx(ctx, "device sync claim lost, keeping the newer claim's outcome")
		case failErr != nil:
			logger.ErrorCtx(ctx, "cannot record device sync failure", log.Error(failErr))
		}

		logger.WarnCtx(
			ctx,
			"device sync failed",
			log.Int("attempts", sync.Attempts),
			log.Error(err),
		)

		return err
	}

	now := time.Now()

	err = h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return sync.RecordSuccess(ctx, conn, scope, now, now.Add(h.interval))
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		// The connector was deleted after the devices were applied.
		logger.InfoCtx(ctx, "device sync claim lost after applying")
	} else if err != nil {
		return fmt.Errorf("cannot record device sync success: %w", err)
	}

	logger.InfoCtx(
		ctx,
		"device sync completed",
		log.Int("created", result.Created),
		log.Int("updated", result.Updated),
		log.Int("linked", result.Linked),
		log.Int64("revoked", result.Revoked),
		log.Int64("unlinked", result.Unlinked),
	)

	return nil
}

func (h *deviceSyncHandler) sync(ctx context.Context, sync *coredata.DeviceSync) (*DeviceSyncResult, error) {
	var (
		connector coredata.Connector
		driver    mdm.Driver
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(sync.ConnectorID)

			if err := connector.LoadByID(ctx, tx, scope, sync.ConnectorID, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			d, err := h.newDriver(ctx, tx, scope, &connector)
			if err != nil {
				return err
			}

			driver = d

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	fetchCtx, cancel := context.WithTimeout(ctx, deviceSyncFetchTimeout)
	defer cancel()

	records, err := driver.ListDevices(fetchCtx)
	if err != nil {
		return nil, fmt.Errorf("cannot list devices: %w", err)
	}

	result, err := h.itam.ApplyDeviceSync(ctx, sync, &connector, records)
	if err != nil {
		return nil, fmt.Errorf("cannot apply device sync: %w", err)
	}

	return result, nil
}

// newDriver builds the provider's device driver. A provider whose credential
// is an OAuth token refreshed on the way has the new token written back, as
// the other connector workers do, so the next caller does not reuse a
// rotated refresh token.
func (h *deviceSyncHandler) newDriver(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	connector *coredata.Connector,
) (mdm.Driver, error) {
	reg, ok := h.providerRegistry.Get(connector.Provider)
	if !ok || !reg.SyncsDevices() {
		return nil, fmt.Errorf("cannot sync devices: provider %q registers no device driver", connector.Provider)
	}

	oauth2Conn := connectorprotocol.OAuth2ConnectionOf(connector.Connection)

	var tokenBefore string
	if oauth2Conn != nil {
		tokenBefore = oauth2Conn.AccessToken
	}

	httpClient, err := provider.HTTPClient(ctx, h.connectorRegistry, h.providerRegistry, connector)
	if err != nil {
		return nil, fmt.Errorf("cannot create http client: %w", err)
	}

	if oauth2Conn != nil && oauth2Conn.AccessToken != tokenBefore {
		connector.UpdatedAt = time.Now()
		if err := connector.UpdateConnection(ctx, tx, scope, h.encryptionKey); err != nil {
			return nil, fmt.Errorf("cannot persist refreshed token: %w", err)
		}
	}

	driver, err := reg.NewDeviceDriver(ctx, httpClient, connector, h.logger, reg.Endpoints)
	if err != nil {
		return nil, fmt.Errorf("cannot create device driver: %w", err)
	}

	return driver, nil
}
