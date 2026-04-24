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
	"sync"
	"time"
	"unicode/utf8"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

type agentRunHandler struct {
	pg            *pg.Client
	store         *coredata.PGCheckpointer
	registry      agent.AgentRegistry
	logger        *log.Logger
	leaseDuration time.Duration
	workerID      string
	shutdownCh    chan struct{}
	shutdownOnce  sync.Once
}

var (
	_ worker.Handler[coredata.AgentRun] = (*agentRunHandler)(nil)
	_ worker.StaleRecoverer             = (*agentRunHandler)(nil)
)

// Claim loads the next pending agent run, marks it RUNNING with a lease
// owned by this worker, and returns the row. When no work is available it
// returns worker.ErrNoTask so the kit can back off until the next tick.
func (h *agentRunHandler) Claim(ctx context.Context) (coredata.AgentRun, error) {
	var (
		run            = coredata.AgentRun{}
		now            = time.Now()
		leaseOwner     = h.workerID
		leaseExpiresAt = now.Add(h.leaseDuration)
	)

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot load next pending agent run: %w", err)
			}

			run.Status = coredata.AgentRunStatusRunning
			run.StartedAt = &now
			run.LeaseOwner = &leaseOwner
			run.LeaseExpiresAt = &leaseExpiresAt
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

// Process executes a single agent run. It spawns a heartbeat goroutine to
// renew the lease while the run is active, and a forwarder goroutine that
// bridges the handler-level shutdown signal to the run's agent stop
// channel so the agent checkpoints cleanly at the next turn boundary.
func (h *agentRunHandler) Process(ctx context.Context, run coredata.AgentRun) error {
	runCtx, cancelRun := context.WithCancelCause(ctx)
	defer cancelRun(nil)

	stopCh := make(chan struct{})
	forwarderDone := make(chan struct{})
	defer close(forwarderDone)
	go func() {
		select {
		case <-h.shutdownCh:
			close(stopCh)
		case <-forwarderDone:
		}
	}()

	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()
	go h.heartbeatLease(heartbeatCtx, run.ID.String(), cancelRun)

	runCtx = agent.WithStopSignal(runCtx, stopCh)
	h.executeRun(runCtx, &run)

	return nil
}

// RecoverStale resets agent runs whose worker lease has expired back to
// PENDING so a fresh supervisor can pick them up on the next cycle.
func (h *agentRunHandler) RecoverStale(ctx context.Context) error {
	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleAgentRuns(ctx, conn)
		},
	); err != nil {
		return fmt.Errorf("cannot reset stale agent runs: %w", err)
	}
	return nil
}

// signalShutdown closes the handler-level shutdown broadcast channel. All
// in-flight Process forwarder goroutines observe the close and propagate
// it to their per-run agent stop channels, letting agents checkpoint at
// the next turn boundary before Process returns.
func (h *agentRunHandler) signalShutdown() {
	h.shutdownOnce.Do(func() { close(h.shutdownCh) })
}

func (h *agentRunHandler) heartbeatLease(
	ctx context.Context,
	runID string,
	cancelRun context.CancelCauseFunc,
) {
	ticker := time.NewTicker(h.leaseDuration / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expiresAt := time.Now().Add(h.leaseDuration)

			if err := h.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					rowsAffected, err := coredata.HeartbeatAgentRunLease(ctx, conn, runID, h.workerID, expiresAt)
					if err != nil {
						return err
					}

					if rowsAffected == 0 {
						return ErrAgentRunLeaseLost
					}

					return nil
				},
			); err != nil {
				h.logger.ErrorCtx(ctx, "cannot heartbeat agent run lease", log.Error(err))
				if errors.Is(err, ErrAgentRunLeaseLost) {
					cancelRun(ErrAgentRunLeaseLost)
				} else {
					cancelRun(fmt.Errorf("%w: %w", ErrAgentRunHeartbeatFailed, err))
				}
				return
			}
		}
	}
}

const (
	// agentRunErrorMessageMaxLen caps the error string persisted to the
	// agent_runs.error_message column. Raw tool or LLM errors can embed
	// URLs with credentials, response snippets containing PII, or partial
	// records from failed DB lookups; the full context is logged while
	// only a truncated summary is stored for caller-visible state.
	agentRunErrorMessageMaxLen = 512
)

func sanitizeAgentRunError(err error) string {
	msg := err.Error()
	if len(msg) <= agentRunErrorMessageMaxLen {
		return msg
	}

	cut := agentRunErrorMessageMaxLen
	for cut > 0 && !utf8.RuneStart(msg[cut]) {
		cut--
	}
	return msg[:cut] + "…"
}

func (h *agentRunHandler) executeRun(ctx context.Context, run *coredata.AgentRun) {
	runID := run.ID.String()

	var (
		result *agent.Result
		runErr error
	)

	if run.Checkpoint != nil {
		h.logger.InfoCtx(ctx, "resuming agent run", log.String("run_id", runID))
		result, runErr = agent.Restore(ctx, h.store, runID, h.registry)
	} else {
		h.logger.InfoCtx(ctx, "starting agent run", log.String("run_id", runID))

		a, err := h.registry.Agent(run.StartAgentName)
		if err != nil {
			runErr = fmt.Errorf("cannot resolve agent %q: %w", run.StartAgentName, err)
		} else {
			var inputMsgs []llm.Message
			if err := json.Unmarshal(run.InputMessages, &inputMsgs); err != nil {
				runErr = fmt.Errorf("cannot unmarshal input messages: %w", err)
			} else {
				result, runErr = a.RunWithOpts(
					ctx,
					inputMsgs,
					agent.WithCheckpointer(h.store, runID),
				)
			}
		}
	}

	// Heartbeat loss: another worker may have taken over. Do not commit
	// any status — stale recovery will handle the row.
	if cause := context.Cause(ctx); errors.Is(cause, ErrAgentRunLeaseLost) || errors.Is(cause, ErrAgentRunHeartbeatFailed) {
		h.logger.WarnCtx(
			context.WithoutCancel(ctx),
			"agent run stopped after heartbeat failure; leaving status for stale recovery",
			log.String("run_id", runID),
			log.Error(cause),
		)
		return
	}

	// Infrastructure-triggered suspension (graceful shutdown): leave the
	// row as RUNNING so stale recovery resets it to PENDING on restart.
	// The checkpoint was already saved by coreLoop before returning
	// SuspendedError, so Restore will pick up where it left off.
	if runErr != nil {
		if _, ok := errors.AsType[*agent.SuspendedError](runErr); ok {
			h.logger.InfoCtx(
				context.WithoutCancel(ctx),
				"agent run suspended by infrastructure; leaving for stale recovery",
				log.String("run_id", runID),
			)
			return
		}
	}

	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.LeaseOwner = nil
	run.LeaseExpiresAt = nil

	switch {
	case runErr == nil:
		run.Status = coredata.AgentRunStatusCompleted
		if result != nil {
			data, err := json.Marshal(result)
			if err != nil {
				h.logger.ErrorCtx(ctx, "cannot marshal agent run result", log.Error(err))
			} else {
				run.Result = data
			}
		}

	default:
		run.Status = coredata.AgentRunStatusFailed
		h.logger.ErrorCtx(
			context.WithoutCancel(ctx),
			"agent run failed",
			log.String("run_id", runID),
			log.Error(runErr),
		)
		msg := sanitizeAgentRunError(runErr)
		run.ErrorMessage = &msg
	}

	commitCtx := context.WithoutCancel(ctx)

	if err := h.pg.WithTx(
		commitCtx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return err
			}

			if run.Status == coredata.AgentRunStatusCompleted {
				if err := run.ClearCheckpoint(ctx, tx, coredata.NewNoScope()); err != nil {
					return err
				}
			}

			return nil
		},
	); err != nil {
		h.logger.ErrorCtx(commitCtx, "cannot commit agent run status", log.Error(err))
	}
}
