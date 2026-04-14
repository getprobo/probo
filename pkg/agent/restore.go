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

package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/llm"
)

// Restore continues a previously suspended or approval-interrupted agent run
// from its last persisted checkpoint. The registry must contain all agents
// that may have been active (including handoff targets).
func Restore(
	ctx context.Context,
	store Checkpointer,
	runID string,
	registry AgentRegistry,
) (*Result, error) {
	cp, err := store.Load(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("cannot load checkpoint: %w", err)
	}
	if cp == nil {
		return nil, fmt.Errorf("cannot restore: no checkpoint for run %s", runID)
	}
	agent, err := registry.Agent(cp.AgentName)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve agent %q: %w", cp.AgentName, err)
	}

	return restoreCheckpoint(ctx, agent, cp, store, runID, registry)
}

func restoreCheckpoint(
	ctx context.Context,
	agent *Agent,
	cp *Checkpoint,
	store Checkpointer,
	runID string,
	registry AgentRegistry,
) (*Result, error) {
	emitHook(agent, func(h RunHooks) { h.OnRunRestore(ctx, agent, cp) })

	switch cp.Status {
	case AgentStatusSuspended:
		return restoreSuspended(ctx, agent, cp, store, runID, registry)

	case AgentStatusAwaitingApproval:
		return restoreAwaitingApproval(ctx, agent, cp, store, runID, registry)

	default:
		return nil, fmt.Errorf("cannot restore: unknown checkpoint status %q", cp.Status)
	}
}

func restoreSuspended(
	ctx context.Context,
	agent *Agent,
	cp *Checkpoint,
	store Checkpointer,
	runID string,
	registry AgentRegistry,
) (*Result, error) {
	if len(cp.InnerCheckpoints) > 0 {
		return restoreNestedSuspended(ctx, agent, cp, store, runID, registry)
	}

	return continueFromMessages(ctx, agent, cp.Messages, cp, store, runID)
}

func continueFromMessages(
	ctx context.Context,
	agent *Agent,
	messages []llm.Message,
	cp *Checkpoint,
	store Checkpointer,
	runID string,
) (*Result, error) {
	messagesCopy := make([]llm.Message, len(messages))
	copy(messagesCopy, messages)

	if cp.Turns >= agent.maxTurns {
		agent.logger.WarnCtx(
			ctx,
			"restored agent run has already reached max turns",
			log.String("agent", agent.name),
			log.Int("turns", cp.Turns),
			log.Int("max_turns", agent.maxTurns),
		)
	}

	return coreLoop(
		ctx,
		agent,
		messagesCopy,
		runOpts{
			callLLM:             blockingCallLLM,
			onEvent:             noopEvent,
			skipInputGuardrails: true,
			skipSessionLoad:     true,
			initialUsage:        cp.Usage,
			initialTurns:        cp.Turns,
			checkpointer:        store,
			runID:               runID,
			toolUsedInRun:       cp.ToolUsedInRun,
		},
	)
}

func restoreNestedSuspended(
	ctx context.Context,
	agent *Agent,
	cp *Checkpoint,
	store Checkpointer,
	runID string,
	registry AgentRegistry,
) (*Result, error) {
	type nestedRestoreEntry struct {
		toolCall            llm.ToolCall
		originalCheckpoint  *Checkpoint
		suspendedCheckpoint *Checkpoint
		result              ToolResult
		completed           bool
		err                 error
	}

	completedByID := make(map[string]ToolResult, len(cp.CompletedCalls))
	for _, cc := range cp.CompletedCalls {
		completedByID[cc.ToolCallID] = cc.Result
	}

	entries := make([]nestedRestoreEntry, len(cp.AllToolCalls))
	var wg sync.WaitGroup
	for i, tc := range cp.AllToolCalls {
		entries[i].toolCall = tc
		result, ok := completedByID[tc.ID]
		if ok {
			entries[i].result = result
			entries[i].completed = true
			continue
		}

		innerCP, ok := cp.InnerCheckpoints[tc.ID]
		if !ok {
			entries[i].err = fmt.Errorf("cannot restore nested tool call %q: missing inner checkpoint", tc.ID)
			continue
		}
		entries[i].originalCheckpoint = innerCP

		innerAgent, err := registry.Agent(innerCP.AgentName)
		if err != nil {
			entries[i].err = fmt.Errorf("cannot resolve inner agent %q: %w", innerCP.AgentName, err)
			continue
		}

		wg.Add(1)
		go func(i int, tc llm.ToolCall, innerAgent *Agent, innerCP *Checkpoint) {
			defer wg.Done()

			result, err := restoreCheckpoint(ctx, innerAgent, innerCP, nil, "", registry)
			if err != nil {
				if se, ok := errors.AsType[*SuspendedError](err); ok {
					if se.Checkpoint == nil {
						entries[i].err = fmt.Errorf("cannot restore nested tool call %q: missing suspension checkpoint", tc.ID)
						return
					}
					entries[i].suspendedCheckpoint = se.Checkpoint
					return
				}
				entries[i].err = fmt.Errorf("cannot restore nested tool call %q: %w", tc.ID, err)
				return
			}

			entries[i].result = ToolResult{Content: result.FinalMessage().Text()}
			entries[i].completed = true
		}(i, tc, innerAgent, innerCP)
	}
	wg.Wait()

	messages := make([]llm.Message, len(cp.Messages))
	copy(messages, cp.Messages)

	completedCalls := make([]CompletedCall, 0, len(cp.AllToolCalls))
	remainingInner := make(map[string]*Checkpoint)
	var restoreErr error
	for _, entry := range entries {
		switch {
		case entry.err != nil:
			if entry.originalCheckpoint != nil {
				remainingInner[entry.toolCall.ID] = entry.originalCheckpoint
			}
			if restoreErr == nil {
				restoreErr = entry.err
			}
			continue

		case entry.suspendedCheckpoint != nil:
			remainingInner[entry.toolCall.ID] = entry.suspendedCheckpoint
			continue

		case !entry.completed:
			if restoreErr == nil {
				restoreErr = fmt.Errorf("cannot restore nested tool call %q: no result", entry.toolCall.ID)
			}
			continue
		}

		completedCalls = append(
			completedCalls,
			CompletedCall{
				ToolCallID: entry.toolCall.ID,
				Result:     entry.result,
			},
		)
		messages = append(
			messages,
			llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: entry.toolCall.ID,
				Parts:      []llm.Part{llm.TextPart{Text: entry.result.Content}},
			},
		)
	}

	saveProgress := func() (*Checkpoint, error) {
		next := *cp
		next.InnerCheckpoints = remainingInner
		next.CompletedCalls = completedCalls
		if store != nil && runID != "" {
			if err := store.Save(ctx, runID, &next); err != nil {
				return nil, fmt.Errorf("cannot save nested restore progress: %w", err)
			}
		}
		return &next, nil
	}

	if restoreErr != nil {
		if _, err := saveProgress(); err != nil {
			return nil, errors.Join(restoreErr, err)
		}
		return nil, restoreErr
	}

	if len(remainingInner) > 0 {
		next, err := saveProgress()
		if err != nil {
			return nil, err
		}
		return nil, &SuspendedError{RunID: runID, Checkpoint: next}
	}

	return continueFromMessages(ctx, agent, messages, cp, store, runID)
}

func restoreAwaitingApproval(
	ctx context.Context,
	agent *Agent,
	cp *Checkpoint,
	store Checkpointer,
	runID string,
	registry AgentRegistry,
) (*Result, error) {
	// Reconstruct an InterruptedError from the checkpoint.
	ie := &InterruptedError{
		ToolCalls:        cp.PendingToolCalls,
		PendingApprovals: cp.PendingApprovals,
		Agent:            agent,
		Messages:         cp.Messages,
		Usage:            cp.Usage,
		Turns:            cp.Turns,
	}

	// Reconstruct outerState if this was a nested interruption.
	if len(cp.InnerCheckpoints) > 0 {
		if len(cp.InnerCheckpoints) > 1 {
			return nil, fmt.Errorf("cannot restore approval checkpoint: expected one inner checkpoint, got %d", len(cp.InnerCheckpoints))
		}
		for toolCallID, innerCP := range cp.InnerCheckpoints {
			innerAgent, err := registry.Agent(innerCP.AgentName)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve inner agent %q: %w", innerCP.AgentName, err)
			}

			innerIE := &InterruptedError{
				ToolCalls:        innerCP.PendingToolCalls,
				PendingApprovals: innerCP.PendingApprovals,
				Agent:            innerAgent,
				Messages:         innerCP.Messages,
				Usage:            innerCP.Usage,
				Turns:            innerCP.Turns,
			}

			ie.Agent = innerAgent
			ie.Messages = innerCP.Messages
			ie.Usage = innerCP.Usage
			ie.Turns = innerCP.Turns
			ie.ToolCalls = innerCP.PendingToolCalls
			ie.PendingApprovals = innerCP.PendingApprovals

			ie.outerState = &outerLoopState{
				agent:          agent,
				messages:       cp.Messages,
				usage:          cp.Usage,
				turns:          cp.Turns,
				allToolCalls:   cp.AllToolCalls,
				toolCallID:     toolCallID,
				completedCalls: cp.CompletedCalls,
				innerInterrupt: innerIE,
			}
			break
		}
	}

	if len(cp.ApprovalInput) > 0 {
		return resumeWithOpts(
			ctx,
			ie,
			ResumeInput{Approvals: cp.ApprovalInput},
			runOpts{
				callLLM:       blockingCallLLM,
				onEvent:       noopEvent,
				checkpointer:  store,
				runID:         runID,
				toolUsedInRun: cp.ToolUsedInRun,
			},
		)
	}

	return nil, ie
}
