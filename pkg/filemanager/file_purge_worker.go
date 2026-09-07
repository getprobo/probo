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

package filemanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	DefaultDeletedFileRetention     = 30 * 24 * time.Hour
	DefaultDeletedFileCleanupPeriod = time.Hour
	DefaultDeletedFileCleanupBatch  = 100
	DefaultDeletedFileMaxPerTick    = 1000
)

type (
	DeletedFilePurgeWorker struct {
		handler  *deletedFilePurgeHandler
		periodic *worker.PeriodicWorker
		logger   *log.Logger
	}

	deletedFilePurgeHandler struct {
		svc        *Service
		logger     *log.Logger
		scope      coredata.Scoper
		retention  time.Duration
		batchSize  int
		maxPerTick int
		now        func() time.Time
	}
)

func NewDeletedFilePurgeWorker(
	svc *Service,
	logger *log.Logger,
	retention time.Duration,
	maxPerTick int,
	opts ...worker.Option,
) *DeletedFilePurgeWorker {
	if retention <= 0 {
		retention = DefaultDeletedFileRetention
	}

	if maxPerTick <= 0 {
		maxPerTick = DefaultDeletedFileMaxPerTick
	}

	workerOpts := append(
		[]worker.Option{worker.WithInterval(DefaultDeletedFileCleanupPeriod)},
		opts...,
	)

	handler := &deletedFilePurgeHandler{
		svc:        svc,
		logger:     logger,
		scope:      coredata.NewNoScope(),
		retention:  retention,
		batchSize:  DefaultDeletedFileCleanupBatch,
		maxPerTick: maxPerTick,
		now:        time.Now,
	}

	return &DeletedFilePurgeWorker{
		handler: handler,
		periodic: worker.NewPeriodic(
			"file-purge-worker",
			handler,
			logger,
			workerOpts...,
		),
		logger: logger,
	}
}

func (w *DeletedFilePurgeWorker) Run(ctx context.Context) error {
	if err := w.handler.Run(context.WithoutCancel(ctx)); err != nil {
		w.logger.ErrorCtx(
			ctx,
			"run failed",
			log.Error(err),
		)
	}

	return w.periodic.Run(ctx)
}

func (h *deletedFilePurgeHandler) Run(ctx context.Context) error {
	before := h.now().Add(-h.retention)

	var afterID *gid.GID

	deleted := 0
	remaining := h.maxPerTick

	var errs []error

	for remaining > 0 {
		limit := min(h.batchSize, remaining)

		var files coredata.Files

		if err := h.svc.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return files.LoadDeletedBefore(
					ctx,
					conn,
					h.scope,
					before,
					afterID,
					limit,
				)
			},
		); err != nil {
			return fmt.Errorf("cannot load deleted files: %w", err)
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			id := file.ID
			afterID = &id
			remaining--

			if err := h.purge(ctx, file, before); err != nil {
				h.logger.ErrorCtx(
					ctx,
					"cannot purge deleted file",
					log.String("file_id", file.ID.String()),
					log.Error(err),
				)
				errs = append(errs, err)

				continue
			}

			deleted++
		}
	}

	if deleted > 0 {
		h.logger.InfoCtx(
			ctx,
			"purged soft-deleted files",
			log.Int("count", deleted),
		)
	}

	return errors.Join(errs...)
}

func (h *deletedFilePurgeHandler) purge(
	ctx context.Context,
	file *coredata.File,
	before time.Time,
) error {
	return h.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var locked coredata.File

			if err := locked.LoadDeletedByIDForUpdateSkipLocked(ctx, tx, h.scope, file.ID, before); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot lock deleted file: %w", err)
			}

			if err := locked.Delete(ctx, tx, h.scope); err != nil {
				return fmt.Errorf("cannot delete file from database: %w", err)
			}

			if err := h.svc.DeleteFile(ctx, &locked); err != nil {
				return fmt.Errorf("cannot delete file from S3: %w", err)
			}

			return nil
		},
	)
}
