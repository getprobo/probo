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

package probot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var (
	ErrExecutionAdapterDuplicate = errors.New("probot execution adapter already registered")
	ErrExecutionAdapterUnknown   = errors.New("probot execution adapter not found")
	ErrAgentProfileDuplicate     = errors.New("probot agent profile already registered")
	ErrAgentProfileUnknown       = errors.New("probot agent profile not found")
)

type (
	ExecutionAdapter interface {
		Provider() string
		Prepare(
			ctx context.Context,
			execution *coredata.AgentExecution,
			registry agent.AgentRegistry,
			input *coredata.AgentInput,
		) (context.Context, agent.AgentRegistry, error)
	}

	OutboundDelivery struct {
		OrganizationID   gid.GID
		Purpose          coredata.BotMessagePurpose
		SourceEventID    string
		SubjectNamespace string
		SubjectKey       string
		Capability       string
		MessageType      string
		Attributes       map[string]any
		Result           OutboundMessage
	}

	MessageDelivery interface {
		DeliverOutbound(ctx context.Context, delivery OutboundDelivery) error
	}

	ExecutionAdapterRegistry struct {
		mu         sync.RWMutex
		byProvider map[string]ExecutionAdapter
	}

	AgentProfileRegistry struct {
		mu     sync.RWMutex
		agents map[string]*agent.Agent
	}

	staticAgentRegistry struct {
		name  string
		agent *agent.Agent
	}
)

func NewExecutionAdapterRegistry() *ExecutionAdapterRegistry {
	return &ExecutionAdapterRegistry{byProvider: make(map[string]ExecutionAdapter)}
}

func (r *ExecutionAdapterRegistry) Register(adapter ExecutionAdapter) error {
	if adapter == nil {
		return fmt.Errorf("cannot register nil Probot execution adapter")
	}

	provider := strings.TrimSpace(adapter.Provider())
	if provider == "" {
		return fmt.Errorf("cannot register Probot execution adapter without provider")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byProvider[provider]; ok {
		return fmt.Errorf("%w: %q", ErrExecutionAdapterDuplicate, provider)
	}

	r.byProvider[provider] = adapter

	return nil
}

func (r *ExecutionAdapterRegistry) Prepare(
	ctx context.Context,
	execution *coredata.AgentExecution,
	registry agent.AgentRegistry,
	input *coredata.AgentInput,
) (context.Context, agent.AgentRegistry, error) {
	if execution.ExecutionKind == coredata.AgentExecutionKindOneShot {
		return ctx, registry, nil
	}

	if execution.Source == nil || *execution.Source == "" {
		return nil, nil, fmt.Errorf("%w: execution source is empty", ErrExecutionAdapterUnknown)
	}

	r.mu.RLock()
	adapter, ok := r.byProvider[*execution.Source]
	r.mu.RUnlock()

	if !ok {
		return nil, nil, fmt.Errorf("%w: %q", ErrExecutionAdapterUnknown, *execution.Source)
	}

	return adapter.Prepare(ctx, execution, registry, input)
}

func NewAgentProfileRegistry() *AgentProfileRegistry {
	return &AgentProfileRegistry{agents: make(map[string]*agent.Agent)}
}

func (r *AgentProfileRegistry) Register(name string, a *agent.Agent) error {
	name = strings.TrimSpace(name)
	if name == "" || a == nil {
		return fmt.Errorf("cannot register empty Probot agent profile")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.agents[name]; ok {
		return fmt.Errorf("%w: %q", ErrAgentProfileDuplicate, name)
	}

	r.agents[name] = a

	return nil
}

func (r *AgentProfileRegistry) Agent(name string) (*agent.Agent, error) {
	r.mu.RLock()
	a, ok := r.agents[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrAgentProfileUnknown, name)
	}

	return a, nil
}

func NewStaticAgentRegistry(name string, a *agent.Agent) agent.AgentRegistry {
	return &staticAgentRegistry{name: name, agent: a}
}

func (r *staticAgentRegistry) Agent(name string) (*agent.Agent, error) {
	if name != r.name || r.agent == nil {
		return nil, fmt.Errorf("%w: %q", ErrAgentProfileUnknown, name)
	}

	return r.agent, nil
}
