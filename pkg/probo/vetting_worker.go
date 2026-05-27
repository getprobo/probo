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

package probo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/vetting"
)

const VettingAgentName = "third_party_vetter"

type (
	vettingMetadata struct {
		ThirdPartyID   gid.GID `json:"third_party_id"`
		OrganizationID gid.GID `json:"organization_id"`
		WebsiteURL     string  `json:"website_url"`
	}

	vettingHandler struct {
		pg         *pg.Client
		vetter     ThirdPartyVetter
		logger     *log.Logger
		staleAfter time.Duration
	}

	VettingWorkerConfig struct {
		StaleAfter time.Duration
	}
)

func NewVettingWorker(
	pgClient *pg.Client,
	vetter ThirdPartyVetter,
	logger *log.Logger,
	cfg VettingWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.AgentRun] {
	staleAfter := cfg.StaleAfter
	if staleAfter == 0 {
		staleAfter = 25 * time.Minute
	}

	h := &vettingHandler{
		pg:         pgClient,
		vetter:     vetter,
		logger:     logger,
		staleAfter: staleAfter,
	}

	return worker.New(
		"vetting-worker",
		h,
		logger,
		opts...,
	)
}

func (h *vettingHandler) Claim(ctx context.Context) (coredata.AgentRun, error) {
	var run coredata.AgentRun

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadNextPendingByAgentForUpdateSkipLocked(ctx, tx, VettingAgentName); err != nil {
				return err
			}

			now := time.Now()
			run.Status = coredata.AgentRunStatusRunning
			run.StartedAt = &now
			run.UpdatedAt = now

			if err := run.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update agent run: %w", err)
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

func (h *vettingHandler) Process(ctx context.Context, run coredata.AgentRun) error {
	if err := h.processRun(ctx, &run); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"vetting worker failure",
			log.Error(err),
			log.String("run_id", run.ID.String()),
		)

		if failErr := h.failRun(ctx, &run, err); failErr != nil {
			h.logger.ErrorCtx(ctx, "cannot mark agent run as failed", log.Error(failErr))
		}

		return err
	}

	return nil
}

func (h *vettingHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleAgentRuns(ctx, conn)
		},
	)
}

func (h *vettingHandler) processRun(ctx context.Context, run *coredata.AgentRun) error {
	var meta vettingMetadata
	if err := json.Unmarshal(run.Metadata, &meta); err != nil {
		return fmt.Errorf("cannot parse vetting metadata: %w", err)
	}

	var procedure string

	if run.InputMessages != nil {
		var msgs []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}

		if err := json.Unmarshal(run.InputMessages, &msgs); err == nil && len(msgs) > 0 && len(msgs[0].Parts) > 1 {
			procedure = msgs[0].Parts[1].Text
		}
	}

	pc := &vetting.PersistenceContext{
		PG:             h.pg,
		ThirdPartyID:   meta.ThirdPartyID,
		OrganizationID: meta.OrganizationID,
		WebsiteURL:     meta.WebsiteURL,
	}

	tools := []agent.Tool{
		vetting.SaveThirdPartyInfoTool(pc),
		vetting.LinkSubThirdPartyTool(pc),
	}

	if _, err := h.vetter.Assess(ctx, meta.WebsiteURL, procedure, nil, tools); err != nil {
		return fmt.Errorf("cannot vet third party: %w", err)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			run.Status = coredata.AgentRunStatusCompleted
			run.StartedAt = nil
			run.UpdatedAt = time.Now()

			return run.Update(ctx, tx, coredata.NewNoScope())
		},
	)
}

func (h *vettingHandler) failRun(ctx context.Context, run *coredata.AgentRun, reason error) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			errMsg := reason.Error()
			run.Status = coredata.AgentRunStatusFailed
			run.StartedAt = nil
			run.ErrorMessage = &errMsg
			run.UpdatedAt = time.Now()

			return run.Update(ctx, tx, coredata.NewNoScope())
		},
	)
}
