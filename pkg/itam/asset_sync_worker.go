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
	"go.probo.inc/probo/pkg/connector"
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
		pg          *pg.Client
		connectors  *connector.Service
		logger      *log.Logger
		kinds       []assetKind
		staleAfter  time.Duration
		onlyAccount *gid.GID
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
	connectors *connector.Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AssetSync] {
	return worker.New(
		"asset-sync-worker",
		newAssetSyncHandler(pgClient, connectors, logger),
		logger,
		opts...,
	)
}

func newAssetSyncHandler(
	pgClient *pg.Client,
	connectors *connector.Service,
	logger *log.Logger,
) *assetSyncHandler {
	kinds := make([]assetKind, len(defaultAssetKinds))
	copy(kinds, defaultAssetKinds)

	return &assetSyncHandler{
		pg:         pgClient,
		connectors: connectors,
		logger:     logger,
		kinds:      kinds,
		staleAfter: assetSyncStaleAfter,
	}
}

func (h *assetSyncHandler) Claim(ctx context.Context) (coredata.AssetSync, error) {
	var sync coredata.AssetSync

	err := h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		account := &coredata.ConnectorAccount{}
		if err := account.LoadNextDueForAssetSyncForUpdateSkipLocked(
			ctx,
			tx,
			coredata.ConnectorProviderAWS,
			time.Now().Add(-assetSyncInterval),
			assetSyncStaleMessage,
			maxStaleAssetSyncs,
			h.onlyAccount,
		); err != nil {
			return err
		}

		attempt, err := h.nextAttempt(ctx, tx, account)
		if err != nil {
			return err
		}

		now := time.Now()
		sync = coredata.AssetSync{
			ID:                 gid.New(account.ID.TenantID(), coredata.AssetSyncEntityType),
			OrganizationID:     account.OrganizationID,
			ConnectorAccountID: account.ID,
			Status:             coredata.AssetSyncStatusRunning,
			CompletedKinds:     []string{},
			AttemptNumber:      attempt,
			StartedAt:          now,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		return sync.Insert(ctx, tx, coredata.NewScopeFromObjectID(account.ID))
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
	account *coredata.ConnectorAccount,
) (int, error) {
	latest := &coredata.AssetSync{}

	err := latest.LoadLatestByConnectorAccountID(
		ctx,
		tx,
		coredata.NewScopeFromObjectID(account.ID),
		account.ID,
	)
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
	account := &coredata.ConnectorAccount{}
	scope := coredata.NewScopeFromObjectID(sync.ID)

	err := h.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return account.LoadByID(ctx, conn, scope, sync.ConnectorAccountID)
	})
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load connector account %s: %w", sync.ConnectorAccountID, err)
	}

	if err := h.syncAccount(ctx, account, &sync); err != nil {
		if errors.Is(err, errSyncNotRunning) {
			return nil
		}

		if failErr := h.failSync(ctx, account, &sync, err); failErr != nil {
			return fmt.Errorf("cannot record asset sync failure: %w", failErr)
		}
	}

	return nil
}

func (h *assetSyncHandler) syncAccount(
	ctx context.Context,
	account *coredata.ConnectorAccount,
	sync *coredata.AssetSync,
) error {
	scope := coredata.NewScopeFromObjectID(account.ID)

	handle, err := h.connectors.Open(ctx, scope, account.ID)
	if err != nil {
		return &kindListError{
			err: fmt.Errorf("cannot open connector account: %w", err),
		}
	}

	session, ok := handle.Session.(*aws.Session)
	if !ok || session == nil {
		return &kindListError{
			kind: h.kindFor(handle.Provider),
			err:  fmt.Errorf("cannot list assets: connector account is not an AWS session"),
		}
	}

	pending := make([]pendingAsset, 0)
	completed := make([]string, 0, len(h.kinds))
	failures := make([]kindFailure, 0)

	for _, item := range h.kinds {
		if item.provider != handle.Provider {
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

	return h.finish(ctx, account, sync, pending, completed, failures)
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
	account *coredata.ConnectorAccount,
	sync *coredata.AssetSync,
	pending []pendingAsset,
	completed []string,
	failures []kindFailure,
) error {
	for _, failure := range failures {
		h.logger.ErrorCtx(
			ctx,
			"cannot list assets",
			log.String("connector_account_id", account.ID.String()),
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

	return h.store(ctx, account, sync, pending, completed, status, syncErr)
}

func (h *assetSyncHandler) store(
	ctx context.Context,
	account *coredata.ConnectorAccount,
	sync *coredata.AssetSync,
	pending []pendingAsset,
	completed []string,
	status coredata.AssetSyncStatus,
	syncErr *string,
) error {
	return h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		now := time.Now()
		scope := coredata.NewScopeFromObjectID(account.ID)
		assets := make(coredata.Assets, 0, len(pending))

		for _, item := range pending {
			assets = append(assets, &coredata.Asset{
				ID:             gid.New(account.ID.TenantID(), coredata.AssetEntityType),
				OrganizationID: account.OrganizationID,
				Name:           item.observation.Name,
				AssetType:      coredata.AssetTypeVirtual,
				Kind:           &item.kind,
				ExternalID:     &item.observation.ExternalID,
				Region:         item.observation.Region,
				Configuration:  item.configuration,
				SchemaVersion:  &item.observation.SchemaVersion,
				LastSeenAt:     &now,
				Status:         coredata.AssetStatusActive,
				ConnectorID:    &account.ConnectorID,
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}

		var existing coredata.Assets
		if err := existing.LoadByDiscoveredIdentities(ctx, tx, scope, assets); err != nil {
			return err
		}

		written, err := assets.UpsertDiscovered(ctx, tx, scope)
		if err != nil {
			return err
		}

		events, err := coredata.DiscoveredAssetChangeEvents(existing, written, sync.ID, now)
		if err != nil {
			return err
		}

		if err := events.Insert(ctx, tx, scope); err != nil {
			return err
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

		return nil
	})
}

func (h *assetSyncHandler) failSync(
	ctx context.Context,
	account *coredata.ConnectorAccount,
	sync *coredata.AssetSync,
	cause error,
) error {
	kind, logged := kindAndCause(cause)
	h.logger.ErrorCtx(
		ctx,
		"cannot list assets",
		log.String("connector_account_id", account.ID.String()),
		log.String("kind", kind),
		log.Error(logged),
	)

	message := assetSyncFailureMessage
	now := time.Now()

	return h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := sync.FinishIfRunning(
			ctx,
			tx,
			coredata.NewScopeFromObjectID(account.ID),
			coredata.AssetSyncStatusFailed,
			&message,
			[]string{},
			now,
		)

		return err
	})
}

func (h *assetSyncHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		syncs, err := coredata.FailStaleAssetSyncs(
			ctx,
			tx,
			time.Now().Add(-h.staleAfter),
			assetSyncStaleMessage,
			h.onlyAccount,
		)
		if err != nil {
			return err
		}

		if len(syncs) == 0 {
			return nil
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
