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

package agentrun

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type Service struct {
	pg *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

func (s *Service) Get(
	ctx context.Context,
	predicate coredata.Predicater,
	agentRunID gid.GID,
) (*coredata.AgentRun, error) {
	run := &coredata.AgentRun{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := run.LoadByID(ctx, conn, predicate, agentRunID); err != nil {
				return fmt.Errorf("cannot load agent run: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return run, nil
}

func (s *Service) ListForOrganizationID(
	ctx context.Context,
	predicate coredata.Predicater,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AgentRunOrderField],
) (*page.Page[*coredata.AgentRun, coredata.AgentRunOrderField], error) {
	var runs coredata.AgentRuns

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, predicate, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := runs.LoadByOrganizationID(ctx, conn, predicate, organization.ID, cursor); err != nil {
				return fmt.Errorf("cannot load agent runs: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(runs, cursor), nil
}

// SubmitApproval records human approval decisions for a run parked in
// AWAITING_APPROVAL and requeues it to PENDING so a worker resumes it.
// decisions is keyed by pending tool-call ID and must cover exactly the
// run's pending approvals (a missing decision would be treated as an
// implicit denial on resume, so partial submissions are rejected). The
// refreshed run is returned.
func (s *Service) SubmitApproval(
	ctx context.Context,
	predicate coredata.Predicater,
	agentRunID gid.GID,
	decisions map[string]agent.ApprovalResult,
) (*coredata.AgentRun, error) {
	run := &coredata.AgentRun{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadByIDForUpdate(ctx, tx, predicate, agentRunID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrAgentRunNotFound
				}

				return fmt.Errorf("cannot load agent run: %w", err)
			}

			if run.Status != coredata.AgentRunStatusAwaitingApproval {
				return ErrNotAwaitingApproval
			}

			if run.Checkpoint == nil {
				return fmt.Errorf("agent run %s has no checkpoint", agentRunID)
			}

			checkpoint, err := agent.MergeApprovalDecisions(run.Checkpoint, decisions)
			if err != nil {
				if errors.Is(err, agent.ErrApprovalDecisionsMismatch) {
					return ErrApprovalDecisionsMismatch
				}

				return fmt.Errorf("cannot merge approval decisions: %w", err)
			}

			run.Checkpoint = checkpoint
			run.Status = coredata.AgentRunStatusPending
			run.StartedAt = nil
			run.UpdatedAt = time.Now()

			if err := run.RequeueForApprovalResume(ctx, tx, predicate); err != nil {
				return fmt.Errorf("cannot requeue agent run for approval resume: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return run, nil
}

func (s *Service) CountForOrganizationID(
	ctx context.Context,
	predicate coredata.Predicater,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			runs := &coredata.AgentRuns{}

			count, err = runs.CountByOrganizationID(ctx, conn, predicate, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count agent runs: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
