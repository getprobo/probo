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

package probot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
)

type fakeCapability struct {
	name         string
	messageTypes []string
	prefixes     []string
	tools        []agent.Tool
	handled      bool
}

type renderOnlyCapability struct{}

func (renderOnlyCapability) Name() string {
	return "render-only"
}

func (renderOnlyCapability) MessageTypes() []string {
	return []string{"render-only"}
}

func (renderOnlyCapability) RenderMessage(
	context.Context,
	bot.Message,
) (bot.MessageIntent, error) {
	return bot.MessageIntent{
		FallbackText: "rendered",
	}, nil
}

func (renderOnlyCapability) BuildOutboundMessage(
	context.Context,
	gid.GID,
	string,
	map[string]any,
) (probot.OutboundMessage, error) {
	return probot.OutboundMessage{
		Intent: bot.MessageIntent{FallbackText: "event rendered"},
	}, nil
}

func (f *fakeCapability) Name() string {
	return f.name
}

func (f *fakeCapability) MessageTypes() []string {
	return f.messageTypes
}

func (f *fakeCapability) ActionPrefixes() []string {
	return f.prefixes
}

func (f *fakeCapability) Tools() []agent.Tool {
	return f.tools
}

func (f *fakeCapability) RenderMessage(
	context.Context,
	bot.Message,
) (bot.MessageIntent, error) {
	return bot.MessageIntent{
		FallbackText: f.name,
	}, nil
}

func (f *fakeCapability) HandleAction(
	context.Context,
	probot.Action,
) (probot.ActionResult, error) {
	f.handled = true
	return probot.ActionResult{Message: "handled"}, nil
}

func TestMessageAnchor_PreservesOpaqueProviderCoordinates(t *testing.T) {
	t.Parallel()

	anchor := probot.MessageAnchor{
		ConversationID: "provider/conversation",
		MessageID:      "provider:message",
	}

	assert.Equal(t, "provider/conversation", anchor.ConversationID)
	assert.Equal(t, "provider:message", anchor.MessageID)
}

func TestCapabilityRegistry_RejectsDuplicateOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		first  *fakeCapability
		second *fakeCapability
	}{
		{
			name:   "name",
			first:  &fakeCapability{name: "first"},
			second: &fakeCapability{name: "first"},
		},
		{
			name:   "message type",
			first:  &fakeCapability{name: "first", messageTypes: []string{"approval"}},
			second: &fakeCapability{name: "second", messageTypes: []string{"approval"}},
		},
		{
			name:   "action prefix",
			first:  &fakeCapability{name: "first", prefixes: []string{"first."}},
			second: &fakeCapability{name: "second", prefixes: []string{"first."}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				registry := probot.NewCapabilityRegistry()
				require.NoError(t, registry.Register(tt.first))
				require.Error(t, registry.Register(tt.second))
			},
		)
	}
}

func TestCapabilityRegistry_DispatchesRenderAndAction(t *testing.T) {
	t.Parallel()

	capability := &fakeCapability{
		name:         "first",
		messageTypes: []string{"approval"},
		prefixes:     []string{"first."},
	}
	registry := probot.NewCapabilityRegistry()
	require.NoError(t, registry.Register(capability))

	intent, err := registry.RenderMessage(
		context.Background(),
		bot.Message{Type: "approval"},
	)
	require.NoError(t, err)
	assert.Equal(t, "first", intent.FallbackText)

	result, err := registry.HandleAction(
		context.Background(),
		probot.Action{
			ID:      "first.approve",
			Message: bot.Message{Type: "approval"},
		},
	)
	require.NoError(t, err)
	assert.True(t, capability.handled)
	assert.Equal(t, "handled", result.Message)
}

func TestCapabilityRegistry_AcceptsRendererWithoutToolsOrActions(t *testing.T) {
	t.Parallel()

	registry := probot.NewCapabilityRegistry()
	require.NoError(t, registry.Register(renderOnlyCapability{}))

	intent, err := registry.RenderMessage(
		context.Background(),
		bot.Message{Type: "render-only"},
	)
	require.NoError(t, err)
	assert.Equal(t, "rendered", intent.FallbackText)
	assert.Empty(t, registry.Tools())
}

func TestCapabilityRegistry_DispatchesOutboundMessage(t *testing.T) {
	t.Parallel()

	registry := probot.NewCapabilityRegistry()
	require.NoError(t, registry.Register(renderOnlyCapability{}))

	result, err := registry.BuildOutboundMessage(
		t.Context(),
		"render-only",
		gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		"render-only",
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "event rendered", result.Intent.FallbackText)
}

func TestCapabilityRegistry_ToolsForMessageType(t *testing.T) {
	t.Parallel()

	registry := probot.NewCapabilityRegistry()
	approvalTool := agent.FunctionTool(
		"approval_tool",
		"Handle an approval.",
		func(context.Context, struct{}) (agent.ToolResult, error) {
			return agent.ToolResult{}, nil
		},
	)
	require.NoError(
		t,
		registry.Register(
			&fakeCapability{
				name:         "approval",
				messageTypes: []string{"approval"},
				tools:        []agent.Tool{approvalTool},
			},
		),
	)

	tools := registry.ToolsForMessageType("approval")
	require.Len(t, tools, 1)
	assert.Equal(t, "approval_tool", tools[0].Name())
	assert.Empty(t, registry.ToolsForMessageType("unknown"))
}

func TestCapabilityRegistry_RejectsUnknownAction(t *testing.T) {
	t.Parallel()

	registry := probot.NewCapabilityRegistry()
	require.NoError(
		t,
		registry.Register(
			&fakeCapability{
				name:         "first",
				messageTypes: []string{"approval"},
				prefixes:     []string{"first."},
			},
		),
	)

	_, err := registry.HandleAction(
		context.Background(),
		probot.Action{
			ID:      "another.approve",
			Message: bot.Message{Type: "approval"},
		},
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, probot.ErrCapabilityNotFound))
}
