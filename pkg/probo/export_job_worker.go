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

package probo

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

const defaultExportJobStaleAfter = 25 * time.Minute

type (
	exportJobHandler struct {
		service    *Service
		logger     *log.Logger
		staleAfter time.Duration
	}

	ExportJobWorkerConfig struct {
		StaleAfter time.Duration
	}
)

var (
	_ worker.Handler[coredata.ExportJob] = (*exportJobHandler)(nil)
	_ worker.StaleRecoverer              = (*exportJobHandler)(nil)
)

func NewExportJobWorker(
	service *Service,
	logger *log.Logger,
	cfg ExportJobWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.ExportJob] {
	staleAfter := cfg.StaleAfter
	if staleAfter <= 0 {
		staleAfter = defaultExportJobStaleAfter
	}

	h := &exportJobHandler{
		service:    service,
		logger:     logger,
		staleAfter: staleAfter,
	}

	return worker.New(
		"export-job-worker",
		h,
		logger,
		opts...,
	)
}

func (h *exportJobHandler) Claim(ctx context.Context) (coredata.ExportJob, error) {
	exportJob, err := h.service.lockExportJob(ctx)
	if err != nil {
		if errors.Is(err, coredata.ErrNoExportJobAvailable) {
			return coredata.ExportJob{}, worker.ErrNoTask
		}

		return coredata.ExportJob{}, err
	}

	return *exportJob, nil
}

func (h *exportJobHandler) Process(ctx context.Context, exportJob coredata.ExportJob) error {
	stopHeartbeat := h.startHeartbeat(ctx, exportJob.ID)
	defer stopHeartbeat()

	if err := h.service.processExportJob(ctx, &exportJob); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"export job worker failure",
			log.Error(err),
			log.String("export_job_id", exportJob.ID.String()),
			log.String("export_job_type", exportJob.Type.String()),
		)

		return err
	}

	return nil
}

func (h *exportJobHandler) RecoverStale(ctx context.Context) error {
	return h.service.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleExportJobs(ctx, conn, h.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale export jobs: %w", err)
			}

			return nil
		},
	)
}

func (h *exportJobHandler) startHeartbeat(ctx context.Context, exportJobID gid.GID) func() {
	done := make(chan struct{})

	interval := max(h.staleAfter/2, time.Second)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := h.service.pg.WithConn(
					ctx,
					func(ctx context.Context, conn pg.Querier) error {
						return coredata.TouchExportJobLease(
							ctx,
							conn,
							coredata.NewScope(exportJobID.TenantID()),
							exportJobID,
						)
					},
				); err != nil {
					h.logger.ErrorCtx(
						ctx,
						"cannot renew export job lease",
						log.Error(err),
						log.String("export_job_id", exportJobID.String()),
					)
				}
			}
		}
	}()

	return func() { close(done) }
}
