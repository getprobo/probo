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

package slack_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

type recordingToolQueue struct {
	keys     []string
	kinds    []coredata.SlackDeliveryOperationKind
	payloads []map[string]any
}

func (q *recordingToolQueue) Queue(
	_ context.Context,
	_ gid.GID,
	operationKey string,
	kind coredata.SlackDeliveryOperationKind,
	payload map[string]any,
) (*coredata.SlackDeliveryOperation, bool, error) {
	q.keys = append(q.keys, operationKey)
	q.kinds = append(q.kinds, kind)
	q.payloads = append(q.payloads, payload)

	return &coredata.SlackDeliveryOperation{}, true, nil
}

func TestTools_OmitChannelFromLLMParams(t *testing.T) {
	t.Parallel()

	for _, tool := range slackchannel.Tools(nil) {
		t.Run(
			tool.Name(),
			func(t *testing.T) {
				t.Parallel()

				var schema struct {
					Properties map[string]json.RawMessage `json:"properties"`
				}

				err := json.Unmarshal(tool.Definition().Parameters, &schema)
				require.NoError(t, err)

				_, hasChannel := schema.Properties["channel"]
				assert.False(t, hasChannel, "channel must come from RunContext, not LLM params")

				_, hasThreadTS := schema.Properties["thread_ts"]
				assert.False(t, hasThreadTS, "thread_ts must come from RunContext, not LLM params")
			},
		)
	}
}

func TestTools_ExposeGenericSendMessageAndSlackReaction(t *testing.T) {
	t.Parallel()

	tools := slackchannel.Tools(nil)
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name())
	}

	assert.True(t, slices.Contains(names, "send_message"))
	assert.True(t, slices.Contains(names, "add_reaction"))
	assert.False(t, slices.Contains(names, "post_message"))
}

func TestTools_SendMessageRequiresToolCallID(t *testing.T) {
	t.Parallel()

	queue := &recordingToolQueue{}
	tool := sendMessageTool(t, queue)
	ctx := agent.WithRunContext(
		t.Context(),
		trustedSlackRunContext(),
	)

	result, err := tool.Execute(ctx, `{"text":"hello"}`)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "cannot queue message without a stable tool call ID")
	assert.Empty(t, queue.keys)
}

func TestTools_SendMessageOperationKeyIncludesToolCallID(t *testing.T) {
	t.Parallel()

	queue := &recordingToolQueue{}
	tool := sendMessageTool(t, queue)
	rc := trustedSlackRunContext()
	ctx := agent.WithToolCallID(
		agent.WithRunContext(t.Context(), rc),
		"call-1",
	)

	result, err := tool.Execute(ctx, `{"text":"hello"}`)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	require.Len(t, queue.keys, 1)
	assert.Equal(t, coredata.SlackDeliveryOperationKindPostMessage, queue.kinds[0])
	assert.Equal(t, rc.MessageAnchor.ConversationID, queue.payloads[0]["channel"])

	ctx = agent.WithToolCallID(
		agent.WithRunContext(t.Context(), rc),
		"call-1",
	)
	_, err = tool.Execute(ctx, `{"text":"hello again"}`)
	require.NoError(t, err)
	require.Len(t, queue.keys, 2)
	assert.Equal(t, queue.keys[0], queue.keys[1])

	ctx = agent.WithToolCallID(
		agent.WithRunContext(t.Context(), rc),
		"call-2",
	)
	_, err = tool.Execute(ctx, `{"text":"other call"}`)
	require.NoError(t, err)
	require.Len(t, queue.keys, 3)
	assert.NotEqual(t, queue.keys[0], queue.keys[2])
}

func sendMessageTool(t *testing.T, queue *recordingToolQueue) agent.Tool {
	t.Helper()

	for _, tool := range slackchannel.Tools(queue) {
		if tool.Name() == "send_message" {
			return tool
		}
	}

	t.Fatal("send_message tool is missing")

	return nil
}

func trustedSlackRunContext() *slackchannel.RunContext {
	tenantID := gid.NewTenantID()

	return &slackchannel.RunContext{
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		MessageAnchor: probot.MessageAnchor{
			ConversationID: "C123",
			MessageID:      "111.222",
		},
		CurrentMessageID: "111.222",
	}
}
