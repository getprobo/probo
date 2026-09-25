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

package task

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

func TestEmitTaskCreated_InsertsCreatedWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	orgID := insertTaskWebhookOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())
	insertTaskWebhookSubscription(t, client, orgID, coredata.WebhookEventTypeTaskCreated)

	task := newWebhookTestTask(orgID, coredata.TaskStateTodo)

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return emitTaskCreated(ctx, tx, scope, task)
		},
	)
	require.NoError(t, err)

	payloads := loadTaskWebhookPayloads(t, client, orgID, coredata.WebhookEventTypeTaskCreated)
	require.Len(t, payloads, 1)
	assert.Equal(t, task.ID.String(), payloads[0]["id"])
	assert.Equal(t, "TODO", payloads[0]["state"])
	assert.Equal(t, float64(1), payloads[0]["rank"])
}

func TestEmitTaskUpdated_InsertsUpdatedWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	orgID := insertTaskWebhookOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())
	insertTaskWebhookSubscription(t, client, orgID, coredata.WebhookEventTypeTaskUpdated)

	previous := newWebhookTestTask(orgID, coredata.TaskStateTodo)
	previous.Rank = 3
	current := *previous
	current.State = coredata.TaskStateDone
	current.Rank = 1
	current.UpdatedAt = previous.UpdatedAt.Add(time.Second)

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return emitTaskUpdated(ctx, tx, scope, previous, &current)
		},
	)
	require.NoError(t, err)

	payloads := loadTaskWebhookPayloads(t, client, orgID, coredata.WebhookEventTypeTaskUpdated)
	require.Len(t, payloads, 1)
	assert.Equal(t, "DONE", payloads[0]["state"])
	assert.Equal(t, float64(1), payloads[0]["rank"])

	updatedFrom := loadTaskWebhookUpdatedFrom(t, client, orgID, coredata.WebhookEventTypeTaskUpdated)
	assert.Equal(t, "TODO", updatedFrom["state"])
	assert.Equal(t, float64(3), updatedFrom["rank"])
}

func TestEmitTaskDeleted_InsertsDeletedWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	orgID := insertTaskWebhookOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())
	insertTaskWebhookSubscription(t, client, orgID, coredata.WebhookEventTypeTaskDeleted)

	task := newWebhookTestTask(orgID, coredata.TaskStateTodo)

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return emitTaskDeleted(ctx, tx, scope, task)
		},
	)
	require.NoError(t, err)

	payloads := loadTaskWebhookPayloads(t, client, orgID, coredata.WebhookEventTypeTaskDeleted)
	require.Len(t, payloads, 1)
	assert.Equal(t, task.ID.String(), payloads[0]["id"])
}

func TestEmitTaskCommentCreated_InsertsCommentWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	orgID := insertTaskWebhookOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())
	insertTaskWebhookSubscription(t, client, orgID, coredata.WebhookEventTypeTaskCommentCreated)

	comment := &coredata.TaskComment{
		ID:             gid.New(orgID.TenantID(), coredata.TaskCommentEntityType),
		OrganizationID: orgID,
		TaskID:         gid.New(orgID.TenantID(), coredata.TaskEntityType),
		OwnerID:        gid.New(orgID.TenantID(), coredata.MembershipProfileEntityType),
		Content:        `{"type":"doc","content":[]}`,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return emitTaskCommentCreated(ctx, tx, scope, comment)
		},
	)
	require.NoError(t, err)

	payloads := loadTaskWebhookPayloads(t, client, orgID, coredata.WebhookEventTypeTaskCommentCreated)
	require.Len(t, payloads, 1)
	assert.Equal(t, comment.ID.String(), payloads[0]["id"])
	assert.Equal(t, comment.TaskID.String(), payloads[0]["taskId"])

	expected, err := json.Marshal(webhooktypes.NewTaskComment(comment))
	require.NoError(t, err)
	actual, err := json.Marshal(payloads[0])
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), string(actual))
}

func TestEmitTaskCommentUpdated_InsertsUpdatedWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	orgID := insertTaskWebhookOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())
	insertTaskWebhookSubscription(t, client, orgID, coredata.WebhookEventTypeTaskCommentUpdated)

	previous := newWebhookTestComment(orgID)
	comment := *previous
	comment.Content = `{"type":"doc","content":[{"type":"paragraph"}]}`
	comment.UpdatedAt = previous.UpdatedAt.Add(time.Second)

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return emitTaskCommentUpdated(ctx, tx, scope, previous, &comment)
		},
	)
	require.NoError(t, err)

	payloads := loadTaskWebhookPayloads(t, client, orgID, coredata.WebhookEventTypeTaskCommentUpdated)
	require.Len(t, payloads, 1)
	assert.Equal(t, comment.Content, payloads[0]["content"])

	updatedFrom := loadTaskWebhookUpdatedFrom(t, client, orgID, coredata.WebhookEventTypeTaskCommentUpdated)
	assert.Equal(t, previous.Content, updatedFrom["content"])
}

func TestEmitTaskCommentDeleted_InsertsDeletedWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	orgID := insertTaskWebhookOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())
	insertTaskWebhookSubscription(t, client, orgID, coredata.WebhookEventTypeTaskCommentDeleted)

	comment := newWebhookTestComment(orgID)

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return emitTaskCommentDeleted(ctx, tx, scope, comment)
		},
	)
	require.NoError(t, err)

	payloads := loadTaskWebhookPayloads(t, client, orgID, coredata.WebhookEventTypeTaskCommentDeleted)
	require.Len(t, payloads, 1)
	assert.Equal(t, comment.ID.String(), payloads[0]["id"])
}

func newWebhookTestTask(orgID gid.GID, state coredata.TaskState) *coredata.Task {
	now := time.Now()

	return &coredata.Task{
		ID:             gid.New(orgID.TenantID(), coredata.TaskEntityType),
		OrganizationID: orgID,
		Name:           "webhook task",
		Content:        `{"type":"doc","content":[]}`,
		State:          state,
		Priority:       coredata.TaskPriorityMedium,
		Rank:           1,
		ReferenceID:    "custom-task-test",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func newWebhookTestComment(orgID gid.GID) *coredata.TaskComment {
	now := time.Now()

	return &coredata.TaskComment{
		ID:             gid.New(orgID.TenantID(), coredata.TaskCommentEntityType),
		OrganizationID: orgID,
		TaskID:         gid.New(orgID.TenantID(), coredata.TaskEntityType),
		OwnerID:        gid.New(orgID.TenantID(), coredata.MembershipProfileEntityType),
		Content:        `{"type":"doc","content":[]}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func insertTaskWebhookOrganization(t *testing.T, client *pg.Client) gid.GID {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	err := client.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				orgID.String(),
				tenantID.String(),
				"task-webhook-org-"+orgID.String(),
				now,
				now,
			)

			return err
		},
	)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = client.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", orgID.String())
					return err
				},
			)
		},
	)

	return orgID
}

func insertTaskWebhookSubscription(
	t *testing.T,
	client *pg.Client,
	orgID gid.GID,
	events ...coredata.WebhookEventType,
) {
	t.Helper()

	now := time.Now()
	subscription := coredata.WebhookSubscription{
		ID:                     gid.New(orgID.TenantID(), coredata.WebhookSubscriptionEntityType),
		OrganizationID:         orgID,
		EndpointURL:            "https://example.test/task-webhook",
		SelectedEvents:         coredata.WebhookEventTypes(events),
		EncryptedSigningSecret: []byte("test-signing-secret"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return subscription.Insert(ctx, tx, coredata.NewScope(orgID.TenantID()))
		},
	)
	require.NoError(t, err)
}

func loadTaskWebhookPayloads(
	t *testing.T,
	client *pg.Client,
	orgID gid.GID,
	eventType coredata.WebhookEventType,
) []map[string]any {
	t.Helper()

	var rows []json.RawMessage

	err := client.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			queryRows, err := conn.Query(
				ctx,
				`SELECT data FROM webhook_data WHERE organization_id = $1 AND event_type = $2 ORDER BY created_at`,
				orgID.String(),
				eventType.String(),
			)
			if err != nil {
				return err
			}
			defer queryRows.Close()

			for queryRows.Next() {
				var data json.RawMessage
				if err := queryRows.Scan(&data); err != nil {
					return err
				}

				rows = append(rows, data)
			}

			return queryRows.Err()
		},
	)
	require.NoError(t, err)

	payloads := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var payload map[string]any
		require.NoError(t, json.Unmarshal(row, &payload))
		payloads = append(payloads, payload)
	}

	return payloads
}

func loadTaskWebhookUpdatedFrom(
	t *testing.T,
	client *pg.Client,
	orgID gid.GID,
	eventType coredata.WebhookEventType,
) map[string]any {
	t.Helper()

	var data json.RawMessage

	err := client.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				`SELECT updated_from FROM webhook_data WHERE organization_id = $1 AND event_type = $2`,
				orgID.String(),
				eventType.String(),
			).Scan(&data)
		},
	)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(data, &payload))

	return payload
}
