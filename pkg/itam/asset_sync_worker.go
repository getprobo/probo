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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	aws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/connectorclient"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	assetSyncFailureMessage = "We couldn't list assets from this account."
	assetSyncStaleMessage   = "The asset sync stopped before it finished."

	assetSyncInterval   = 24 * time.Hour
	assetSyncStaleAfter = 15 * time.Minute
	maxStaleAssetSyncs  = 3
)

var (
	errSyncNotRunning = errors.New("asset sync is no longer running")

	defaultAssetKinds = []assetKind{
		{
			key:      CloudFrontDistributionKind,
			provider: coredata.ConnectorProviderAWS,
			list:     ListCloudFrontDistributions,
		},
		{
			key:      EC2InstanceKind,
			provider: coredata.ConnectorProviderAWS,
			list:     ListEC2Instances,
		},
		{
			key:      RDSDBInstanceKind,
			provider: coredata.ConnectorProviderAWS,
			list:     ListRDSDBInstances,
		},
	}
)

type (
	assetSyncHandler struct {
		pg         *pg.Client
		clients    *connectorclient.Client
		logger     *log.Logger
		kinds      []assetKind
		staleAfter time.Duration
		onlySource *gid.GID
	}

	pendingAsset struct {
		kind          string
		observation   AssetObservation
		configuration json.RawMessage
	}

	kindListError struct {
		kind string
		err  error
	}

	kindFailure struct {
		kind string
		err  error
	}
)

func (e *kindListError) Error() string {
	return e.err.Error()
}

func (e *kindListError) Unwrap() error {
	return e.err
}

func NewAssetSyncWorker(
	pgClient *pg.Client,
	clients *connectorclient.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AssetSync] {
	return worker.New(
		"asset-sync-worker",
		newAssetSyncHandler(pgClient, clients, logger),
		logger,
		opts...,
	)
}

func newAssetSyncHandler(
	pgClient *pg.Client,
	clients *connectorclient.Client,
	logger *log.Logger,
) *assetSyncHandler {
	kinds := make([]assetKind, len(defaultAssetKinds))
	copy(kinds, defaultAssetKinds)

	return &assetSyncHandler{
		pg:         pgClient,
		clients:    clients,
		logger:     logger,
		kinds:      kinds,
		staleAfter: assetSyncStaleAfter,
	}
}

func (h *assetSyncHandler) Claim(ctx context.Context) (coredata.AssetSync, error) {
	var sync coredata.AssetSync

	err := h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AssetSource{}
		if err := source.LoadNextDueForUpdateSkipLocked(ctx, tx, time.Now(), h.onlySource); err != nil {
			return err
		}

		attempt, err := h.nextAttempt(ctx, tx, source)
		if err != nil {
			return err
		}

		now := time.Now()
		sync = coredata.AssetSync{
			ID:             gid.New(source.ID.TenantID(), coredata.AssetSyncEntityType),
			OrganizationID: source.OrganizationID,
			AssetSourceID:  source.ID,
			Status:         coredata.AssetSyncStatusRunning,
			CompletedKinds: []string{},
			AttemptNumber:  attempt,
			StartedAt:      now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return sync.Insert(ctx, tx, coredata.NewScopeFromObjectID(source.ID))
	})
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.AssetSync{}, worker.ErrNoTask
		}

		return coredata.AssetSync{}, fmt.Errorf("cannot claim asset sync: %w", err)
	}

	return sync, nil
}

func (h *assetSyncHandler) nextAttempt(
	ctx context.Context,
	tx pg.Tx,
	source *coredata.AssetSource,
) (int, error) {
	latest := &coredata.AssetSync{}
	err := latest.LoadLatestBySourceID(ctx, tx, coredata.NewScopeFromObjectID(source.ID), source.ID)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return 1, nil
	}

	if err != nil {
		return 0, fmt.Errorf("cannot load latest asset sync: %w", err)
	}

	if latest.Status == coredata.AssetSyncStatusFailed &&
		latest.Error != nil &&
		*latest.Error == assetSyncStaleMessage {
		return latest.AttemptNumber + 1, nil
	}

	return 1, nil
}

func (h *assetSyncHandler) Process(ctx context.Context, sync coredata.AssetSync) error {
	source := &coredata.AssetSource{}

	err := h.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByID(ctx, conn, coredata.NewScopeFromObjectID(sync.AssetSourceID), sync.AssetSourceID)
	})
	if err != nil {
		return fmt.Errorf("cannot load asset source %s: %w", sync.AssetSourceID, err)
	}

	if err := h.syncSource(ctx, source, &sync); err != nil {
		if errors.Is(err, errSyncNotRunning) {
			return nil
		}

		if failErr := h.failSync(ctx, source, &sync, err); failErr != nil {
			return fmt.Errorf("cannot record asset sync failure: %w", failErr)
		}
	}

	return nil
}

func (h *assetSyncHandler) syncSource(
	ctx context.Context,
	source *coredata.AssetSource,
	sync *coredata.AssetSync,
) error {
	kind := h.kindFor(source.Provider)

	if source.ConnectorAccountID == nil {
		return &kindListError{
			kind: kind,
			err:  fmt.Errorf("cannot list assets: connector account is gone"),
		}
	}

	var handle *connectorclient.Handle

	err := h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var openErr error
		handle, openErr = h.clients.Open(
			ctx,
			tx,
			coredata.NewScopeFromObjectID(source.ID),
			*source.ConnectorAccountID,
		)

		return openErr
	})
	if err != nil {
		return &kindListError{kind: kind, err: fmt.Errorf("cannot open connector account: %w", err)}
	}

	session, ok := handle.Session.(*aws.Session)
	if !ok || session == nil {
		return &kindListError{
			kind: kind,
			err:  fmt.Errorf("cannot list assets: connector account is not an AWS session"),
		}
	}

	pending := make([]pendingAsset, 0)
	completed := make([]string, 0, len(h.kinds))
	failures := make([]kindFailure, 0)

	for _, item := range h.kinds {
		if item.provider != source.Provider {
			continue
		}

		observed, err := item.list(ctx, session)
		if err != nil {
			failures = append(failures, kindFailure{kind: item.key, err: err})
			continue
		}

		kindPending, err := pendingForKind(item.key, observed)
		if err != nil {
			failures = append(failures, kindFailure{kind: item.key, err: err})
			continue
		}

		pending = append(pending, kindPending...)
		completed = append(completed, item.key)
	}

	return h.finish(ctx, source, sync, pending, completed, failures)
}

func pendingForKind(kind string, observed []AssetObservation) ([]pendingAsset, error) {
	pending := make([]pendingAsset, 0, len(observed))

	for _, observation := range observed {
		raw, err := json.Marshal(observation.Configuration)
		if err != nil {
			return nil, fmt.Errorf("cannot encode asset configuration: %w", err)
		}

		if len(raw) == 0 || string(raw) == "null" {
			raw = []byte(`{}`)
		}

		pending = append(pending, pendingAsset{
			kind:          kind,
			observation:   observation,
			configuration: raw,
		})
	}

	return pending, nil
}

func (h *assetSyncHandler) finish(
	ctx context.Context,
	source *coredata.AssetSource,
	sync *coredata.AssetSync,
	pending []pendingAsset,
	completed []string,
	failures []kindFailure,
) error {
	for _, failure := range failures {
		h.logger.ErrorCtx(
			ctx,
			"cannot list assets",
			log.String("assets_source_id", source.ID.String()),
			log.String("kind", failure.kind),
			log.Error(failure.err),
		)
	}

	status := coredata.AssetSyncStatusSucceeded
	var syncErr *string
	if len(failures) > 0 {
		status = coredata.AssetSyncStatusFailed
		message := assetSyncFailureMessage
		syncErr = &message
	}

	return h.store(ctx, source, sync, pending, completed, status, syncErr)
}

func (h *assetSyncHandler) store(
	ctx context.Context,
	source *coredata.AssetSource,
	sync *coredata.AssetSync,
	pending []pendingAsset,
	completed []string,
	status coredata.AssetSyncStatus,
	syncErr *string,
) error {
	return h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		now := time.Now()
		scope := coredata.NewScopeFromObjectID(source.ID)

		for _, item := range pending {
			asset := &coredata.InventoryAsset{
				ID:             gid.New(source.ID.TenantID(), coredata.InventoryAssetEntityType),
				OrganizationID: source.OrganizationID,
				AssetSourceID:  source.ID,
				Kind:           item.kind,
				ExternalID:     item.observation.ExternalID,
				Name:           item.observation.Name,
				Region:         item.observation.Region,
				Configuration:  item.configuration,
				SchemaVersion:  item.observation.SchemaVersion,
				CreatedAt:      now,
				UpdatedAt:      now,
				LastSeenAt:     now,
			}

			if err := asset.Upsert(ctx, tx, scope); err != nil {
				return err
			}
		}

		updated, err := sync.FinishIfRunning(
			ctx,
			tx,
			scope,
			status,
			syncErr,
			completed,
			now,
		)
		if err != nil {
			return err
		}

		if !updated {
			return errSyncNotRunning
		}

		return source.SetNextSyncAt(ctx, tx, scope, now.Add(assetSyncInterval))
	})
}

func (h *assetSyncHandler) failSync(
	ctx context.Context,
	source *coredata.AssetSource,
	sync *coredata.AssetSync,
	cause error,
) error {
	kind, logged := kindAndCause(cause)
	h.logger.ErrorCtx(
		ctx,
		"cannot list assets",
		log.String("assets_source_id", source.ID.String()),
		log.String("kind", kind),
		log.Error(logged),
	)

	message := assetSyncFailureMessage
	now := time.Now()

	return h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		scope := coredata.NewScopeFromObjectID(source.ID)

		updated, err := sync.FinishIfRunning(
			ctx,
			tx,
			scope,
			coredata.AssetSyncStatusFailed,
			&message,
			[]string{},
			now,
		)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		return source.SetNextSyncAt(ctx, tx, scope, now.Add(assetSyncInterval))
	})
}

func (h *assetSyncHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		syncs, err := coredata.FailStaleAssetSyncs(
			ctx,
			tx,
			time.Now().Add(-h.staleAfter),
			assetSyncStaleMessage,
		)
		if err != nil {
			return err
		}

		if len(syncs) == 0 {
			return nil
		}

		now := time.Now()

		for _, sync := range syncs {
			next := now.Add(-time.Second)
			if sync.AttemptNumber >= maxStaleAssetSyncs {
				next = now.Add(assetSyncInterval)
			}

			source := &coredata.AssetSource{ID: sync.AssetSourceID}
			if err := source.SetNextSyncAt(ctx, tx, coredata.NewScopeFromObjectID(sync.ID), next); err != nil {
				return fmt.Errorf("cannot schedule asset source %s: %w", sync.AssetSourceID, err)
			}
		}

		h.logger.InfoCtx(ctx, "swept stale asset syncs", log.Int("count", len(syncs)))

		return nil
	})
}

func (h *assetSyncHandler) kindFor(provider coredata.ConnectorProvider) string {
	for _, kind := range h.kinds {
		if kind.provider == provider {
			return kind.key
		}
	}

	return ""
}

func kindAndCause(err error) (string, error) {
	if listErr, ok := errors.AsType[*kindListError](err); ok && listErr != nil {
		return listErr.kind, listErr.err
	}

	return "", err
}
