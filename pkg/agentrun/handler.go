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

package agentrun

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

type handler struct {
	pg            *pg.Client
	store         *coredata.PGCheckpointer
	registry      agent.AgentRegistry
	logger        *log.Logger
	leaseDuration time.Duration
	shutdownCh    chan struct{}
	shutdownOnce  sync.Once
}

var (
	_ worker.Handler[coredata.AgentRun] = (*handler)(nil)
	_ worker.StaleRecoverer             = (*handler)(nil)
)

// Claim loads the next pending agent run, marks it RUNNING with a lease
// owned by this worker, and returns the row. When no work is available it
// returns worker.ErrNoTask so the kit can back off until the next tick.
func (h *handler) Claim(ctx context.Context) (coredata.AgentRun, error) {
	var (
		run            = coredata.AgentRun{}
		now            = time.Now()
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
			run.LeaseExpiresAt = &leaseExpiresAt
			run.LeaseGeneration++
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

// Process executes a single agent run. It spawns a heartbeat goroutine
// that renews the lease while the run is active, and a forwarder
// goroutine that converts the handler-level shutdown broadcast into a
// per-run ctx cancellation so the agent loop checkpoints cleanly at
// its next turn boundary.
//
// The returned error mirrors the run outcome so the worker kit's
// task metrics and OTel span status reflect actual agent failures.
// nil is returned for both successful runs and graceful exits
// (lease loss, infrastructure suspension) where the row state is
// already consistent.
func (h *handler) Process(ctx context.Context, run coredata.AgentRun) error {
	runCtx, cancelRun := context.WithCancelCause(ctx)
	defer cancelRun(nil)
	leaseGeneration := run.LeaseGeneration

	forwarderDone := make(chan struct{})
	defer close(forwarderDone)

	go func() {
		select {
		case <-h.shutdownCh:
			cancelRun(agent.ErrSuspendForCheckpoint)
		case <-forwarderDone:
		}
	}()

	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()

	go h.heartbeatLease(heartbeatCtx, run.ID.String(), leaseGeneration, cancelRun)

	return h.executeRun(runCtx, &run, leaseGeneration)
}

// RecoverStale resets agent runs whose worker lease has expired back to
// PENDING so a fresh worker can pick them up on the next cycle.
func (h *handler) RecoverStale(ctx context.Context) error {
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
func (h *handler) signalShutdown() {
	h.shutdownOnce.Do(func() { close(h.shutdownCh) })
}

func (h *handler) heartbeatLease(
	ctx context.Context,
	runID string,
	leaseGeneration int64,
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
					rowsAffected, err := coredata.HeartbeatAgentRunLease(
						ctx,
						conn,
						runID,
						leaseGeneration,
						expiresAt,
					)
					if err != nil {
						return err
					}

					if rowsAffected == 0 {
						return ErrLeaseLost
					}

					return nil
				},
			); err != nil {
				h.logger.ErrorCtx(ctx, "cannot heartbeat agent run lease", log.Error(err))

				if errors.Is(err, ErrLeaseLost) {
					cancelRun(ErrLeaseLost)
				} else {
					cancelRun(fmt.Errorf("%w: %w", ErrHeartbeatFailed, err))
				}

				return
			}
		}
	}
}

const (
	// errorMessageMaxLen caps the error string persisted to the
	// agent_runs.error_message column. Raw tool or LLM errors can embed
	// URLs with credentials, response snippets containing PII, or partial
	// records from failed DB lookups; the full context is logged while
	// only a truncated summary is stored for caller-visible state.
	errorMessageMaxLen = 512
)

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

type leasedCheckpointer struct {
	store           *coredata.PGCheckpointer
	leaseGeneration int64
}

func (s leasedCheckpointer) Save(ctx context.Context, runID string, cp *agent.Checkpoint) error {
	return s.store.SaveForLease(ctx, runID, cp, s.leaseGeneration)
}

func (s leasedCheckpointer) Load(ctx context.Context, runID string) (*agent.Checkpoint, error) {
	return s.store.Load(ctx, runID)
}

func (h *handler) executeRun(
	ctx context.Context,
	run *coredata.AgentRun,
	leaseGeneration int64,
) error {
	runID := run.ID.String()
	checkpointer := leasedCheckpointer{
		store:           h.store,
		leaseGeneration: leaseGeneration,
	}

	var (
		result *agent.Result
		runErr error
	)

	if run.Checkpoint != nil {
		h.logger.InfoCtx(ctx, "resuming agent run", log.String("run_id", runID))
		result, runErr = agent.Restore(ctx, checkpointer, runID, h.registry)
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
				result, runErr = a.Run(
					ctx,
					inputMsgs,
					agent.WithCheckpointer(checkpointer, runID),
				)
			}
		}
	}

	// Heartbeat loss: another worker may have taken over. Do not commit
	// any status — stale recovery will handle the row. Surface the cause
	// so the worker kit logs and traces a failure for this attempt.
	if cause := context.Cause(ctx); errors.Is(cause, ErrLeaseLost) || errors.Is(cause, ErrHeartbeatFailed) {
		h.logger.WarnCtx(
			context.WithoutCancel(ctx),
			"agent run stopped after heartbeat failure; leaving status for stale recovery",
			log.String("run_id", runID),
			log.Error(cause),
		)

		return cause
	}

	// Infrastructure-triggered suspension (graceful shutdown): leave the
	// row as RUNNING so stale recovery resets it to PENDING on restart.
	// The checkpoint was already saved by coreLoop before returning
	// SuspendedError, so Restore will pick up where it left off. This
	// is not a failure from the worker kit's perspective.
	if runErr != nil {
		if _, ok := errors.AsType[*agent.SuspendedError](runErr); ok {
			h.logger.InfoCtx(
				context.WithoutCancel(ctx),
				"agent run suspended by infrastructure; leaving for stale recovery",
				log.String("run_id", runID),
			)

			return nil
		}
	}

	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.LeaseExpiresAt = nil

	if runErr == nil {
		run.Status = coredata.AgentRunStatusCompleted

		if result != nil {
			data, err := json.Marshal(result)
			if err != nil {
				h.logger.ErrorCtx(ctx, "cannot marshal agent run result", log.Error(err))
				runErr = fmt.Errorf("cannot marshal agent run result: %w", err)
			} else {
				run.Result = data
			}
		}
	}

	if runErr != nil {
		run.Status = coredata.AgentRunStatusFailed
		run.Result = nil

		h.logger.ErrorCtx(
			context.WithoutCancel(ctx),
			"agent run failed",
			log.String("run_id", runID),
			log.Error(runErr),
		)
		msg := sanitizeError(runErr)
		run.ErrorMessage = &msg
	}

	commitCtx := context.WithoutCancel(ctx)

	if err := h.pg.WithTx(
		commitCtx,
		func(ctx context.Context, tx pg.Tx) error {
			rowsAffected, err := coredata.CommitAgentRunResult(ctx, tx, run, leaseGeneration)
			if err != nil {
				return err
			}

			if rowsAffected == 0 {
				return ErrLeaseLost
			}

			if run.Status == coredata.AgentRunStatusCompleted {
				if err := run.ClearCheckpoint(ctx, tx, coredata.NewNoScope()); err != nil {
					return err
				}
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, ErrLeaseLost) {
			h.logger.WarnCtx(
				commitCtx,
				"agent run lost lease before commit; discarding stale completion",
				log.String("run_id", runID),
			)

			return nil
		}

		h.logger.ErrorCtx(commitCtx, "cannot commit agent run status", log.Error(err))
		return fmt.Errorf("cannot commit agent run status: %w", err)
	}

	return runErr
}
