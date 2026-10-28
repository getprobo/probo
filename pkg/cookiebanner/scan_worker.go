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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp/remote"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/cookiebanner/scanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/netcheck"
)

const (
	// maxRunningScans caps scans in flight on one database. Every probod
	// on that database shares one Chrome, so the cap is not per process.
	maxRunningScans = 1

	// scanHeartbeatInterval keeps updated_at moving between pages. A page
	// that never finishes would otherwise look stale and get paused while
	// this process is still scanning it.
	scanHeartbeatInterval = 30 * time.Second
)

type scanRunner func(ctx context.Context, in scanner.Input) (*scanner.Result, error)

type cookieBannerScanHandler struct {
	svc               *Service
	pg                *pg.Client
	chromeAddr        string
	apiBaseURL        string
	logger            *log.Logger
	run               scanRunner
	maxRunning        int
	heartbeatInterval time.Duration
	shutdownCh        chan struct{}
	shutdownOnce      sync.Once
	now               func() time.Time
}

// ScanWorker claims cookie banner scans and runs them. Run watches ctx so
// a shutdown pauses the scan in flight. The kit worker's own context is
// not cancelled on shutdown, so the pause can still be written.
type ScanWorker struct {
	handler   *cookieBannerScanHandler
	kitWorker *worker.Worker[coredata.CookieBannerScan]
}

func NewCookieBannerScanWorker(
	svc *Service,
	pgClient *pg.Client,
	chromeAddr string,
	apiBaseURL string,
	logger *log.Logger,
	opts ...worker.Option,
) *ScanWorker {
	h := &cookieBannerScanHandler{
		svc:               svc,
		pg:                pgClient,
		chromeAddr:        normalizeChromeAddr(chromeAddr),
		apiBaseURL:        apiBaseURL,
		logger:            logger,
		maxRunning:        maxRunningScans,
		heartbeatInterval: scanHeartbeatInterval,
		shutdownCh:        make(chan struct{}),
		now:               time.Now,
	}
	h.run = h.execute

	// MaxConcurrency stays 1 even when opts try to raise it: one process
	// must not hold two browser contexts, and the cluster cap is enforced
	// again in Claim. Copy opts first so append cannot overwrite the
	// caller's slice.
	kitOpts := make([]worker.Option, len(opts), len(opts)+1)
	copy(kitOpts, opts)
	kitOpts = append(kitOpts, worker.WithMaxConcurrency(1))

	return &ScanWorker{
		handler: h,
		kitWorker: worker.New(
			"cookie-banner-scan-worker",
			h,
			logger,
			kitOpts...,
		),
	}
}

func (w *ScanWorker) Run(ctx context.Context) error {
	context.AfterFunc(ctx, w.handler.signalShutdown)

	return w.kitWorker.Run(ctx)
}

func (h *cookieBannerScanHandler) signalShutdown() {
	h.shutdownOnce.Do(func() { close(h.shutdownCh) })
}

func (h *cookieBannerScanHandler) shutdownSignalled() bool {
	select {
	case <-h.shutdownCh:
		return true
	default:
		return false
	}
}

func (h *cookieBannerScanHandler) Claim(ctx context.Context) (coredata.CookieBannerScan, error) {
	var scan coredata.CookieBannerScan

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.LockCookieBannerScanClaims(ctx, tx); err != nil {
				return err
			}

			running, err := coredata.CountRunningCookieBannerScans(ctx, tx)
			if err != nil {
				return err
			}

			if running >= h.maxRunning {
				return worker.ErrNoTask
			}

			if err := scan.LoadNextClaimableForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return scan.MarkRunning(ctx, tx, coredata.NewScopeFromObjectID(scan.ID), h.now())
		},
	)
	if err != nil {
		if errors.Is(err, worker.ErrNoTask) || errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CookieBannerScan{}, worker.ErrNoTask
		}

		return coredata.CookieBannerScan{}, fmt.Errorf("cannot claim cookie banner scan: %w", err)
	}

	return scan, nil
}

func (h *cookieBannerScanHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.FailStaleCookieBannerScans(ctx, conn, scanner.StaleRunningAfter); err != nil {
				return fmt.Errorf("cannot recover stale cookie banner scans: %w", err)
			}

			return nil
		},
	)
}

func (h *cookieBannerScanHandler) Process(ctx context.Context, scan coredata.CookieBannerScan) error {
	scope := coredata.NewScopeFromObjectID(scan.ID)

	var (
		banner     coredata.CookieBanner
		checkpoint *scanner.Checkpoint
		consent    *scanner.Consent
	)

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banner.LoadByID(ctx, conn, scope, scan.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			decoded, err := unmarshalCheckpoint(scan.Checkpoint)
			if err != nil {
				return err
			}

			checkpoint = decoded

			loaded, err := h.scanConsent(ctx, conn, scope, &banner)
			if err != nil {
				return err
			}

			consent = loaded

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			h.logLost(ctx, scan.ID, "cookie banner scan lost before it started")

			return nil
		}

		h.logger.WarnCtx(
			ctx,
			"cannot prepare cookie banner scan",
			log.Error(err),
			log.String("cookie_banner_scan_id", scan.ID.String()),
		)

		return h.markFailed(ctx, scope, &scan, coredata.CookieBannerScanFailureReasonInternal)
	}

	scanCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-h.shutdownCh:
			cancel()
		case <-scanCtx.Done():
		}
	}()

	var lost atomic.Bool

	heartbeatDone := make(chan struct{})

	go func() {
		defer close(heartbeatDone)

		h.beat(ctx, scanCtx, scope, scan.ID, &lost, cancel)
	}()

	result, runErr := h.run(
		scanCtx,
		scanner.Input{
			Origin:     banner.Origin,
			BannerID:   banner.ID.String(),
			APIBaseURL: h.apiBaseURL,
			Consent:    consent,
			Checkpoint: checkpoint,
			Logger:     h.logger,
			BrowserURL: h.chromeAddr,
			OnPage: func(pageCtx context.Context, page scanner.Page, checkpoint scanner.Checkpoint) error {
				return h.acceptPage(pageCtx, scope, &scan, &banner, page, checkpoint, &lost, cancel)
			},
		},
	)

	cancel()
	<-heartbeatDone

	if lost.Load() {
		h.logLost(ctx, scan.ID, "cookie banner scan lost while running")

		return nil
	}

	switch {
	case result != nil && result.Paused:
		return h.markPaused(ctx, scope, &scan, &result.Checkpoint)
	case runErr != nil && h.shutdownSignalled():
		return h.markPaused(ctx, scope, &scan, nil)
	case runErr != nil:
		if result != nil {
			scan.PagesScanned = result.PagesScanned
		}

		h.logger.WarnCtx(
			ctx,
			"cookie banner scan failed",
			log.Error(runErr),
			log.String("cookie_banner_scan_id", scan.ID.String()),
		)

		return h.markFailed(ctx, scope, &scan, scanFailureReason(runErr))
	case result == nil:
		return h.markFailed(ctx, scope, &scan, coredata.CookieBannerScanFailureReasonInternal)
	default:
		return h.complete(ctx, scope, &scan, result)
	}
}

func (h *cookieBannerScanHandler) execute(ctx context.Context, in scanner.Input) (*scanner.Result, error) {
	// The allocator outlives scan cancellation so a pause can close the
	// tab itself. Cancelling it here would look like a browser failure.
	allocCtx, cancel := remote.NewAllocator(context.WithoutCancel(ctx), h.chromeAddr)
	defer cancel()

	return scanner.Run(ctx, allocCtx, in)
}

func (h *cookieBannerScanHandler) beat(
	ctx context.Context,
	scanCtx context.Context,
	scope coredata.Scoper,
	scanID gid.GID,
	lost *atomic.Bool,
	cancel context.CancelFunc,
) {
	ticker := time.NewTicker(h.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-scanCtx.Done():
			return
		case <-ticker.C:
			err := h.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					beat := coredata.CookieBannerScan{ID: scanID}

					return beat.Touch(ctx, conn, scope, h.now())
				},
			)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				lost.Store(true)
				cancel()

				return
			}

			if err != nil {
				h.logger.WarnCtx(
					ctx,
					"cannot touch cookie banner scan",
					log.Error(err),
					log.String("cookie_banner_scan_id", scanID.String()),
				)
			}
		}
	}
}

func (h *cookieBannerScanHandler) acceptPage(
	ctx context.Context,
	scope coredata.Scoper,
	scan *coredata.CookieBannerScan,
	banner *coredata.CookieBanner,
	page scanner.Page,
	checkpoint scanner.Checkpoint,
	lost *atomic.Bool,
	cancel context.CancelFunc,
) error {
	raw, err := marshalCheckpoint(checkpoint)
	if err != nil {
		return err
	}

	scan.Checkpoint = raw

	err = h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := scan.SaveCheckpoint(ctx, tx, scope, h.now()); err != nil {
				return err
			}

			if _, err := h.svc.ingestScanPage(ctx, tx, banner, scan.ID, page); err != nil {
				return err
			}

			return nil
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		lost.Store(true)
		cancel()

		return nil
	}

	if err != nil {
		return err
	}

	// The scanner returns no result when a later page fails to ingest, so
	// MarkFailed can only keep the progress recorded on the scan itself.
	scan.PagesScanned = checkpoint.PagesScanned

	return nil
}

func (h *cookieBannerScanHandler) scanConsent(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	banner *coredata.CookieBanner,
) (*scanner.Consent, error) {
	if banner.State != coredata.CookieBannerStateActive {
		return nil, nil
	}

	var version coredata.CookieBannerVersion
	if err := version.LoadLatestPublishedByCookieBannerID(ctx, conn, scope, banner.ID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load published cookie banner version: %w", err)
	}

	snapshot, err := version.GetSnapshot()
	if err != nil {
		return nil, fmt.Errorf("cannot read cookie banner version snapshot: %w", err)
	}

	slugs := make([]string, 0, len(snapshot.Categories))
	for _, category := range snapshot.Categories {
		if category.Kind == coredata.CookieCategoryKindUncategorised || category.Slug == "" {
			continue
		}

		slugs = append(slugs, category.Slug)
	}

	return &scanner.Consent{
		Version:       version.Version,
		CategorySlugs: slugs,
	}, nil
}

func (h *cookieBannerScanHandler) complete(
	ctx context.Context,
	scope coredata.Scoper,
	scan *coredata.CookieBannerScan,
	result *scanner.Result,
) error {
	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cookies, storage, err := new(coredata.DetectedTrackers).CountByCookieBannerScanID(ctx, tx, scope, scan.ID)
			if err != nil {
				return err
			}

			resources, err := new(coredata.TrackerResources).CountByCookieBannerScanID(ctx, tx, scope, scan.ID)
			if err != nil {
				return err
			}

			scan.Truncated = result.Truncated
			scan.PagesScanned = result.PagesScanned
			scan.CookiesFound = cookies
			scan.StorageFound = storage
			scan.ResourcesFound = resources

			return scan.MarkCompleted(ctx, tx, scope, h.now())
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		h.logLost(ctx, scan.ID, "cookie banner scan lost before it could complete")

		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot complete cookie banner scan: %w", err)
	}

	return nil
}

func (h *cookieBannerScanHandler) markPaused(
	ctx context.Context,
	scope coredata.Scoper,
	scan *coredata.CookieBannerScan,
	checkpoint *scanner.Checkpoint,
) error {
	if checkpoint != nil {
		raw, err := marshalCheckpoint(*checkpoint)
		if err != nil {
			return err
		}

		scan.Checkpoint = raw
	}

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return scan.MarkPaused(ctx, conn, scope, h.now())
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		h.logLost(ctx, scan.ID, "cookie banner scan lost before it could pause")

		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot pause cookie banner scan: %w", err)
	}

	return nil
}

func (h *cookieBannerScanHandler) markFailed(
	ctx context.Context,
	scope coredata.Scoper,
	scan *coredata.CookieBannerScan,
	reason coredata.CookieBannerScanFailureReason,
) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return scan.MarkFailed(ctx, conn, scope, reason, h.now())
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		h.logLost(ctx, scan.ID, "cookie banner scan lost before it could fail")

		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot fail cookie banner scan: %w", err)
	}

	return nil
}

func (h *cookieBannerScanHandler) logLost(ctx context.Context, scanID gid.GID, msg string) {
	h.logger.InfoCtx(
		ctx,
		msg,
		log.String("cookie_banner_scan_id", scanID.String()),
	)
}

func scanFailureReason(err error) coredata.CookieBannerScanFailureReason {
	switch {
	case errors.Is(err, scanner.ErrNoPageLoaded):
		return coredata.CookieBannerScanFailureReasonNoPagesLoaded
	case errors.Is(err, scanner.ErrBlocked), errors.Is(err, netcheck.ErrBlockedAddress):
		return coredata.CookieBannerScanFailureReasonBlocked
	case errors.Is(err, scanner.ErrUnreachable):
		return coredata.CookieBannerScanFailureReasonUnreachable
	default:
		return coredata.CookieBannerScanFailureReasonInternal
	}
}

func marshalCheckpoint(checkpoint scanner.Checkpoint) (*json.RawMessage, error) {
	raw, err := json.Marshal(checkpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot encode scan checkpoint: %w", err)
	}

	encoded := json.RawMessage(raw)

	return &encoded, nil
}

func unmarshalCheckpoint(raw *json.RawMessage) (*scanner.Checkpoint, error) {
	if raw == nil || len(*raw) == 0 {
		return nil, nil
	}

	var checkpoint scanner.Checkpoint
	if err := json.Unmarshal(*raw, &checkpoint); err != nil {
		return nil, fmt.Errorf("cannot decode scan checkpoint: %w", err)
	}

	return &checkpoint, nil
}

func normalizeChromeAddr(addr string) string {
	if addr == "" || strings.Contains(addr, "://") {
		return addr
	}

	return "ws://" + addr
}
