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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

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

	pgClient, organizationID := toolsDatabase(t)
	queue := slackchannel.NewDeliveryService(pgClient)
	tool := sendMessageTool(t, queue)
	rc := trustedSlackRunContext(organizationID)
	ctx := agent.WithRunContext(t.Context(), rc)

	result, err := tool.Execute(ctx, `{"text":"hello"}`)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "cannot queue message without a stable tool call ID")
	assert.Empty(t, loadToolOperationKeys(t, pgClient, organizationID))
}

func TestTools_SendMessageOperationKeyIncludesToolCallID(t *testing.T) {
	t.Parallel()

	pgClient, organizationID := toolsDatabase(t)
	queue := slackchannel.NewDeliveryService(pgClient)
	tool := sendMessageTool(t, queue)
	rc := trustedSlackRunContext(organizationID)

	result, err := tool.Execute(
		agent.WithToolCallID(agent.WithRunContext(t.Context(), rc), "call-1"),
		`{"text":"hello"}`,
	)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	keys := loadToolOperationKeys(t, pgClient, organizationID)
	require.Len(t, keys, 1)
	assert.Equal(t, rc.MessageAnchor.ConversationID, loadToolPayload(t, pgClient, organizationID, keys[0])["channel"])

	_, err = tool.Execute(
		agent.WithToolCallID(agent.WithRunContext(t.Context(), rc), "call-1"),
		`{"text":"hello again"}`,
	)
	require.NoError(t, err)
	keys = loadToolOperationKeys(t, pgClient, organizationID)
	require.Len(t, keys, 1)

	_, err = tool.Execute(
		agent.WithToolCallID(agent.WithRunContext(t.Context(), rc), "call-2"),
		`{"text":"other call"}`,
	)
	require.NoError(t, err)
	keys = loadToolOperationKeys(t, pgClient, organizationID)
	require.Len(t, keys, 2)
	assert.NotEqual(t, keys[0], keys[1])
}

func sendMessageTool(t *testing.T, queue *slackchannel.DeliveryService) agent.Tool {
	t.Helper()

	for _, tool := range slackchannel.Tools(queue) {
		if tool.Name() == "send_message" {
			return tool
		}
	}

	t.Fatal("send_message tool is missing")

	return nil
}

func trustedSlackRunContext(organizationID gid.GID) *probot.RunContext {
	return &probot.RunContext{
		OrganizationID: organizationID,
		MessageAnchor: probot.MessageAnchor{
			ConversationID: "C123",
			MessageID:      "111.222",
		},
		CurrentMessageID: "111.222",
	}
}

func toolsDatabase(t *testing.T) (*pg.Client, gid.GID) {
	t.Helper()

	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`INSERT INTO organizations
					 (id, tenant_id, name, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5)`,
					organizationID,
					tenantID,
					"slack-tools-"+organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(
						ctx,
						`DELETE FROM organizations WHERE id = $1`,
						organizationID,
					)

					return err
				},
			)
		},
	)

	return pgClient, organizationID
}

func loadToolOperationKeys(
	t *testing.T,
	pgClient *pg.Client,
	organizationID gid.GID,
) []string {
	t.Helper()

	var keys []string
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := conn.Query(
					ctx,
					`SELECT operation_key
					 FROM slack_delivery_operations
					 WHERE organization_id = $1
					 ORDER BY created_at ASC, id ASC`,
					organizationID,
				)
				if err != nil {
					return err
				}
				defer rows.Close()

				for rows.Next() {
					var key string
					if err := rows.Scan(&key); err != nil {
						return err
					}
					keys = append(keys, key)
				}

				return rows.Err()
			},
		),
	)

	return keys
}

func loadToolPayload(
	t *testing.T,
	pgClient *pg.Client,
	organizationID gid.GID,
	operationKey string,
) map[string]any {
	t.Helper()

	var payload map[string]any
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT payload
					 FROM slack_delivery_operations
					 WHERE organization_id = $1 AND operation_key = $2`,
					organizationID,
					operationKey,
				).Scan(&payload)
			},
		),
	)

	return payload
}
