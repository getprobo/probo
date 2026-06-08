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

package agentrun_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

func TestWorker_PicksUpAndCompletes(t *testing.T) {
	client := test.PGClient(t)
	ag := newDummyAgent(
		"echo-agent",
		[]*llm.ChatCompletionResponse{
			stopResponse("Done."),
		},
	)

	run := insertPendingRun(
		t,
		client,
		"echo-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "go"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"echo-agent": ag}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

func TestWorker_StopAndResume(t *testing.T) {
	client := test.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolReady)
			<-toolRelease

			return agent.ToolResult{Content: "work done"}, nil
		},
	)

	ag := newDummyAgent(
		"worker-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_1",
				Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
			}),
			stopResponse("All done after resume."),
		},
		slowTool,
	)

	run := insertPendingRun(
		t,
		client,
		"worker-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do work"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
	)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	go func() { _ = runWorker.Run(ctx1) }()

	select {
	case <-toolReady:
	case <-ctx1.Done():
		t.Fatal("timed out waiting for tool to start")
	}

	cancel1()

	select {
	case <-runWorker.ShutdownBroadcast():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker shutdown broadcast")
	}

	close(toolRelease)

	// Graceful shutdown must commit the run back to PENDING (with its
	// checkpoint intact) so another worker resumes it. Nothing relies on
	// a lease timeout to requeue it.
	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentRunStatusPending &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	suspended := loadAgentRun(t, client, run.ID)
	assert.Equal(
		t,
		coredata.AgentRunStatusPending,
		suspended.Status,
		"graceful shutdown must requeue the run as PENDING without manual recovery",
	)
	assert.Nil(t, suspended.Result)
	assert.Nil(t, suspended.ErrorMessage)

	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusSuspended, cp.Status)

	// No manual reset: the run is already PENDING from the graceful
	// shutdown, so a fresh worker must pick it up and resume on its own.
	runWorker2 := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
	)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	go func() { _ = runWorker2.Run(ctx2) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

// TestWorker_AwaitsApprovalDoesNotFail covers the regression where a tool
// call requiring approval surfaced as InterruptedError and was committed
// as FAILED. The run must instead park in AWAITING_APPROVAL with its
// checkpoint (and the pending approvals) preserved, and must not be
// re-claimed while it rests.
func TestWorker_AwaitsApprovalDoesNotFail(t *testing.T) {
	client := test.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	dangerTool := agent.FunctionTool[struct{}](
		"danger",
		"Performs a dangerous action",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			return agent.ToolResult{Content: "must not run before approval"}, nil
		},
	)

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_danger",
				Function: llm.FunctionCall{Name: "danger", Arguments: `{}`},
			}),
		},
	}

	ag := agent.New(
		"approval-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(dangerTool),
		agent.WithApproval(agent.ApprovalConfig{ToolNames: []string{"danger"}}),
	)

	run := insertPendingRun(
		t,
		client,
		"approval-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do the dangerous thing"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"approval-agent": ag}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusAwaitingApproval
		},
		10*time.Second,
		200*time.Millisecond,
	)

	awaiting := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusAwaitingApproval, awaiting.Status)
	assert.Nil(t, awaiting.Result)
	assert.Nil(t, awaiting.ErrorMessage)
	assert.NotNil(t, awaiting.Checkpoint)

	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusAwaitingApproval, cp.Status)
	require.Len(t, cp.PendingApprovals, 1)
	assert.Equal(t, "danger", cp.PendingApprovals[0].Function.Name)

	// The run must stay parked: only one mock response exists, so a
	// re-claim would error with "no more mock responses" and flip it to
	// FAILED. Confirm it holds AWAITING_APPROVAL.
	time.Sleep(time.Second)

	stillAwaiting := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusAwaitingApproval, stillAwaiting.Status)
}

// TestWorker_ApprovalApprovedResumesAndCompletes is the full happy-path
// approval cycle: the run parks in AWAITING_APPROVAL, a decision approves
// the pending tool call via the service, and the same worker resumes from
// the checkpoint, executes the approved tool, and completes.
func TestWorker_ApprovalApprovedResumesAndCompletes(t *testing.T) {
	client := test.PGClient(t)

	var executed atomic.Bool

	dangerTool := agent.FunctionTool[struct{}](
		"danger",
		"Performs a dangerous action",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			executed.Store(true)

			return agent.ToolResult{Content: "danger executed"}, nil
		},
	)

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_danger",
				Function: llm.FunctionCall{Name: "danger", Arguments: `{}`},
			}),
			stopResponse("all done"),
		},
	}

	ag := agent.New(
		"approval-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(dangerTool),
		agent.WithApproval(agent.ApprovalConfig{ToolNames: []string{"danger"}}),
	)

	run := insertPendingRun(
		t,
		client,
		"approval-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "go"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"approval-agent": ag}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusAwaitingApproval
		},
		10*time.Second,
		200*time.Millisecond,
	)

	svc := agentrun.NewService(client)
	_, err := svc.SubmitApproval(
		context.Background(),
		coredata.NewNoScope(),
		run.ID,
		map[string]agent.ApprovalResult{"tc_danger": {Approved: true}},
	)
	require.NoError(t, err)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
	assert.True(t, executed.Load(), "approved tool must execute on resume")
}

// TestWorker_ApprovalDeniedResumesAndCompletes covers the denial path: the
// run resumes without executing the gated tool and completes, with the
// denial fed back to the model as the tool result.
func TestWorker_ApprovalDeniedResumesAndCompletes(t *testing.T) {
	client := test.PGClient(t)

	var executed atomic.Bool

	dangerTool := agent.FunctionTool[struct{}](
		"danger",
		"Performs a dangerous action",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			executed.Store(true)

			return agent.ToolResult{Content: "danger executed"}, nil
		},
	)

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_danger",
				Function: llm.FunctionCall{Name: "danger", Arguments: `{}`},
			}),
			stopResponse("acknowledged the denial"),
		},
	}

	ag := agent.New(
		"approval-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(dangerTool),
		agent.WithApproval(agent.ApprovalConfig{ToolNames: []string{"danger"}}),
	)

	run := insertPendingRun(
		t,
		client,
		"approval-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "go"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"approval-agent": ag}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusAwaitingApproval
		},
		10*time.Second,
		200*time.Millisecond,
	)

	svc := agentrun.NewService(client)
	_, err := svc.SubmitApproval(
		context.Background(),
		coredata.NewNoScope(),
		run.ID,
		map[string]agent.ApprovalResult{"tc_danger": {Approved: false, Message: "denied by reviewer"}},
	)
	require.NoError(t, err)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
	assert.False(t, executed.Load(), "denied tool must not execute on resume")
}

// TestWorker_StopAndResumeAcrossHandoff exercises tree suspension where the
// active branch is a handed-off child agent. The checkpoint must record the
// child as active, and restore must resolve it from the registry so the
// resumed run continues in that branch and completes.
func TestWorker_StopAndResumeAcrossHandoff(t *testing.T) {
	client := test.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolReady)
			<-toolRelease

			return agent.ToolResult{Content: "child work done"}, nil
		},
	)

	childAgent := newDummyAgent(
		"child-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_child",
				Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
			}),
			stopResponse("child done"),
		},
		slowTool,
	)

	rootProvider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_root",
				Function: llm.FunctionCall{Name: "transfer_to_child_agent", Arguments: `{}`},
			}),
		},
	}

	rootAgent := agent.New(
		"root-agent",
		newTestClient(rootProvider),
		agent.WithModel("test-model"),
		agent.WithHandoffs(childAgent),
	)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{
			"root-agent":  rootAgent,
			"child-agent": childAgent,
		},
	}

	run := insertPendingRun(
		t,
		client,
		"root-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do work"}}}},
	)

	runWorker := newTestWorker(client, registry)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	go func() { _ = runWorker.Run(ctx1) }()

	select {
	case <-toolReady:
	case <-ctx1.Done():
		t.Fatal("timed out waiting for child agent tool to start")
	}

	cancel1()

	select {
	case <-runWorker.ShutdownBroadcast():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker shutdown broadcast")
	}

	close(toolRelease)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentRunStatusPending &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusSuspended, cp.Status)
	assert.Equal(
		t,
		"child-agent",
		cp.AgentName,
		"checkpoint must record the handed-off child as the active agent",
	)

	runWorker2 := newTestWorker(client, registry)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	go func() { _ = runWorker2.Run(ctx2) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

func TestWorker_StopAndResumeNestedSubAgent(t *testing.T) {
	client := test.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	var readyOnce sync.Once

	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			readyOnce.Do(func() { close(toolReady) })
			<-toolRelease

			return agent.ToolResult{Content: "inner work done"}, nil
		},
	)

	innerAgent := newDummyAgent(
		"inner-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_inner",
					Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
				},
			),
			stopResponse("inner done"),
		},
		slowTool,
	)

	outerAgent := newDummyAgent(
		"outer-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_outer",
					Function: llm.FunctionCall{Name: "call_inner", Arguments: `{"input":"delegate"}`},
				},
			),
			stopResponse("outer done"),
		},
		innerAgent.AsTool("call_inner", "Call inner"),
	)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{
			"outer-agent": outerAgent,
			"inner-agent": innerAgent,
		},
	}

	run := insertPendingRun(
		t,
		client,
		"outer-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do work"}}}},
	)

	runWorker := newTestWorker(client, registry)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	go func() { _ = runWorker.Run(ctx1) }()

	select {
	case <-toolReady:
	case <-ctx1.Done():
		t.Fatal("timed out waiting for nested sub-agent tool to start")
	}

	cancel1()

	select {
	case <-runWorker.ShutdownBroadcast():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker shutdown broadcast")
	}

	close(toolRelease)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentRunStatusPending &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusSuspended, cp.Status)
	assert.Equal(t, "outer-agent", cp.AgentName)

	innerCP, ok := cp.InnerCheckpoints["tc_outer"]
	require.True(t, ok, "expected nested checkpoint for outer tool call")
	require.NotNil(t, innerCP)
	assert.Equal(t, "inner-agent", innerCP.AgentName)
	assert.Equal(t, agent.AgentStatusSuspended, innerCP.Status)

	runWorker2 := newTestWorker(client, registry)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	go func() { _ = runWorker2.Run(ctx2) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

func TestWorker_StopAndResumeNestedSubAgentMultiLevel(t *testing.T) {
	client := test.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	var readyOnce sync.Once

	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			readyOnce.Do(func() { close(toolReady) })
			<-toolRelease

			return agent.ToolResult{Content: "grandchild work done"}, nil
		},
	)

	grandchildAgent := newDummyAgent(
		"grandchild-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_grandchild",
					Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
				},
			),
			stopResponse("grandchild done"),
		},
		slowTool,
	)

	childAgent := newDummyAgent(
		"child-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_child",
					Function: llm.FunctionCall{Name: "call_grandchild", Arguments: `{"input":"delegate deeper"}`},
				},
			),
			stopResponse("child done"),
		},
		grandchildAgent.AsTool("call_grandchild", "Call grandchild"),
	)

	outerAgent := newDummyAgent(
		"outer-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_outer",
					Function: llm.FunctionCall{Name: "call_child", Arguments: `{"input":"delegate"}`},
				},
			),
			stopResponse("outer done"),
		},
		childAgent.AsTool("call_child", "Call child"),
	)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{
			"outer-agent":      outerAgent,
			"child-agent":      childAgent,
			"grandchild-agent": grandchildAgent,
		},
	}

	run := insertPendingRun(
		t,
		client,
		"outer-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do work"}}}},
	)

	runWorker := newTestWorker(client, registry)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	go func() { _ = runWorker.Run(ctx1) }()

	select {
	case <-toolReady:
	case <-ctx1.Done():
		t.Fatal("timed out waiting for grandchild tool to start")
	}

	cancel1()

	select {
	case <-runWorker.ShutdownBroadcast():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker shutdown broadcast")
	}

	close(toolRelease)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentRunStatusPending &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, "outer-agent", cp.AgentName)

	childCP, ok := cp.InnerCheckpoints["tc_outer"]
	require.True(t, ok)
	require.NotNil(t, childCP)
	assert.Equal(t, "child-agent", childCP.AgentName)

	grandchildCP, ok := childCP.InnerCheckpoints["tc_child"]
	require.True(t, ok)
	require.NotNil(t, grandchildCP)
	assert.Equal(t, "grandchild-agent", grandchildCP.AgentName)
	assert.Equal(t, agent.AgentStatusSuspended, grandchildCP.Status)

	runWorker2 := newTestWorker(client, registry)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	go func() { _ = runWorker2.Run(ctx2) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

// TestWorker_ReclaimedRunDoesNotClobberWinner simulates the residual
// manual-recovery risk now that leasing is gone: a human moves a still
// in-flight run back to PENDING (resetRunToPending) while worker A is
// blocked in a tool. Worker B then claims and finishes it. When worker A
// finally returns, its commit must be discarded because the row is no
// longer RUNNING. The CommitAgentRunResult `status = 'RUNNING'` guard is
// the only fence protecting the winner's result.
func TestWorker_ReclaimedRunDoesNotClobberWinner(t *testing.T) {
	client := test.PGClient(t)

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolReady)
			<-toolRelease

			return agent.ToolResult{Content: "work done"}, nil
		},
	)

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_1",
				Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
			}),
			stopResponse("winner result"),
			stopResponse("stale result"),
		},
	}

	ag := agent.New(
		"worker-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(slowTool),
	)

	run := insertPendingRun(
		t,
		client,
		"worker-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do work"}}}},
	)

	runWorkerA := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
		agentrun.WithWorkerMaxConcurrency(1),
	)

	ctxA, cancelA := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelA()

	go func() { _ = runWorkerA.Run(ctxA) }()

	select {
	case <-toolReady:
	case <-ctxA.Done():
		t.Fatal("timed out waiting for first worker tool call")
	}

	resetRunToPending(t, client, run.ID)

	runWorkerB := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
		agentrun.WithWorkerMaxConcurrency(1),
	)

	ctxB, cancelB := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelB()

	go func() { _ = runWorkerB.Run(ctxB) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		15*time.Second,
		200*time.Millisecond,
	)

	winner := loadAgentRun(t, client, run.ID)
	winnerResult := append(json.RawMessage(nil), winner.Result...)
	require.NotNil(t, winnerResult)

	close(toolRelease)

	require.Eventually(
		t,
		func() bool {
			provider.mu.Lock()
			defer provider.mu.Unlock()

			return provider.calls >= 3
		},
		15*time.Second,
		200*time.Millisecond,
	)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			if err != nil {
				return false
			}

			return r.Status == coredata.AgentRunStatusCompleted && string(r.Result) == string(winnerResult)
		},
		10*time.Second,
		200*time.Millisecond,
	)
}

func TestWorker_UnknownAgentFails(t *testing.T) {
	client := test.PGClient(t)

	run := insertPendingRun(
		t,
		client,
		"missing-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "go"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusFailed
		},
		10*time.Second,
		200*time.Millisecond,
	)

	failed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusFailed, failed.Status)
	require.NotNil(t, failed.ErrorMessage)
	assert.Contains(t, *failed.ErrorMessage, "cannot resolve agent")
}

func TestWorker_InvalidInputMessagesFails(t *testing.T) {
	client := test.PGClient(t)
	ag := newDummyAgent(
		"worker-agent",
		[]*llm.ChatCompletionResponse{
			stopResponse("Done."),
		},
	)

	run := insertPendingRun(
		t,
		client,
		"worker-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "go"}}}},
	)

	overwriteRunInputMessagesRaw(t, client, run.ID, `"invalid-json"`)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusFailed
		},
		10*time.Second,
		200*time.Millisecond,
	)

	failed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusFailed, failed.Status)
	require.NotNil(t, failed.ErrorMessage)
	assert.Contains(t, *failed.ErrorMessage, "cannot unmarshal input messages")
}

func TestWorker_SIGTERM(t *testing.T) {
	if os.Getenv("TEST_SIGTERM_SUBPROCESS") == "1" {
		runSIGTERMSubprocess(t)
		return
	}

	// Skip when the test database is unreachable so the parent does not
	// wait on a subprocess that skips itself for the same reason and never
	// prints READY.
	test.PGClient(t)

	cmd := exec.Command(os.Args[0], "-test.run=^TestWorker_SIGTERM$")

	cmd.Env = append(os.Environ(), "TEST_SIGTERM_SUBPROCESS=1")

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	cmd.Stderr = cmd.Stdout

	require.NoError(t, cmd.Start())

	ready := make(chan struct{})
	scanDone := make(chan struct{})

	var (
		linesMu sync.Mutex
		lines   []string
	)

	snapshotLines := func() string {
		linesMu.Lock()
		defer linesMu.Unlock()

		return strings.Join(lines, "\n")
	}

	go func() {
		defer close(scanDone)

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			linesMu.Lock()

			lines = append(lines, line)
			linesMu.Unlock()

			if line == "READY" {
				close(ready)
			}
		}
	}()

	select {
	case <-ready:
	case <-time.After(20 * time.Second):
		_ = cmd.Process.Kill()

		t.Fatalf("subprocess did not become ready for SIGTERM\n%s", snapshotLines())
	}

	require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))

	if err := cmd.Wait(); err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, snapshotLines())
	}

	<-scanDone
}

func runSIGTERMSubprocess(t *testing.T) {
	client := test.PGClient(t)

	workStarted := make(chan struct{})

	ag := newDummyAgent(
		"battle-agent",
		battleTestResponses(),
		makeBattleTools(workStarted)...,
	)

	run := insertPendingRun(
		t,
		client,
		"battle-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "start"}}}},
	)

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"battle-agent": ag}},
		agentrun.WithWorkerInterval(150*time.Millisecond),
		agentrun.WithWorkerMaxConcurrency(1),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	go func() {
		_ = runWorker.Run(ctx)
	}()

	select {
	case <-workStarted:
	case <-time.After(15 * time.Second):
		t.Fatal("tool did not start before SIGTERM")
	}

	_, _ = fmt.Fprintln(os.Stdout, "READY")

	select {
	case <-runWorker.ShutdownBroadcast():
	case <-time.After(15 * time.Second):
		t.Fatal("worker did not broadcast shutdown after SIGTERM")
	}

	time.Sleep(300 * time.Millisecond)

	// The in-flight run may checkpoint or be recovered later depending on
	// timing, but it must still be queryable after graceful shutdown.
	_, err := tryLoadAgentRun(client, run.ID)
	require.NoError(t, err)
}

type workInput struct {
	Step string `json:"step"`
}

func makeBattleTools(workStarted chan<- struct{}) []agent.Tool {
	return []agent.Tool{
		agent.FunctionTool[workInput](
			"do_work",
			"Performs interruptible work",
			func(ctx context.Context, _ workInput) (agent.ToolResult, error) {
				close(workStarted)
				<-ctx.Done()

				return agent.ToolResult{Content: "interrupted"}, ctx.Err()
			},
		),
	}
}

func battleTestResponses() []*llm.ChatCompletionResponse {
	return []*llm.ChatCompletionResponse{
		toolCallResponse(llm.ToolCall{
			ID: "tc_battle_1",
			Function: llm.FunctionCall{
				Name:      "do_work",
				Arguments: `{"step":"one"}`,
			},
		}),
		stopResponse("done"),
	}
}
