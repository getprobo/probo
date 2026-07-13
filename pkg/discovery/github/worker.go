// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const errorMessageMaxLen = 512

type discoveryHandler struct {
	pg     *pg.Client
	runner *Runner
	logger *log.Logger
}

func NewDiscoveryWorker(
	pgClient *pg.Client,
	runner *Runner,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AgentRun] {
	h := &discoveryHandler{
		pg:     pgClient,
		runner: runner,
		logger: logger,
	}

	return worker.New(
		"github-discovery-worker",
		h,
		logger,
		opts...,
	)
}

func (h *discoveryHandler) Claim(ctx context.Context) (coredata.AgentRun, error) {
	var run coredata.AgentRun

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadNextPendingGitHubDiscoveryForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			run.Status = coredata.AgentRunStatusRunning
			run.StartedAt = &now
			run.UpdatedAt = now

			if err := run.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot mark github discovery run running: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.AgentRun{}, worker.ErrNoTask
		}

		return coredata.AgentRun{}, err
	}

	return run, nil
}

func (h *discoveryHandler) Process(ctx context.Context, run coredata.AgentRun) error {
	result, runErr := h.runner.Run(ctx, &run)

	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.Result = nil
	run.ErrorMessage = nil

	if runErr == nil {
		run.Status = coredata.AgentRunStatusCompleted

		data, err := json.Marshal(result)
		if err != nil {
			runErr = fmt.Errorf("cannot marshal github discovery result: %w", err)
		} else {
			run.Result = data
		}
	}

	if runErr != nil {
		run.Status = coredata.AgentRunStatusFailed

		h.logger.ErrorCtx(
			context.WithoutCancel(ctx),
			"github discovery run failed",
			log.String("run_id", run.ID.String()),
			log.Error(runErr),
		)

		msg := sanitizeError(runErr)
		run.ErrorMessage = &msg
	}

	return h.pg.WithTx(
		context.WithoutCancel(ctx),
		func(ctx context.Context, tx pg.Tx) error {
			rowsAffected, err := coredata.CommitAgentRunResult(ctx, tx, &run)
			if err != nil {
				return err
			}

			if rowsAffected == 0 {
				h.logger.WarnCtx(
					ctx,
					"github discovery run no longer RUNNING at commit; discarding result",
					log.String("run_id", run.ID.String()),
				)
			}

			return nil
		},
	)
}

func sanitizeError(err error) string {
	msg := err.Error()
	if len(msg) <= errorMessageMaxLen {
		return msg
	}

	cut := errorMessageMaxLen
	for cut > 0 && !utf8.RuneStart(msg[cut]) {
		cut--
	}

	return msg[:cut] + "…"
}
