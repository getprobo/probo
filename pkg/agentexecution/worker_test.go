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

package agentexecution_test

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
	"go.probo.inc/probo/pkg/agentexecution"
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

	run, input := insertPendingTurn(t, client, "echo-agent", "go")

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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

func TestWorker_StopAndResume(t *testing.T) {
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

	run, input := insertPendingTurn(t, client, "worker-agent", "do work")

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

	// Graceful shutdown must park the run in SUSPENDED (with its
	// checkpoint intact) so another worker resumes it. Nothing relies on
	// a lease timeout to requeue it.
	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentExecution(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentExecutionStatusSuspended &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	suspended := loadAgentExecution(t, client, run.ID)
	assert.Equal(
		t,
		coredata.AgentExecutionStatusSuspended,
		suspended.Status,
		"graceful shutdown must park the run as SUSPENDED without manual recovery",
	)
	assert.Nil(t, suspended.ErrorMessage)

	cp := loadCheckpoint(t, client, run.ID)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusSuspended, cp.Status)

	// No manual reset: the run is already SUSPENDED from the graceful
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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
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

	run, _ := insertPendingTurn(t, client, "approval-agent", "do the dangerous thing")

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
			r, err := tryLoadAgentExecution(client, run.ID)
			return err == nil && r.Status == coredata.AgentExecutionStatusAwaitingApproval
		},
		10*time.Second,
		200*time.Millisecond,
	)

	awaiting := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusAwaitingApproval, awaiting.Status)
	assert.Nil(t, awaiting.ErrorMessage)
	assert.NotNil(t, awaiting.Checkpoint)

	cp := loadCheckpoint(t, client, run.ID)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusAwaitingApproval, cp.Status)
	require.Len(t, cp.PendingApprovals, 1)
	assert.Equal(t, "danger", cp.PendingApprovals[0].Function.Name)

	// The run must stay parked: only one mock response exists, so a
	// re-claim would error with "no more mock responses" and flip it to
	// FAILED. Confirm it holds AWAITING_APPROVAL.
	time.Sleep(time.Second)

	stillAwaiting := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusAwaitingApproval, stillAwaiting.Status)
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

	run, input := insertPendingTurn(t, client, "approval-agent", "go")

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
			r, err := tryLoadAgentExecution(client, run.ID)
			return err == nil && r.Status == coredata.AgentExecutionStatusAwaitingApproval
		},
		10*time.Second,
		200*time.Millisecond,
	)

	svc := agentexecution.NewService(client)
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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
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

	run, input := insertPendingTurn(t, client, "approval-agent", "go")

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
			r, err := tryLoadAgentExecution(client, run.ID)
			return err == nil && r.Status == coredata.AgentExecutionStatusAwaitingApproval
		},
		10*time.Second,
		200*time.Millisecond,
	)

	svc := agentexecution.NewService(client)
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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
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

	run, input := insertPendingTurn(t, client, "root-agent", "do work")

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
			r, err := tryLoadAgentExecution(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentExecutionStatusSuspended &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	cp := loadCheckpoint(t, client, run.ID)
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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

func TestWorker_StopAndResumeNestedSubAgent(t *testing.T) {
	client := test.PGClient(t)

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

	run, input := insertPendingTurn(t, client, "outer-agent", "do work")

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
			r, err := tryLoadAgentExecution(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentExecutionStatusSuspended &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	cp := loadCheckpoint(t, client, run.ID)
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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

func TestWorker_StopAndResumeNestedSubAgentMultiLevel(t *testing.T) {
	client := test.PGClient(t)

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

	run, input := insertPendingTurn(t, client, "outer-agent", "do work")

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
			r, err := tryLoadAgentExecution(client, run.ID)

			return err == nil &&
				r.Status == coredata.AgentExecutionStatusSuspended &&
				r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
	)

	cp := loadCheckpoint(t, client, run.ID)
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
			return conversationSettled(client, run.ID, input.ID)
		},
		10*time.Second,
		200*time.Millisecond,
	)

	completed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, completed.Status)
	assert.NotNil(t, loadAgentInput(t, client, input.ID).ProcessedAt)
	assert.Nil(t, completed.Checkpoint)
	assert.Nil(t, completed.ErrorMessage)
}

// TestWorker_LeaseLossCancelsAndFencesStaleWriter simulates an execution
// being administratively recovered while its first worker remains blocked
// in a tool. The heartbeat detects owner-token loss, requests graceful
// suspension, and all stale checkpoint/result writes remain fenced.
func TestWorker_LeaseLossCancelsAndFencesStaleWriter(t *testing.T) {
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
		},
	}

	ag := agent.New(
		"worker-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(slowTool),
	)

	run, input := insertPendingTurn(t, client, "worker-agent", "do work")

	runWorkerA := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
		agentexecution.WithWorkerMaxConcurrency(1),
		agentexecution.WithWorkerHeartbeatInterval(50*time.Millisecond),
		agentexecution.WithWorkerStaleAfter(500*time.Millisecond),
	)

	ctxA, cancelA := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelA()

	go func() { _ = runWorkerA.Run(ctxA) }()

	select {
	case <-toolReady:
	case <-ctxA.Done():
		t.Fatal("timed out waiting for first worker tool call")
	}

	resetExecutionToIdle(t, client, run.ID)

	runWorkerB := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"worker-agent": ag}},
		agentexecution.WithWorkerMaxConcurrency(1),
	)

	ctxB, cancelB := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelB()

	go func() { _ = runWorkerB.Run(ctxB) }()

	require.Eventually(
		t,
		func() bool {
			return conversationSettled(client, run.ID, input.ID)
		},
		15*time.Second,
		200*time.Millisecond,
	)

	winner := loadAgentExecution(t, client, run.ID)
	winnerSession := append(json.RawMessage(nil), winner.SessionMessages...)
	require.NotEmpty(t, winnerSession)

	close(toolRelease)

	require.Eventually(
		t,
		func() bool {
			current := loadAgentExecution(t, client, run.ID)

			return current.Status == coredata.AgentExecutionStatusIdle &&
				string(current.SessionMessages) == string(winnerSession)
		},
		10*time.Second,
		200*time.Millisecond,
	)
}

func TestWorker_UnknownAgentFails(t *testing.T) {
	client := test.PGClient(t)

	run, _ := insertPendingTurn(t, client, "missing-agent", "go")

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
			r, err := tryLoadAgentExecution(client, run.ID)
			return err == nil && r.Status == coredata.AgentExecutionStatusFailed
		},
		10*time.Second,
		200*time.Millisecond,
	)

	failed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusFailed, failed.Status)
	require.NotNil(t, failed.ErrorMessage)
	assert.Contains(t, *failed.ErrorMessage, "cannot resolve agent")
}

func TestWorker_InvalidAgentInputFails(t *testing.T) {
	client := test.PGClient(t)
	ag := newDummyAgent(
		"worker-agent",
		[]*llm.ChatCompletionResponse{
			stopResponse("Done."),
		},
	)

	run, input := insertPendingTurn(t, client, "worker-agent", "go")

	overwriteAgentInputMessageRaw(t, client, input.ID, `"invalid-json"`)

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
			r, err := tryLoadAgentExecution(client, run.ID)
			return err == nil && r.Status == coredata.AgentExecutionStatusFailed
		},
		10*time.Second,
		200*time.Millisecond,
	)

	failed := loadAgentExecution(t, client, run.ID)
	assert.Equal(t, coredata.AgentExecutionStatusFailed, failed.Status)
	require.NotNil(t, failed.ErrorMessage)
	assert.Contains(t, *failed.ErrorMessage, "cannot unmarshal input")
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

	run, _ := insertPendingTurn(t, client, "battle-agent", "start")

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"battle-agent": ag}},
		agentexecution.WithWorkerInterval(150*time.Millisecond),
		agentexecution.WithWorkerMaxConcurrency(1),
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
	_, err := tryLoadAgentExecution(client, run.ID)
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
