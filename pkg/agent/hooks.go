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

	"go.probo.inc/probo/pkg/llm"
)

// RunHooks receives callbacks on lifecycle events for the entire agent run.
type RunHooks interface {
	OnRunStart(ctx context.Context, agent *Agent, messages []llm.Message)
	OnRunEnd(ctx context.Context, agent *Agent, result *Result, err error)
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
