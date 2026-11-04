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
	"errors"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type cmpDomainTask struct {
	domain *coredata.CMPDomain
}

type cmpDomainHandler struct {
	svc    *Service
	pg     *pg.Client
	logger *log.Logger
}

func NewCMPDomainWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[cmpDomainTask] {
	h := &cmpDomainHandler{
		svc:    svc,
		pg:     pgClient,
		logger: logger,
	}

	opts = append([]worker.Option{worker.WithMaxConcurrency(1)}, opts...)

	return worker.New(
		"cmp-domain-worker",
		h,
		logger,
		opts...,
	)
}

func (h *cmpDomainHandler) Claim(ctx context.Context) (cmpDomainTask, error) {
	var task cmpDomainTask

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var domain coredata.CMPDomain
			if err := domain.LoadNextForSourcePromotionForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			if err := domain.ClearSourcePromotionRequested(ctx, tx); err != nil {
				return err
			}

			task.domain = &domain

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return cmpDomainTask{}, worker.ErrNoTask
		}

		return cmpDomainTask{}, err
	}

	return task, nil
}

func (h *cmpDomainHandler) Process(ctx context.Context, task cmpDomainTask) error {
	if task.domain == nil {
		return nil
	}

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return h.svc.promoteSiblingCookieSources(ctx, tx, task.domain)
		},
	)
	if err != nil {
		if retryErr := h.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return task.domain.SetSourcePromotionRequested(ctx, tx)
			},
		); retryErr != nil {
			h.logger.ErrorCtx(ctx, "cannot requeue cmp domain source promotion", log.Error(retryErr))
		}

		return err
	}

	return nil
}
