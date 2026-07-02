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

package agent

import (
	"context"

	"go.probo.inc/probo/pkg/llm"
)

// RunHooks receives callbacks on lifecycle events for the entire agent run.
type RunHooks interface {
	OnRunStart(ctx context.Context, agent *Agent, messages []llm.Message)
	OnRunEnd(ctx context.Context, agent *Agent, result *Result, err error)
	OnRunRestore(ctx context.Context, agent *Agent, checkpoint *Checkpoint)
	OnRunSnapshot(ctx context.Context, agent *Agent, checkpoint *Checkpoint)
	OnLLMStart(ctx context.Context, agent *Agent, messages []llm.Message)
	OnLLMEnd(ctx context.Context, agent *Agent, response *llm.ChatCompletionResponse, err error)
	OnToolStart(ctx context.Context, agent *Agent, tool Tool, arguments string)
	OnToolEnd(ctx context.Context, agent *Agent, tool Tool, result ToolResult, err error)
	OnHandoff(ctx context.Context, from *Agent, to *Agent)
	OnGuardrailTripped(ctx context.Context, agent *Agent, name string, result *GuardrailResult)
}

// NoOpHooks is a RunHooks implementation that does nothing.
type NoOpHooks struct{}

var _ RunHooks = NoOpHooks{}

func (NoOpHooks) OnRunStart(context.Context, *Agent, []llm.Message)                    {}
func (NoOpHooks) OnRunEnd(context.Context, *Agent, *Result, error)                     {}
func (NoOpHooks) OnRunRestore(context.Context, *Agent, *Checkpoint)                    {}
func (NoOpHooks) OnRunSnapshot(context.Context, *Agent, *Checkpoint)                   {}
func (NoOpHooks) OnLLMStart(context.Context, *Agent, []llm.Message)                    {}
func (NoOpHooks) OnLLMEnd(context.Context, *Agent, *llm.ChatCompletionResponse, error) {}
func (NoOpHooks) OnToolStart(context.Context, *Agent, Tool, string)                    {}
func (NoOpHooks) OnToolEnd(context.Context, *Agent, Tool, ToolResult, error)           {}
func (NoOpHooks) OnHandoff(context.Context, *Agent, *Agent)                            {}
func (NoOpHooks) OnGuardrailTripped(context.Context, *Agent, string, *GuardrailResult) {}

// AgentHooks receives callbacks on lifecycle events for a specific agent.
// Set via WithAgentHooks on an individual agent.
type AgentHooks interface {
	OnStart(ctx context.Context, agent *Agent)
	OnEnd(ctx context.Context, agent *Agent, output string)
	OnHandoff(ctx context.Context, agent *Agent, source *Agent)
	OnToolStart(ctx context.Context, agent *Agent, tool Tool)
	OnToolEnd(ctx context.Context, agent *Agent, tool Tool, result ToolResult)
	OnLLMStart(ctx context.Context, agent *Agent, messages []llm.Message)
	OnLLMEnd(ctx context.Context, agent *Agent, response *llm.ChatCompletionResponse, err error)
}

// NoOpAgentHooks is an AgentHooks implementation that does nothing.
type NoOpAgentHooks struct{}

var _ AgentHooks = NoOpAgentHooks{}

func (NoOpAgentHooks) OnStart(context.Context, *Agent)                                      {}
func (NoOpAgentHooks) OnEnd(context.Context, *Agent, string)                                {}
func (NoOpAgentHooks) OnHandoff(context.Context, *Agent, *Agent)                            {}
func (NoOpAgentHooks) OnToolStart(context.Context, *Agent, Tool)                            {}
func (NoOpAgentHooks) OnToolEnd(context.Context, *Agent, Tool, ToolResult)                  {}
func (NoOpAgentHooks) OnLLMStart(context.Context, *Agent, []llm.Message)                    {}
func (NoOpAgentHooks) OnLLMEnd(context.Context, *Agent, *llm.ChatCompletionResponse, error) {}
