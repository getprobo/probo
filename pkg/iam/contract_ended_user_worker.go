// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package iam

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type ContractEndedUserWorker = worker.Worker[struct{}]

type contractEndedUserHandler struct {
	pg        *pg.Client
	service   *OrganizationService
	logger    *log.Logger
	batchSize int
	lastRunAt atomic.Int64
}

const (
	DefaultContractEndedUserWorkerInterval  = time.Hour
	DefaultContractEndedUserWorkerBatchSize = 100
)

func NewContractEndedUserWorker(
	pgClient *pg.Client,
	service *OrganizationService,
	logger *log.Logger,
	opts ...worker.Option,
) *ContractEndedUserWorker {
	h := &contractEndedUserHandler{
		pg:        pgClient,
		service:   service,
		logger:    logger.Named("contract-ended-user-worker"),
		batchSize: DefaultContractEndedUserWorkerBatchSize,
	}

	return worker.New(
		"contract-ended-user-worker",
		h,
		logger,
		append(
			[]worker.Option{
				worker.WithInterval(DefaultContractEndedUserWorkerInterval),
				worker.WithMaxConcurrency(1),
			},
			opts...,
		)...,
	)
}

func (h *contractEndedUserHandler) Claim(_ context.Context) (struct{}, error) {
	now := time.Now().UnixNano()
	last := h.lastRunAt.Load()

	if last > 0 && now-last < int64(DefaultContractEndedUserWorkerInterval) {
		return struct{}{}, worker.ErrNoTask
	}

	if !h.lastRunAt.CompareAndSwap(last, now) {
		return struct{}{}, worker.ErrNoTask
	}

	return struct{}{}, nil
}

func (h *contractEndedUserHandler) Process(ctx context.Context, _ struct{}) error {
	now := time.Now()
	total, err := h.deactivateContractEndedUsers(ctx, now)
	if err != nil {
		return err
	}

	h.logger.InfoCtx(
		ctx,
		"contract-ended user worker completed",
		log.Int("users_deactivated", total),
	)

	return nil
}

func (h *contractEndedUserHandler) deactivateContractEndedUsers(ctx context.Context, now time.Time) (int, error) {
	total := 0

	for {
		profiles, err := h.loadContractEndedActiveProfiles(ctx, currentDate(now))
		if err != nil {
			return 0, fmt.Errorf("cannot load contract-ended active profiles: %w", err)
		}

		if len(profiles) == 0 {
			return total, nil
		}

		for _, profile := range profiles {
			deactivated, err := h.service.DeactivateUserForContractEnd(ctx, profile.ID, now)
			if err != nil {
				return total, fmt.Errorf("cannot deactivate contract-ended user: %w", err)
			}

			if deactivated {
				total++
			}
		}

		if len(profiles) < h.batchSize {
			return total, nil
		}
	}
}

func (h *contractEndedUserHandler) loadContractEndedActiveProfiles(
	ctx context.Context,
	currentDate time.Time,
) (coredata.MembershipProfiles, error) {
	var profiles coredata.MembershipProfiles

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return profiles.LoadContractEndedActive(ctx, conn, h.batchSize, currentDate)
		},
	)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}
