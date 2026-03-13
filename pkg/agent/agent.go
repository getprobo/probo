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
	"fmt"
	"io"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/llm"
)

const DefaultMaxTurns = 10

type (
	Option func(*Agent)

	Agent struct {
		name               string
		handoffDescription string
		instructions       string
		instructionsFunc   func(ctx context.Context, a *Agent) string
		model              string
		modelSettings      ModelSettings
		tools              []Tool
		handoffs           []*Handoff
		mcpServers         []*MCPServer
		maxTurns           int
		maxToolDepth       int
		client             *llm.Client
		logger             *log.Logger
		hooks              []RunHooks
		agentHooks         AgentHooks
		inputGuardrails    []InputGuardrail
		outputGuardrails   []OutputGuardrail
		session            Session
		sessionID          string
		outputType         *OutputType
		toolUseBehavior    ToolUseBehavior
		resetToolChoice    bool
		responseFormat     *llm.ResponseFormat
		approval           *ApprovalConfig
	}
)

func New(name string, client *llm.Client, opts ...Option) *Agent {
	a := &Agent{
		name:            name,
		client:          client,
		maxTurns:        DefaultMaxTurns,
		maxToolDepth:    DefaultMaxToolDepth,
		toolUseBehavior: RunLLMAgain(),
		resetToolChoice: true,
		logger:          log.NewLogger(log.WithOutput(io.Discard)),
	}

	for _, opt := range opts {
		opt(a)
	}

	a.logger = a.logger.Named("agent").With(log.String("agent", name))

	return a
}

func (a *Agent) Name() string {
	return a.name
}

func (a *Agent) AsTool(name, description string) Tool {
	return newAgentTool(a, name, description)
}

func (a *Agent) HandoffDescription() string {
	return a.handoffDescription
}

// Clone creates a shallow copy of the agent with the given options applied.
func (a *Agent) Clone(opts ...Option) *Agent {
	cp := *a

	cp.tools = make([]Tool, len(a.tools))
	copy(cp.tools, a.tools)

	cp.handoffs = make([]*Handoff, len(a.handoffs))
	copy(cp.handoffs, a.handoffs)

	cp.mcpServers = make([]*MCPServer, len(a.mcpServers))
	copy(cp.mcpServers, a.mcpServers)

	cp.hooks = make([]RunHooks, len(a.hooks))
	copy(cp.hooks, a.hooks)

	cp.inputGuardrails = make([]InputGuardrail, len(a.inputGuardrails))
	copy(cp.inputGuardrails, a.inputGuardrails)

	cp.outputGuardrails = make([]OutputGuardrail, len(a.outputGuardrails))
	copy(cp.outputGuardrails, a.outputGuardrails)

	if a.approval != nil {
		newApproval := ApprovalConfig{
			ShouldApprove: a.approval.ShouldApprove,
		}
		if len(a.approval.ToolNames) > 0 {
			newApproval.ToolNames = make([]string, len(a.approval.ToolNames))
			copy(newApproval.ToolNames, a.approval.ToolNames)
			newApproval.toolNameSet = buildToolNameSet(newApproval.ToolNames)
		}
		cp.approval = &newApproval
	}

	for _, opt := range opts {
		opt(&cp)
	}

	return &cp
}

// clone creates a shallow copy suitable for overriding pointer fields
// (e.g. responseFormat) without affecting the original. Slice fields remain
// shared; use the exported Clone method when slice mutations are needed.
func (a *Agent) clone() *Agent {
	cp := *a
	return &cp
}

func WithInstructions(s string) Option {
	return func(a *Agent) {
		a.instructions = s
		a.instructionsFunc = nil
	}
}

func WithInstructionsFunc(fn func(ctx context.Context, a *Agent) string) Option {
	return func(a *Agent) {
		a.instructionsFunc = fn
		a.instructions = ""
	}
}

func WithHandoffDescription(desc string) Option {
	return func(a *Agent) {
		a.handoffDescription = desc
	}
}

func WithModel(m string) Option {
	return func(a *Agent) {
		a.model = m
	}
}

func WithModelSettings(s ModelSettings) Option {
	return func(a *Agent) {
		a.modelSettings = s
	}
}

func WithTools(tools ...Tool) Option {
	return func(a *Agent) {
		a.tools = append(a.tools, tools...)
	}
}

func WithHandoffs(agents ...*Agent) Option {
	return func(a *Agent) {
		for _, ag := range agents {
			a.handoffs = append(a.handoffs, &Handoff{Agent: ag})
		}
	}
}

func WithHandoffConfigs(handoffs ...*Handoff) Option {
	return func(a *Agent) {
		a.handoffs = append(a.handoffs, handoffs...)
	}
}

func WithMaxTurns(n int) Option {
	return func(a *Agent) {
		if n < 1 {
			n = 1
		}
		a.maxTurns = n
	}
}

func WithMaxToolDepth(n int) Option {
	return func(a *Agent) {
		if n < 1 {
			n = 1
		}
		a.maxToolDepth = n
	}
}

func WithTemperature(t float64) Option {
	return func(a *Agent) {
		a.modelSettings.Temperature = &t
	}
}

func WithTopP(p float64) Option {
	return func(a *Agent) {
		a.modelSettings.TopP = &p
	}
}

func WithFrequencyPenalty(p float64) Option {
	return func(a *Agent) {
		a.modelSettings.FrequencyPenalty = &p
	}
}

func WithPresencePenalty(p float64) Option {
	return func(a *Agent) {
		a.modelSettings.PresencePenalty = &p
	}
}

func WithMaxTokens(n int) Option {
	return func(a *Agent) {
		a.modelSettings.MaxTokens = &n
	}
}

func WithToolChoice(tc llm.ToolChoice) Option {
	return func(a *Agent) {
		a.modelSettings.ToolChoice = &tc
	}
}

func WithParallelToolCalls(enabled bool) Option {
	return func(a *Agent) {
		a.modelSettings.ParallelToolCalls = &enabled
	}
}

func WithLogger(l *log.Logger) Option {
	return func(a *Agent) {
		a.logger = l
	}
}

func WithHooks(hooks ...RunHooks) Option {
	return func(a *Agent) {
		a.hooks = append(a.hooks, hooks...)
	}
}

func WithAgentHooks(hooks AgentHooks) Option {
	return func(a *Agent) {
		a.agentHooks = hooks
	}
}

func WithInputGuardrails(guards ...InputGuardrail) Option {
	return func(a *Agent) {
		a.inputGuardrails = append(a.inputGuardrails, guards...)
	}
}

func WithOutputGuardrails(guards ...OutputGuardrail) Option {
	return func(a *Agent) {
		a.outputGuardrails = append(a.outputGuardrails, guards...)
	}
}

func WithSession(s Session, sessionID string) Option {
	return func(a *Agent) {
		a.session = s
		a.sessionID = sessionID
	}
}

func WithOutputType(t *OutputType) Option {
	return func(a *Agent) {
		a.outputType = t
	}
}

func WithToolUseBehavior(b ToolUseBehavior) Option {
	return func(a *Agent) {
		a.toolUseBehavior = b
	}
}

func WithResetToolChoice(reset bool) Option {
	return func(a *Agent) {
		a.resetToolChoice = reset
	}
}

func WithMCPServers(servers ...*MCPServer) Option {
	return func(a *Agent) {
		for _, s := range servers {
			if s != nil {
				a.mcpServers = append(a.mcpServers, s)
			}
		}
	}
}

func WithApproval(config ApprovalConfig) Option {
	config.toolNameSet = buildToolNameSet(config.ToolNames)
	return func(a *Agent) {
		a.approval = &config
	}
}

func (a *Agent) resolveTools(ctx context.Context) ([]ToolDescriptor, map[string]ToolDescriptor, error) {
	var all []ToolDescriptor

	for _, t := range a.tools {
		all = append(all, t)
	}

	for _, h := range a.handoffs {
		all = append(all, h.tool())
	}

	for _, s := range a.mcpServers {
		mcpTools, err := s.Tools(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot resolve MCP tools from %q: %w", s.name, err)
		}
		for _, t := range mcpTools {
			all = append(all, t)
		}
	}

	toolMap := make(map[string]ToolDescriptor, len(all))
	for _, t := range all {
		name := t.Name()
		if _, exists := toolMap[name]; exists {
			return nil, nil, fmt.Errorf("cannot resolve tools: duplicate tool name %q", name)
		}
		toolMap[name] = t
	}

	return all, toolMap, nil
}

func (a *Agent) buildSystemPrompt(ctx context.Context) string {
	instr := a.instructions
	if a.instructionsFunc != nil {
		instr = a.instructionsFunc(ctx, a)
	}

	data := systemPromptData{
		Instructions: instr,
	}

	for _, h := range a.handoffs {
		desc := h.toolDescription()
		if len([]rune(desc)) > 200 {
			desc = string([]rune(desc)[:200]) + "..."
		}

		data.Handoffs = append(
			data.Handoffs,
			systemPromptHandoff{
				Name:        h.Agent.name,
				Description: desc,
			},
		)
	}

	return buildSystemPrompt(data)
}
