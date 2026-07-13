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
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
)

// RunHandler executes non-LLM agent runs scheduled through agent_runs.
type RunHandler interface {
	Run(ctx context.Context, run *coredata.AgentRun) (json.RawMessage, error)
}

// Registry resolves LLM agents and imperative run handlers by name.
type Registry struct {
	agents  map[string]*agent.Agent
	runners map[string]RunHandler
}

func NewRegistry() *Registry {
	return &Registry{
		agents:  make(map[string]*agent.Agent),
		runners: make(map[string]RunHandler),
	}
}

func (r *Registry) RegisterAgent(name string, a *agent.Agent) {
	r.agents[name] = a
}

func (r *Registry) RegisterRunHandler(name string, handler RunHandler) {
	r.runners[name] = handler
}

func (r *Registry) Agent(name string) (*agent.Agent, error) {
	a, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q not found", name)
	}

	return a, nil
}

func (r *Registry) RunHandler(name string) (RunHandler, bool) {
	handler, ok := r.runners[name]

	return handler, ok
}
