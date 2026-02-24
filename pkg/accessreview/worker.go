package accessreview

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type SourceFetchWorker struct {
	pg                    *pg.Client
	tenantRuntimeProvider TenantRuntimeProvider
	logger                *log.Logger
	interval              time.Duration
	maxConcurrency        int
}

type SourceFetchWorkerOption func(*SourceFetchWorker)

func WithSourceFetchWorkerInterval(interval time.Duration) SourceFetchWorkerOption {
	return func(w *SourceFetchWorker) {
		if interval > 0 {
			w.interval = interval
		}
	}
}

func WithSourceFetchWorkerMaxConcurrency(maxConcurrency int) SourceFetchWorkerOption {
	return func(w *SourceFetchWorker) {
		if maxConcurrency > 0 {
			w.maxConcurrency = maxConcurrency
		}
	}
}

func newSourceFetchWorker(
	pgClient *pg.Client,
	tenantRuntimeProvider TenantRuntimeProvider,
	logger *log.Logger,
	opts ...SourceFetchWorkerOption,
) *SourceFetchWorker {
	worker := &SourceFetchWorker{
		pg:                    pgClient,
		tenantRuntimeProvider: tenantRuntimeProvider,
		logger:                logger,
		interval:              30 * time.Second,
		maxConcurrency:        1,
	}

	for _, opt := range opts {
		opt(worker)
	}

	return worker
}

func (w *SourceFetchWorker) Run(ctx context.Context) error {
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, w.maxConcurrency)
	)
	defer wg.Wait()

LOOP:
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(w.interval):
		for {
			if err := w.processNext(ctx, sem, &wg); err != nil {
				if !errors.Is(err, coredata.ErrNoAccessReviewCampaignSourceFetchAvailable) && !errors.Is(err, context.Canceled) {
					w.logger.ErrorCtx(ctx, "cannot process access review campaign source fetch", log.Error(err))
				}
				break
			}
		}

		goto LOOP
	}
}

func (w *SourceFetchWorker) processNext(
	ctx context.Context,
	sem chan struct{},
	wg *sync.WaitGroup,
) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Once a source fetch is claimed we should complete the lifecycle updates.
	nonCancelableCtx := context.WithoutCancel(ctx)

	sourceFetch, err := w.lockSourceFetch(nonCancelableCtx)
	if err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(sourceFetch coredata.AccessReviewCampaignSourceFetch) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.processSourceFetch(nonCancelableCtx, &sourceFetch); err != nil {
			w.logger.ErrorCtx(nonCancelableCtx, "cannot process access review campaign source fetch", log.Error(err))
		}
	}(*sourceFetch)

	return nil
}

func (w *SourceFetchWorker) processSourceFetch(ctx context.Context, sourceFetch *coredata.AccessReviewCampaignSourceFetch) error {
	tenantRuntime := w.tenantRuntimeProvider(sourceFetch.TenantID)
	if tenantRuntime == nil {
		return fmt.Errorf("tenant runtime provider returned nil for tenant %s", sourceFetch.TenantID)
	}

	campaign, err := tenantRuntime.AccessReviewCampaigns().Get(ctx, sourceFetch.AccessReviewCampaignID)
	if err != nil {
		commitErr := w.commitFailedSourceFetch(
			ctx,
			sourceFetch,
			fmt.Errorf("cannot load campaign: %w", err),
		)
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign: %w, and cannot commit failed source fetch: %w", err, commitErr)
		}
		return fmt.Errorf("cannot load campaign: %w", err)
	}

	count, err := tenantRuntime.SnapshotSource(ctx, campaign, sourceFetch.AccessSourceID)
	if err != nil {
		commitErr := w.commitFailedSourceFetch(ctx, sourceFetch, err)
		if commitErr != nil {
			return fmt.Errorf("cannot snapshot source: %w, and cannot commit failed source fetch: %w", err, commitErr)
		}

		if finalizeErr := w.finalizeCampaignFetchLifecycle(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID); finalizeErr != nil {
			return fmt.Errorf("cannot finalize campaign after failed source fetch: %w", finalizeErr)
		}
		return fmt.Errorf("cannot snapshot source: %w", err)
	}

	if err := w.commitSuccessfulSourceFetch(ctx, sourceFetch, count); err != nil {
		return fmt.Errorf("cannot commit successful source fetch: %w", err)
	}

	if err := w.finalizeCampaignFetchLifecycle(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID); err != nil {
		return fmt.Errorf("cannot finalize campaign fetch lifecycle: %w", err)
	}

	return nil
}

func (w *SourceFetchWorker) lockSourceFetch(ctx context.Context) (*coredata.AccessReviewCampaignSourceFetch, error) {
	var (
		sourceFetch = &coredata.AccessReviewCampaignSourceFetch{}
	)

	err := w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := sourceFetch.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot load next queued source fetch: %w", err)
			}

			now := time.Now()
			sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusFetching
			sourceFetch.AttemptCount++
			sourceFetch.LastError = nil
			sourceFetch.StartedAt = ref.Ref(now)
			sourceFetch.CompletedAt = nil
			sourceFetch.UpdatedAt = now

			scope := coredata.NewScope(sourceFetch.TenantID)
			if err := sourceFetch.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update source fetch to FETCHING: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return sourceFetch, nil
}

func (w *SourceFetchWorker) commitFailedSourceFetch(
	ctx context.Context,
	sourceFetch *coredata.AccessReviewCampaignSourceFetch,
	failureErr error,
) error {
	var (
		now    = time.Now()
		errMsg = failureErr.Error()
		scope  = coredata.NewScopeFromObjectID(sourceFetch.AccessReviewCampaignID)
	)

	sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusFailed
	sourceFetch.LastError = &errMsg
	sourceFetch.CompletedAt = ref.Ref(now)
	sourceFetch.UpdatedAt = now

	return w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := sourceFetch.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update source fetch as FAILED: %w", err)
			}

			return nil
		},
	)
}

func (w *SourceFetchWorker) commitSuccessfulSourceFetch(
	ctx context.Context,
	sourceFetch *coredata.AccessReviewCampaignSourceFetch,
	fetchedAccountsCount int,
) error {
	var (
		now   = time.Now()
		scope = coredata.NewScopeFromObjectID(sourceFetch.AccessReviewCampaignID)
	)

	sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusSuccess
	sourceFetch.FetchedAccountsCount = fetchedAccountsCount
	sourceFetch.LastError = nil
	sourceFetch.CompletedAt = ref.Ref(now)
	sourceFetch.UpdatedAt = now

	return w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := sourceFetch.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update source fetch as SUCCESS: %w", err)
			}
			return nil
		},
	)
}

func (w *SourceFetchWorker) finalizeCampaignFetchLifecycle(
	ctx context.Context,
	tenantID gid.TenantID,
	campaignID gid.GID,
) error {
	var (
		scope         = coredata.NewScope(tenantID)
		tenantRuntime = w.tenantRuntimeProvider(tenantID)
	)
	if tenantRuntime == nil {
		return fmt.Errorf("tenant runtime provider returned nil for tenant %s", tenantID)
	}

	var (
		readyForFinalization bool
		hasFailure           bool
		campaign             *coredata.AccessReviewCampaign
	)

	err := w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			fetches := coredata.AccessReviewCampaignSourceFetches{}
			if err := fetches.LoadByCampaignID(ctx, tx, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load source fetches: %w", err)
			}
			if len(fetches) == 0 {
				return nil
			}

			for _, fetch := range fetches {
				if !fetch.Status.IsTerminal() {
					return nil
				}
				if fetch.Status == coredata.AccessReviewCampaignSourceFetchStatusFailed {
					hasFailure = true
				}
			}

			campaign = &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, tx, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status == coredata.AccessReviewCampaignStatusCompleted ||
				campaign.Status == coredata.AccessReviewCampaignStatusCancelled ||
				campaign.Status == coredata.AccessReviewCampaignStatusPendingActions ||
				campaign.Status == coredata.AccessReviewCampaignStatusFailed {
				return nil
			}

			readyForFinalization = true
			return nil
		},
	)
	if err != nil {
		return err
	}
	if !readyForFinalization {
		return nil
	}

	if hasFailure {
		return w.markCampaignFailed(ctx, tenantID, campaignID)
	}

	if err := tenantRuntime.Diff(ctx, campaign); err != nil {
		if failErr := w.markCampaignFailed(ctx, tenantID, campaignID); failErr != nil {
			return fmt.Errorf("cannot diff campaign: %w, and cannot mark campaign failed: %w", err, failErr)
		}
		return fmt.Errorf("cannot diff campaign: %w", err)
	}

	return w.pg.WithTx(ctx, func(tx pg.Conn) error {
		c := &coredata.AccessReviewCampaign{}
		if err := c.LoadByID(ctx, tx, scope, campaignID); err != nil {
			return fmt.Errorf("cannot reload campaign: %w", err)
		}
		if c.Status != coredata.AccessReviewCampaignStatusInProgress {
			return nil
		}

		c.Status = coredata.AccessReviewCampaignStatusPendingActions
		c.UpdatedAt = time.Now()
		if err := c.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot set campaign status to pending actions: %w", err)
		}

		return nil
	})
}

func (w *SourceFetchWorker) markCampaignFailed(ctx context.Context, tenantID gid.TenantID, campaignID gid.GID) error {
	scope := coredata.NewScope(tenantID)

	return w.pg.WithTx(ctx, func(conn pg.Conn) error {
		campaign := &coredata.AccessReviewCampaign{}
		if err := campaign.LoadByID(ctx, conn, scope, campaignID); err != nil {
			return fmt.Errorf("cannot load campaign for failure update: %w", err)
		}
		campaign.Status = coredata.AccessReviewCampaignStatusFailed
		campaign.UpdatedAt = time.Now()
		if err := campaign.Update(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot mark campaign as failed: %w", err)
		}

		return nil
	})
}
