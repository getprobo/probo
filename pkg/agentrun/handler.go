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
	pg           *pg.Client
	store        *coredata.PGCheckpointer
	registry     *Registry
	logger       *log.Logger
	shutdownCh   chan struct{}
	shutdownOnce sync.Once
}

var _ worker.Handler[coredata.AgentRun] = (*handler)(nil)

// Claim loads the next pending agent run and marks it RUNNING. When no
// work is available it returns worker.ErrNoTask so the kit backs off
// until the next tick. The FOR UPDATE SKIP LOCKED select guarantees only
// one worker claims a given row; there is no lease, so a worker that
// crashes mid-run leaves the row RUNNING for manual recovery.
func (h *handler) Claim(ctx context.Context) (coredata.AgentRun, error) {
	var (
		run = coredata.AgentRun{}
		now = time.Now()
	)

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot load next pending agent run: %w", err)
			}

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

// Process executes a single agent run. It spawns a forwarder goroutine
// that converts the handler-level shutdown broadcast into a per-run ctx
// cancellation so the agent loop checkpoints cleanly at its next turn
// boundary.
//
// The returned error mirrors the run outcome so the worker kit's task
// metrics and OTel span status reflect actual agent failures. nil is
// returned for successful runs and for known stops (graceful suspend,
// awaiting approval) where the row was already committed to a resumable
// state.
func (h *handler) Process(ctx context.Context, run coredata.AgentRun) error {
	runCtx, cancelRun := context.WithCancelCause(ctx)
	defer cancelRun(nil)

	forwarderDone := make(chan struct{})
	defer close(forwarderDone)

	go func() {
		select {
		case <-h.shutdownCh:
			cancelRun(agent.ErrSuspendForCheckpoint)
		case <-forwarderDone:
		}
	}()

	return h.executeRun(runCtx, &run)
}

// signalShutdown closes the handler-level shutdown broadcast channel. All
// in-flight Process forwarder goroutines observe the close and propagate
// it to their per-run agent stop channels, letting agents checkpoint at
// the next turn boundary before Process returns.
func (h *handler) signalShutdown() {
	h.shutdownOnce.Do(func() { close(h.shutdownCh) })
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

func (h *handler) executeRun(ctx context.Context, run *coredata.AgentRun) error {
	if handler, ok := h.registry.RunHandler(run.StartAgentName); ok {
		return h.executeRunHandler(ctx, run, handler)
	}

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
				result, runErr = a.Run(
					ctx,
					inputMsgs,
					agent.WithCheckpointer(h.store, runID),
				)
			}
		}
	}

	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.Result = nil
	run.ErrorMessage = nil

	if runErr == nil {
		run.Status = coredata.AgentRunStatusCompleted

		if result != nil {
			data, err := json.Marshal(result)
			if err != nil {
				runErr = fmt.Errorf("cannot marshal agent run result: %w", err)
			} else {
				run.Result = data
			}
		}
	}

	// Known stops are not failures: the agent loop already saved a
	// checkpoint before returning. Graceful suspend returns the run to
	// PENDING so any worker resumes it from the checkpoint; an approval
	// interruption parks it in AWAITING_APPROVAL until an approval
	// decision requeues it. Anything else is a genuine failure.
	if runErr != nil {
		if _, ok := errors.AsType[*agent.SuspendedError](runErr); ok {
			run.Status = coredata.AgentRunStatusPending
			runErr = nil
		} else if _, ok := errors.AsType[*agent.InterruptedError](runErr); ok {
			run.Status = coredata.AgentRunStatusAwaitingApproval
			runErr = nil
		} else {
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
	}

	return h.commitRunResult(ctx, run, runErr)
}

func (h *handler) executeRunHandler(
	ctx context.Context,
	run *coredata.AgentRun,
	runHandler RunHandler,
) error {
	runID := run.ID.String()

	if run.Checkpoint != nil {
		return h.commitRunHandlerFailure(
			ctx,
			run,
			fmt.Errorf("run handler %q does not support checkpoints", run.StartAgentName),
		)
	}

	h.logger.InfoCtx(
		ctx,
		"starting run handler",
		log.String("run_id", runID),
		log.String("agent", run.StartAgentName),
	)

	resultJSON, runErr := runHandler.Run(ctx, run)

	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.Result = nil
	run.ErrorMessage = nil

	if runErr == nil {
		run.Status = coredata.AgentRunStatusCompleted
		run.Result = resultJSON
	} else {
		run.Status = coredata.AgentRunStatusFailed

		h.logger.ErrorCtx(
			context.WithoutCancel(ctx),
			"run handler failed",
			log.String("run_id", runID),
			log.String("agent", run.StartAgentName),
			log.Error(runErr),
		)

		msg := sanitizeError(runErr)
		run.ErrorMessage = &msg
	}

	return h.commitRunResult(ctx, run, runErr)
}

func (h *handler) commitRunHandlerFailure(
	ctx context.Context,
	run *coredata.AgentRun,
	runErr error,
) error {
	now := time.Now()
	run.UpdatedAt = now
	run.StartedAt = nil
	run.Result = nil
	run.Status = coredata.AgentRunStatusFailed

	msg := sanitizeError(runErr)
	run.ErrorMessage = &msg

	h.logger.ErrorCtx(
		context.WithoutCancel(ctx),
		"run handler failed",
		log.String("run_id", run.ID.String()),
		log.String("agent", run.StartAgentName),
		log.Error(runErr),
	)

	return h.commitRunResult(ctx, run, runErr)
}

func (h *handler) commitRunResult(
	ctx context.Context,
	run *coredata.AgentRun,
	runErr error,
) error {
	commitCtx := context.WithoutCancel(ctx)

	if err := h.pg.WithTx(
		commitCtx,
		func(ctx context.Context, tx pg.Tx) error {
			rowsAffected, err := coredata.CommitAgentRunResult(ctx, tx, run)
			if err != nil {
				return err
			}

			if rowsAffected == 0 {
				h.logger.WarnCtx(
					ctx,
					"agent run no longer RUNNING at commit; discarding result",
					log.String("run_id", run.ID.String()),
				)

				return nil
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

		return fmt.Errorf("cannot commit agent run status: %w", err)
	}

	return runErr
}
