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

package mailman

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	MailingListWorker struct {
		service        *Service
		pg             *pg.Client
		logger         *log.Logger
		interval       time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	MailingListWorkerOption func(*MailingListWorker)
)

func WithMailingListWorkerInterval(d time.Duration) MailingListWorkerOption {
	return func(w *MailingListWorker) { w.interval = d }
}

func WithMailingListWorkerStaleAfter(d time.Duration) MailingListWorkerOption {
	return func(w *MailingListWorker) { w.staleAfter = d }
}

func WithMailingListWorkerMaxConcurrency(n int) MailingListWorkerOption {
	return func(w *MailingListWorker) {
		if n > 0 {
			w.maxConcurrency = n
		}
	}
}

func NewMailingListWorker(
	service *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...MailingListWorkerOption,
) *MailingListWorker {
	w := &MailingListWorker{
		service:        service,
		pg:             pgClient,
		logger:         logger,
		interval:       10 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 5,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *MailingListWorker) Run(ctx context.Context) error {
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
		// From there we should not accept cancellations anymore.
		nonCancelableCtx := context.WithoutCancel(ctx)
		w.recoverStaleRows(nonCancelableCtx)

		for {
			if err := w.processNext(ctx, sem, &wg); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					w.logger.ErrorCtx(nonCancelableCtx, "cannot claim mailing list update", log.Error(err))
				}
				break
			}
		}

		goto LOOP
	}
}

func (w *MailingListWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		mlu              coredata.MailingListUpdate
		now              = time.Now()
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := mlu.LoadNextEnqueuedForUpdateSkipLocked(nonCancelableCtx, tx); err != nil {
				return err
			}

			scope := coredata.NewScopeFromObjectID(mlu.ID)
			mlu.Status = coredata.MailingListUpdateStatusProcessing
			mlu.UpdatedAt = now

			if err := mlu.Update(nonCancelableCtx, tx, scope); err != nil {
				return fmt.Errorf("cannot claim mailing list update: %w", err)
			}

			return nil
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(mlu coredata.MailingListUpdate) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.sendAndCommit(nonCancelableCtx, &mlu); err != nil {
			w.logger.ErrorCtx(nonCancelableCtx, "cannot send mailing list update",
				log.Error(err),
				log.String("mailing_list_update_id", mlu.ID.String()),
			)

			if err := w.resetEnqueued(nonCancelableCtx, &mlu); err != nil {
				w.logger.ErrorCtx(nonCancelableCtx, "cannot reset mailing list update to enqueued",
					log.Error(err),
					log.String("mailing_list_update_id", mlu.ID.String()),
				)
			}
		}
	}(mlu)

	return nil
}

func (w *MailingListWorker) sendAndCommit(ctx context.Context, mlu *coredata.MailingListUpdate) error {
	if err := w.service.CreateUpdateEmails(ctx, mlu.MailingListID, mlu.ID, mlu.Title, mlu.Body); err != nil {
		return fmt.Errorf("cannot create update emails: %w", err)
	}

	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			scope := coredata.NewScopeFromObjectID(mlu.ID)

			var current coredata.MailingListUpdate
			if err := current.LoadByID(ctx, tx, scope, mlu.ID); err != nil {
				return fmt.Errorf("cannot reload mailing list update: %w", err)
			}

			if current.Status != coredata.MailingListUpdateStatusProcessing {
				return fmt.Errorf("unexpected status %s, expected PROCESSING", current.Status)
			}

			mlu.Status = coredata.MailingListUpdateStatusSent
			mlu.UpdatedAt = time.Now()

			if err := mlu.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot mark mailing list update as sent: %w", err)
			}

			return nil
		},
	)
}

func (w *MailingListWorker) resetEnqueued(ctx context.Context, mlu *coredata.MailingListUpdate) error {
	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			scope := coredata.NewScopeFromObjectID(mlu.ID)
			mlu.Status = coredata.MailingListUpdateStatusEnqueued
			mlu.UpdatedAt = time.Now()

			if err := mlu.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot reset mailing list update: %w", err)
			}

			return nil
		},
	)
}

func (w *MailingListWorker) recoverStaleRows(ctx context.Context) {
	err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := coredata.ResetStaleProcessingMailingListUpdates(ctx, conn, w.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale processing mailing list updates: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		w.logger.ErrorCtx(ctx, "cannot recover stale processing mailing list updates", log.Error(err))
	}
}
