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

package agentrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

type (
	handler struct {
		pg                *pg.Client
		registry          agent.AgentRegistry
		preparer          ExecutionPreparer
		logger            *log.Logger
		heartbeatInterval time.Duration
		staleAfter        time.Duration
		retryBase         time.Duration
		retryMax          time.Duration
		shutdownCh        chan struct{}
		shutdownOnce      sync.Once
		now               func() time.Time
	}

	leaseCheckpointer struct {
		pg         *pg.Client
		execution  *coredata.AgentExecution
		ownerToken string
	}
)

const (
	errorMessageMaxLen           = 512
	maxCheckpointBytes           = 10 * 1024 * 1024
	leaseLostErrorString         = "agent execution lease lost"
	conversationalUserInputBatch = 1
)

var (
	_ worker.Handler[coredata.AgentExecution] = (*handler)(nil)
	_ worker.StaleRecoverer                   = (*handler)(nil)
	_ agent.Checkpointer                      = (*leaseCheckpointer)(nil)

	errAgentExecutionLeaseLost = errors.New(leaseLostErrorString)
	errAgentExecutionFailed    = errors.New("agent execution failed")
)

func (h *handler) Claim(ctx context.Context) (coredata.AgentExecution, error) {
	ownerToken, err := uuid.NewV7()
	if err != nil {
		return coredata.AgentExecution{}, fmt.Errorf("cannot generate agent execution owner token: %w", err)
	}

	var execution coredata.AgentExecution

	err = h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return execution.ClaimNextForUpdateSkipLocked(
				ctx,
				tx,
				h.now(),
				ownerToken.String(),
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.AgentExecution{}, worker.ErrNoTask
		}

		return coredata.AgentExecution{}, fmt.Errorf("cannot claim agent execution: %w", err)
	}

	return execution, nil
}

func (h *handler) Process(ctx context.Context, execution coredata.AgentExecution) error {
	if execution.ProcessingOwnerToken == nil || *execution.ProcessingOwnerToken == "" {
		return fmt.Errorf("claimed agent execution has no owner token")
	}

	ownerToken := *execution.ProcessingOwnerToken

	runCtx, cancelRun := context.WithCancelCause(ctx)
	defer cancelRun(nil)

	heartbeatStop := make(chan struct{})

	heartbeatExited := make(chan struct{})
	lease := coredata.AgentExecution{ID: execution.ID}

	go h.maintainLease(
		runCtx,
		cancelRun,
		lease,
		ownerToken,
		heartbeatStop,
		heartbeatExited,
	)

	var stopHeartbeatOnce sync.Once

	stopHeartbeat := func() {
		stopHeartbeatOnce.Do(
			func() {
				close(heartbeatStop)
				<-heartbeatExited
			},
		)
	}

	preparedCtx := runCtx
	registry := h.registry
	var (
		err    error
		runErr error
	)

	switch execution.ExecutionKind {
	case coredata.AgentExecutionKindOneShot:
		preparedCtx, registry, err = h.preparer.Prepare(runCtx, &execution, h.registry, nil)
		if err != nil {
			stopHeartbeat()

			return h.handleFailure(ctx, &execution, nil, ownerToken, fmt.Errorf("cannot prepare agent execution: %w", err))
		}

		runErr = h.processOneShot(
			preparedCtx,
			ctx,
			registry,
			&execution,
			ownerToken,
			stopHeartbeat,
		)
	case coredata.AgentExecutionKindConversational:
		runErr = h.processConversation(
			runCtx,
			ctx,
			&execution,
			ownerToken,
			stopHeartbeat,
		)
	default:
		stopHeartbeat()

		return h.handleFailure(
			ctx,
			&execution,
			nil,
			ownerToken,
			fmt.Errorf("unsupported agent execution kind %q", execution.ExecutionKind),
		)
	}

	stopHeartbeat()

	if errors.Is(context.Cause(runCtx), errAgentExecutionLeaseLost) {
		return errAgentExecutionLeaseLost
	}

	return runErr
}

func (h *handler) RecoverStale(ctx context.Context) error {
	now := h.now()

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeadLetterAgentInputsForStaleExecutions(
				ctx,
				tx,
				now,
				h.staleAfter,
				coredata.AgentExecutionStaleLeaseError,
			); err != nil {
				return fmt.Errorf("cannot recover stale agent execution inputs: %w", err)
			}

			return coredata.ResetStaleAgentExecutionLeases(ctx, tx, now, h.staleAfter)
		},
	)
}

func (h *handler) maintainLease(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	lease coredata.AgentExecution,
	ownerToken string,
	stop <-chan struct{},
	exited chan<- struct{},
) {
	defer close(exited)

	ticker := time.NewTicker(h.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-h.shutdownCh:
			cancel(agent.ErrSuspendForCheckpoint)

			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := h.now()

			err := h.pg.WithConn(
				context.WithoutCancel(ctx),
				func(ctx context.Context, conn pg.Querier) error {
					return lease.Heartbeat(
						ctx,
						conn,
						coredata.NewScopeFromObjectID(lease.ID),
						ownerToken,
						now,
					)
				},
			)
			if err != nil {
				select {
				case <-stop:
					return
				default:
				}

				if errors.Is(err, coredata.ErrResourceNotFound) {
					cancel(errAgentExecutionLeaseLost)
				} else {
					cancel(fmt.Errorf("cannot heartbeat agent execution: %w", err))
				}

				return
			}
		}
	}
}

func (h *handler) processOneShot(
	runCtx context.Context,
	commitCtx context.Context,
	registry agent.AgentRegistry,
	execution *coredata.AgentExecution,
	ownerToken string,
	stopHeartbeat func(),
) error {
	checkpointer := &leaseCheckpointer{
		pg:         h.pg,
		execution:  execution,
		ownerToken: ownerToken,
	}

	result, runErr := h.runAgent(runCtx, registry, execution, checkpointer, execution.InputMessages)

	stopHeartbeat()

	if cause := context.Cause(runCtx); cause != nil && !errors.Is(cause, agent.ErrSuspendForCheckpoint) {
		return cause
	}

	execution.Result = nil
	execution.ErrorMessage = nil

	switch {
	case runErr == nil:
		execution.Status = coredata.AgentRunStatusCompleted

		if result != nil {
			data, err := json.Marshal(result)
			if err != nil {
				runErr = fmt.Errorf("cannot marshal agent run result: %w", err)
			} else {
				execution.Result = data
			}
		}
	case isSuspended(runErr):
		execution.Status = coredata.AgentRunStatusPending
		runErr = nil
	case isInterrupted(runErr):
		execution.Status = coredata.AgentRunStatusAwaitingApproval
		runErr = nil
	}

	if runErr != nil {
		return h.handleFailure(commitCtx, execution, nil, ownerToken, runErr)
	}

	now := h.now()

	err := h.pg.WithConn(
		context.WithoutCancel(commitCtx),
		func(ctx context.Context, conn pg.Querier) error {
			return execution.CommitOneShotResult(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(execution.ID),
				ownerToken,
				now,
			)
		},
	)
	if err != nil {
		return h.mapLeaseWriteError("cannot commit one-shot agent execution", err)
	}

	return nil
}

func (h *handler) processConversation(
	runCtx context.Context,
	commitCtx context.Context,
	execution *coredata.AgentExecution,
	ownerToken string,
	stopHeartbeat func(),
) error {
	inputs, err := h.loadConversationInputs(commitCtx, execution, ownerToken)
	if err != nil {
		stopHeartbeat()

		return h.handleFailure(commitCtx, execution, nil, ownerToken, err)
	}

	if len(inputs) == 0 {
		stopHeartbeat()

		return h.handleFailure(
			commitCtx,
			execution,
			nil,
			ownerToken,
			errors.New("cannot process conversational execution without pending inputs"),
		)
	}

	if inputs[0].Purpose != coredata.AgentInputPurposeUser {
		stopHeartbeat()

		return h.handleFailure(
			commitCtx,
			execution,
			inputs,
			ownerToken,
			fmt.Errorf("conversational execution received non-user input %q", inputs[0].Purpose),
		)
	}

	preparedCtx, registry, err := h.preparer.Prepare(runCtx, execution, h.registry, inputs[0])
	if err != nil {
		stopHeartbeat()

		return h.handleFailure(
			commitCtx,
			execution,
			inputs,
			ownerToken,
			fmt.Errorf("cannot prepare agent execution: %w", err),
		)
	}

	messages, err := decodeConversationMessages(execution.SessionMessages, inputs)
	if err != nil {
		stopHeartbeat()

		return h.handleFailure(commitCtx, execution, inputs, ownerToken, err)
	}

	checkpointer := &leaseCheckpointer{
		pg:         h.pg,
		execution:  execution,
		ownerToken: ownerToken,
	}

	result, runErr := h.runAgent(preparedCtx, registry, execution, checkpointer, messages)

	stopHeartbeat()

	if cause := context.Cause(runCtx); cause != nil &&
		!errors.Is(cause, agent.ErrSuspendForCheckpoint) {
		return cause
	}

	if isSuspended(runErr) || isInterrupted(runErr) {
		status := coredata.AgentRunStatusPending
		if isInterrupted(runErr) {
			status = coredata.AgentRunStatusAwaitingApproval
		}

		return h.releaseResting(commitCtx, execution, ownerToken, status)
	}

	if runErr != nil {
		return h.handleFailure(commitCtx, execution, inputs, ownerToken, runErr)
	}

	if result == nil {
		return h.handleFailure(
			commitCtx,
			execution,
			inputs,
			ownerToken,
			errors.New("agent execution returned no result"),
		)
	}

	encodedSession, err := json.Marshal(result.Messages)
	if err != nil {
		return h.handleFailure(
			commitCtx,
			execution,
			inputs,
			ownerToken,
			fmt.Errorf("cannot marshal conversational session messages: %w", err),
		)
	}

	execution.SessionMessages = encodedSession

	now := h.now()

	err = h.pg.WithTx(
		context.WithoutCancel(commitCtx),
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(execution.ID)
			for _, input := range inputs {
				if err := input.MarkProcessed(ctx, tx, scope, ownerToken, now); err != nil {
					return fmt.Errorf("cannot mark conversational input processed: %w", err)
				}
			}

			return execution.CommitConversationalSuccess(ctx, tx, scope, ownerToken, now)
		},
	)
	if err != nil {
		return h.mapLeaseWriteError("cannot commit conversational agent execution", err)
	}

	return nil
}

func (h *handler) runAgent(
	ctx context.Context,
	registry agent.AgentRegistry,
	execution *coredata.AgentExecution,
	checkpointer agent.Checkpointer,
	inputMessages json.RawMessage,
) (*agent.Result, error) {
	runID := execution.ID.String()
	if execution.Checkpoint != nil {
		h.logger.InfoCtx(ctx, "resuming agent execution", log.String("run_id", runID))

		return agent.Restore(ctx, checkpointer, runID, registry)
	}

	h.logger.InfoCtx(ctx, "starting agent execution", log.String("run_id", runID))

	a, err := registry.Agent(execution.StartAgentName)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve agent %q: %w", execution.StartAgentName, err)
	}

	var messages []llm.Message
	if err := json.Unmarshal(inputMessages, &messages); err != nil {
		return nil, fmt.Errorf("cannot unmarshal input messages: %w", err)
	}

	return a.Run(ctx, messages, agent.WithCheckpointer(checkpointer, runID))
}

func (h *handler) loadConversationInputs(
	ctx context.Context,
	execution *coredata.AgentExecution,
	ownerToken string,
) (coredata.AgentInputs, error) {
	var inputs coredata.AgentInputs

	err := h.pg.WithTx(
		context.WithoutCancel(ctx),
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(execution.ID)
			if len(execution.ProcessingInputIDs) > 0 {
				if err := inputs.LoadPendingByIDs(
					ctx,
					tx,
					scope,
					execution.ID,
					ownerToken,
					execution.ProcessingInputIDs,
					conversationalUserInputBatch,
				); err != nil {
					return fmt.Errorf("cannot load claimed conversational inputs: %w", err)
				}

				if len(inputs) > 0 {
					return nil
				}

				if err := execution.ClearProcessingInputIDs(
					ctx,
					tx,
					scope,
					ownerToken,
					h.now(),
				); err != nil {
					return fmt.Errorf("cannot clear stale conversational input ids: %w", err)
				}
			}

			if err := inputs.LoadPendingByAgentRunID(
				ctx,
				tx,
				scope,
				execution.ID,
				ownerToken,
				h.now(),
				conversationalUserInputBatch,
			); err != nil {
				return fmt.Errorf("cannot load pending conversational inputs: %w", err)
			}

			if len(inputs) == 0 {
				return errors.New("cannot process conversational execution without pending inputs")
			}

			inputIDs := make([]string, len(inputs))
			for idx, input := range inputs {
				inputIDs[idx] = input.ID.String()
			}

			return execution.SetProcessingInputIDs(
				ctx,
				tx,
				scope,
				ownerToken,
				inputIDs,
				h.now(),
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare conversational input batch: %w", err)
	}

	return inputs, nil
}

func (h *handler) handleFailure(
	ctx context.Context,
	execution *coredata.AgentExecution,
	inputs coredata.AgentInputs,
	ownerToken string,
	runErr error,
) error {
	h.logger.ErrorCtx(
		context.WithoutCancel(ctx),
		"agent execution failed",
		log.String("run_id", execution.ID.String()),
		log.String("error_type", fmt.Sprintf("%T", runErr)),
	)

	message := sanitizeError(runErr)
	now := h.now()
	scope := coredata.NewScopeFromObjectID(execution.ID)

	terminal := execution.AttemptCount >= execution.MaxAttempts
	for _, input := range inputs {
		if input.AttemptCount+1 >= input.MaxAttempts {
			terminal = true
			break
		}
	}

	var persistErr error
	if terminal {
		persistErr = h.pg.WithTx(
			context.WithoutCancel(ctx),
			func(ctx context.Context, tx pg.Tx) error {
				for _, input := range inputs {
					if err := input.DeadLetter(ctx, tx, scope, ownerToken, message, now); err != nil {
						return fmt.Errorf("cannot dead-letter conversational input: %w", err)
					}
				}

				return execution.DeadLetter(ctx, tx, scope, ownerToken, message, now)
			},
		)
	} else {
		nextAttemptAt := now.Add(h.retryDelay(execution.AttemptCount))
		persistErr = h.pg.WithTx(
			context.WithoutCancel(ctx),
			func(ctx context.Context, tx pg.Tx) error {
				for _, input := range inputs {
					if err := input.RecordFailure(
						ctx,
						tx,
						scope,
						ownerToken,
						message,
						nextAttemptAt,
						now,
					); err != nil {
						return fmt.Errorf("cannot record conversational input failure: %w", err)
					}
				}

				return execution.RecordFailure(
					ctx,
					tx,
					scope,
					ownerToken,
					message,
					nextAttemptAt,
					now,
				)
			},
		)
	}

	if persistErr != nil {
		return h.mapLeaseWriteError(
			fmt.Sprintf("cannot persist agent execution failure after %v", runErr),
			persistErr,
		)
	}

	return errAgentExecutionFailed
}

func (h *handler) releaseResting(
	ctx context.Context,
	execution *coredata.AgentExecution,
	ownerToken string,
	status coredata.AgentRunStatus,
) error {
	now := h.now()

	err := h.pg.WithConn(
		context.WithoutCancel(ctx),
		func(ctx context.Context, conn pg.Querier) error {
			return execution.ReleaseToStatus(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(execution.ID),
				ownerToken,
				status,
				now,
			)
		},
	)
	if err != nil {
		return h.mapLeaseWriteError("cannot release resting agent execution", err)
	}

	return nil
}

func (h *handler) retryDelay(attempt int) time.Duration {
	delay := h.retryBase
	for idx := 1; idx < attempt && delay < h.retryMax; idx++ {
		if delay > h.retryMax/2 {
			return h.retryMax
		}

		delay *= 2
	}

	if delay > h.retryMax {
		return h.retryMax
	}

	return delay
}

func (h *handler) mapLeaseWriteError(action string, err error) error {
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("%s: %w", action, errAgentExecutionLeaseLost)
	}

	return fmt.Errorf("%s: %w", action, err)
}

func (h *handler) signalShutdown() {
	h.shutdownOnce.Do(func() { close(h.shutdownCh) })
}

func (s *leaseCheckpointer) Save(ctx context.Context, runID string, cp *agent.Checkpoint) error {
	if runID != s.execution.ID.String() {
		return fmt.Errorf("cannot save agent execution checkpoint: run ID mismatch")
	}

	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("cannot marshal agent execution checkpoint: %w", err)
	}

	if len(data) > maxCheckpointBytes {
		return fmt.Errorf(
			"cannot marshal agent execution checkpoint: size %d exceeds limit %d",
			len(data),
			maxCheckpointBytes,
		)
	}

	err = s.pg.WithConn(
		context.WithoutCancel(ctx),
		func(ctx context.Context, conn pg.Querier) error {
			return s.execution.SaveCheckpoint(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(s.execution.ID),
				s.ownerToken,
				data,
				time.Now(),
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return errAgentExecutionLeaseLost
		}

		return fmt.Errorf("cannot persist agent execution checkpoint: %w", err)
	}

	return nil
}

func (s *leaseCheckpointer) Load(ctx context.Context, runID string) (*agent.Checkpoint, error) {
	if runID != s.execution.ID.String() {
		return nil, fmt.Errorf("cannot load agent execution checkpoint: run ID mismatch")
	}

	if len(s.execution.Checkpoint) == 0 {
		return nil, nil
	}

	if len(s.execution.Checkpoint) > maxCheckpointBytes {
		return nil, fmt.Errorf(
			"cannot load agent execution checkpoint: size %d exceeds limit %d",
			len(s.execution.Checkpoint),
			maxCheckpointBytes,
		)
	}

	err := s.pg.WithConn(
		context.WithoutCancel(ctx),
		func(ctx context.Context, conn pg.Querier) error {
			lease := *s.execution

			return lease.Heartbeat(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(lease.ID),
				s.ownerToken,
				time.Now(),
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, errAgentExecutionLeaseLost
		}

		return nil, fmt.Errorf("cannot verify agent execution lease: %w", err)
	}

	cp := new(agent.Checkpoint)
	if err := json.Unmarshal(s.execution.Checkpoint, cp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal agent execution checkpoint: %w", err)
	}

	return cp, nil
}

func decodeConversationMessages(
	sessionMessages json.RawMessage,
	inputs coredata.AgentInputs,
) (json.RawMessage, error) {
	var messages []llm.Message
	if len(sessionMessages) > 0 {
		if err := json.Unmarshal(sessionMessages, &messages); err != nil {
			return nil, fmt.Errorf("cannot unmarshal conversational session messages: %w", err)
		}
	}

	for _, input := range inputs {
		if input.Purpose != coredata.AgentInputPurposeUser {
			continue
		}

		var message llm.Message
		if err := json.Unmarshal(input.Message, &message); err != nil {
			return nil, fmt.Errorf("cannot unmarshal conversational input %s: %w", input.ID, err)
		}

		messages = append(messages, message)
	}

	data, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal conversational input messages: %w", err)
	}

	return data, nil
}

func isSuspended(err error) bool {
	_, ok := errors.AsType[*agent.SuspendedError](err)

	return ok
}

func isInterrupted(err error) bool {
	_, ok := errors.AsType[*agent.InterruptedError](err)

	return ok
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
