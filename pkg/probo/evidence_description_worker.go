// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/evidencedescriber"
	"go.probo.inc/probo/pkg/filemanager"
)

type (
	EvidenceDescriptionWorker struct {
		pg             *pg.Client
		fileManager    *filemanager.Service
		describer      *evidencedescriber.Describer
		logger         *log.Logger
		interval       time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	EvidenceDescriptionWorkerOption func(*EvidenceDescriptionWorker)
)

func WithEvidenceDescriptionWorkerInterval(d time.Duration) EvidenceDescriptionWorkerOption {
	return func(w *EvidenceDescriptionWorker) { w.interval = d }
}

func WithEvidenceDescriptionWorkerStaleAfter(d time.Duration) EvidenceDescriptionWorkerOption {
	return func(w *EvidenceDescriptionWorker) { w.staleAfter = d }
}

func WithEvidenceDescriptionWorkerMaxConcurrency(n int) EvidenceDescriptionWorkerOption {
	return func(w *EvidenceDescriptionWorker) {
		if n > 0 {
			w.maxConcurrency = n
		}
	}
}

func NewEvidenceDescriptionWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	describer *evidencedescriber.Describer,
	logger *log.Logger,
	opts ...EvidenceDescriptionWorkerOption,
) *EvidenceDescriptionWorker {
	w := &EvidenceDescriptionWorker{
		pg:             pgClient,
		fileManager:    fileManager,
		describer:      describer,
		logger:         logger,
		interval:       10 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 10,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *EvidenceDescriptionWorker) Run(ctx context.Context) error {
	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, w.maxConcurrency)
		ticker = time.NewTicker(w.interval)
	)
	defer ticker.Stop()
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			nonCancelableCtx := context.WithoutCancel(ctx)
			w.recoverStaleRows(nonCancelableCtx)

			for {
				if err := w.processNext(ctx, sem, &wg); err != nil {
					if !errors.Is(err, coredata.ErrResourceNotFound) {
						w.logger.ErrorCtx(nonCancelableCtx, "cannot claim evidence for description", log.Error(err))
					}
					break
				}
			}
		}
	}
}

func (w *EvidenceDescriptionWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		evidence         = coredata.Evidence{}
		now              = time.Now()
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := evidence.LoadNextPendingDescriptionForUpdateSkipLocked(
				nonCancelableCtx,
				tx,
			); err != nil {
				return err
			}

			evidence.DescriptionStatus = coredata.EvidenceDescriptionStatusProcessing
			evidence.DescriptionProcessingStartedAt = &now
			evidence.UpdatedAt = now
			if err := evidence.Update(nonCancelableCtx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(evidence coredata.Evidence) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.describeAndCommit(nonCancelableCtx, &evidence); err != nil {
			w.logger.ErrorCtx(
				nonCancelableCtx,
				"evidence description worker failure",
				log.Error(err),
				log.String("evidence_id", evidence.ID.String()),
			)

			if err := w.failEvidence(nonCancelableCtx, &evidence); err != nil {
				w.logger.ErrorCtx(nonCancelableCtx, "cannot mark evidence description as failed", log.Error(err))
			}
		}
	}(evidence)

	return nil
}

func (w *EvidenceDescriptionWorker) describeAndCommit(
	ctx context.Context,
	evidence *coredata.Evidence,
) error {
	if evidence.EvidenceFileId == nil {
		return fmt.Errorf("evidence %s has no file", evidence.ID)
	}

	scope := coredata.NewScopeFromObjectID(evidence.ID)

	var file coredata.File
	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := file.LoadByID(ctx, conn, scope, *evidence.EvidenceFileId); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("cannot load file: %w", err)
	}

	base64Data, mimeType, err := w.fileManager.GetFileBase64(ctx, &file)
	if err != nil {
		return fmt.Errorf("cannot download file: %w", err)
	}

	description, err := w.describer.Describe(ctx, file.FileName, mimeType, base64Data)
	if err != nil {
		return fmt.Errorf("cannot describe evidence: %w", err)
	}

	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			evidence.Description = description
			evidence.DescriptionStatus = coredata.EvidenceDescriptionStatusCompleted
			evidence.DescriptionProcessingStartedAt = nil
			evidence.UpdatedAt = time.Now()
			if err := evidence.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	)
}

func (w *EvidenceDescriptionWorker) failEvidence(
	ctx context.Context,
	evidence *coredata.Evidence,
) error {
	scope := coredata.NewScopeFromObjectID(evidence.ID)

	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			evidence.DescriptionStatus = coredata.EvidenceDescriptionStatusFailed
			evidence.DescriptionProcessingStartedAt = nil
			evidence.UpdatedAt = time.Now()
			if err := evidence.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	)
}

func (w *EvidenceDescriptionWorker) recoverStaleRows(ctx context.Context) {
	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := coredata.ResetStaleDescriptionProcessing(ctx, conn, w.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale description processing: %w", err)
			}
			return nil
		},
	); err != nil {
		w.logger.ErrorCtx(ctx, "cannot recover stale evidence descriptions", log.Error(err))
	}
}
