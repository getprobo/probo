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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
)

type testExecutionAdapter struct {
	called bool
}

func (a *testExecutionAdapter) Provider() string {
	return "test"
}

func (a *testExecutionAdapter) Prepare(
	ctx context.Context,
	_ *coredata.AgentExecution,
	registry agent.AgentRegistry,
	_ *coredata.AgentInput,
) (context.Context, agent.AgentRegistry, error) {
	a.called = true
	return ctx, registry, nil
}

func TestExecutionAdapterRegistry(t *testing.T) {
	t.Parallel()

	registry := NewExecutionAdapterRegistry()
	adapter := &testExecutionAdapter{}
	require.NoError(t, registry.Register(adapter))
	assert.ErrorIs(t, registry.Register(&testExecutionAdapter{}), ErrExecutionAdapterDuplicate)

	source := "test"
	_, _, err := registry.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			Source: &source,
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.True(t, adapter.called)

	unknown := "unknown"
	_, _, err = registry.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			Source: &unknown,
		},
		nil,
		nil,
	)
	assert.ErrorIs(t, err, ErrExecutionAdapterUnknown)
}

func TestExecutionAdapterRegistryRequiresSource(t *testing.T) {
	t.Parallel()

	registry := NewExecutionAdapterRegistry()
	_, _, err := registry.Prepare(
		t.Context(),
		&coredata.AgentExecution{},
		nil,
		nil,
	)
	assert.ErrorIs(t, err, ErrExecutionAdapterUnknown)
}

func TestAgentProfileRegistry(t *testing.T) {
	t.Parallel()

	registry := NewAgentProfileRegistry()
	profile := agent.New("Probot", nil)
	require.NoError(t, registry.Register("probot", profile))
	assert.ErrorIs(t, registry.Register("probot", profile), ErrAgentProfileDuplicate)

	resolved, err := registry.Agent("probot")
	require.NoError(t, err)
	assert.Same(t, profile, resolved)

	_, err = registry.Agent("unknown")
	assert.ErrorIs(t, err, ErrAgentProfileUnknown)
}
