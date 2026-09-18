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

package console_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const e2eLinearWebhookSecret = "e2e-linear-webhook-secret"

func TestTaskLinearSync_UpdateEnqueuesOutboundAndInboundDoesNotLoop(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).
		WithName("Linear sync source").
		Create()

	externalID := "linear-issue-" + factory.SafeName("ext")
	seedTaskLinearLink(t, owner, taskID, externalID)

	const linkQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					externalLink {
						provider
						identifier
						url
						origin
					}
				}
			}
		}
	`

	var linkResult struct {
		Node *struct {
			ExternalLink *struct {
				Provider   string `json:"provider"`
				Identifier string `json:"identifier"`
				Origin     string `json:"origin"`
			} `json:"externalLink"`
		} `json:"node"`
	}

	err := owner.Execute(linkQuery, map[string]any{"id": taskID}, &linkResult)
	require.NoError(t, err)
	require.NotNil(t, linkResult.Node)
	require.NotNil(t, linkResult.Node.ExternalLink)
	assert.Equal(t, "LINEAR", linkResult.Node.ExternalLink.Provider)
	assert.Equal(t, "ENG-1", linkResult.Node.ExternalLink.Identifier)
	assert.Equal(t, "PROBO", linkResult.Node.ExternalLink.Origin)

	const updateQuery = `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
					name
				}
			}
		}
	`

	var updateResult struct {
		UpdateTask struct {
			Task struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"task"`
		} `json:"updateTask"`
	}

	err = owner.Execute(updateQuery, map[string]any{
		"input": map[string]any{
			"taskId": taskID,
			"name":   "Linear sync updated",
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Equal(t, "Linear sync updated", updateResult.UpdateTask.Task.Name)

	require.Eventually(t, func() bool {
		return countTaskSyncJobs(t, taskID) >= 1
	}, 5*time.Second, 100*time.Millisecond)

	jobsAfterUpdate := countTaskSyncJobs(t, taskID)
	require.GreaterOrEqual(t, jobsAfterUpdate, 1)

	deliveryID := factory.SafeName("delivery")
	status := postLinearWebhook(t, deliveryID, map[string]any{
		"action":           "update",
		"type":             "Issue",
		"webhookTimestamp": time.Now().UnixMilli(),
		"actor":            map[string]any{"id": "linear-user-1"},
		"data": map[string]any{
			"id":          externalID,
			"identifier":  "ENG-99",
			"title":       "Inbound Linear title",
			"description": "Inbound body",
			"priority":    1,
			"dueDate":     "2026-10-01",
			"updatedAt":   time.Now().Add(time.Minute).Format(time.RFC3339),
			"url":         "https://linear.app/eng/issue/ENG-99",
			"state":       map[string]any{"type": "started"},
		},
	})
	require.Equal(t, http.StatusOK, status)

	require.Eventually(t, func() bool {
		return taskName(t, owner, taskID) == "Inbound Linear title"
	}, 30*time.Second, 200*time.Millisecond)

	assert.Equal(t, jobsAfterUpdate, countTaskSyncJobs(t, taskID))

	duplicateStatus := postLinearWebhook(t, deliveryID, map[string]any{
		"action":           "update",
		"type":             "Issue",
		"webhookTimestamp": time.Now().UnixMilli(),
		"actor":            map[string]any{"id": "linear-user-1"},
		"data": map[string]any{
			"id":        externalID,
			"title":     "Should not reprocess",
			"updatedAt": time.Now().Add(2 * time.Minute).Format(time.RFC3339),
			"state":     map[string]any{"type": "started"},
		},
	})
	require.Equal(t, http.StatusOK, duplicateStatus)

	require.Never(t, func() bool {
		return taskName(t, owner, taskID) != "Inbound Linear title"
	}, 3*time.Second, 200*time.Millisecond)

	assert.Equal(t, jobsAfterUpdate, countTaskSyncJobs(t, taskID))
}

func TestTaskLinearSync_UnlinkedCreateIsIgnored(t *testing.T) {
	t.Parallel()

	status := postLinearWebhook(t, factory.SafeName("unlinked"), map[string]any{
		"action":           "create",
		"type":             "Issue",
		"webhookTimestamp": time.Now().UnixMilli(),
		"actor":            map[string]any{"id": "linear-user-1"},
		"data": map[string]any{
			"id":        "unlinked-issue",
			"title":     "Should be ignored",
			"updatedAt": time.Now().Format(time.RFC3339),
			"state":     map[string]any{"type": "unstarted"},
		},
	})
	require.Equal(t, http.StatusOK, status)
}

func seedTaskLinearLink(t *testing.T, owner *testutil.Client, taskID, externalID string) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	parsedTaskID, err := gid.ParseGID(taskID)
	require.NoError(t, err)

	orgID := owner.GetOrganizationID()
	tenantID := orgID.TenantID()
	now := time.Now().UTC()

	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)

	_, err = conn.Exec(ctx, `
		INSERT INTO connectors (
			id, tenant_id, organization_id, provider, protocol, settings, encrypted_connection, created_at, updated_at
		) VALUES (
			$1, $2, $3, 'LINEAR', 'OAUTH2', '{}', '\x', $4, $4
		)
	`, connectorID, tenantID, orgID, now)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `
		INSERT INTO task_external_links (
			task_id, tenant_id, organization_id, connector_id, provider,
			external_id, external_identifier, external_url, destination, origin,
			remote_updated_at, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 'LINEAR',
			$5, 'ENG-1', 'https://linear.app/eng/issue/ENG-1', '{"team_id":"team-1"}', 'PROBO',
			$6, '{"app_actor_id":"probo-app"}', $7, $7
		)
	`, parsedTaskID, tenantID, orgID, connectorID, externalID, now.Add(-time.Hour), now)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)
		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, _ = cleanupConn.Exec(cleanupCtx, `DELETE FROM task_external_links WHERE task_id = $1`, parsedTaskID)
		_, _ = cleanupConn.Exec(cleanupCtx, `DELETE FROM connectors WHERE id = $1`, connectorID)
	})
}

func countTaskSyncJobs(t *testing.T, taskID string) int {
	t.Helper()

	ctx := context.Background()

	conn := dialTestPg(t, ctx)
	defer func() { _ = conn.Close(ctx) }()

	var count int

	err := conn.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM task_sync_jobs
		WHERE payload->>'task_id' = $1
			AND direction = 'OUTBOUND'
	`, taskID).Scan(&count)
	require.NoError(t, err)

	return count
}

func taskName(t *testing.T, owner *testutil.Client, taskID string) string {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					name
				}
			}
		}
	`

	var result struct {
		Node *struct {
			Name string `json:"name"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": taskID}, &result)
	require.NoError(t, err)

	if result.Node == nil {
		return ""
	}

	return result.Node.Name
}

func postLinearWebhook(t *testing.T, deliveryID string, payload map[string]any) int {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte(e2eLinearWebhookSecret))
	_, _ = mac.Write(body)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/linear/v1/webhooks",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Linear-Signature", hex.EncodeToString(mac.Sum(nil)))
	req.Header.Set("Linear-Delivery", deliveryID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func TestTask_LinearTeamsWhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					linearTeams {
						id
						name
						key
					}
				}
			}
		}
	`

	var result struct {
		Node *struct {
			LinearTeams []struct {
				ID string `json:"id"`
			} `json:"linearTeams"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": owner.GetOrganizationID().String()}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node)
	assert.Empty(t, result.Node.LinearTeams)
}

func TestTask_UnlinkExternalWhenNotLinked(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	const mutation = `
		mutation($input: UnlinkTaskExternalInput!) {
			unlinkTaskExternal(input: $input) {
				task { id }
			}
		}
	`

	err := owner.ExecuteShouldFail(mutation, map[string]any{
		"input": map[string]any{"taskId": taskID},
	})
	require.Error(t, err)
}

func TestTaskLinearSync_InvalidSignatureRejected(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"update","type":"Issue","webhookTimestamp":1}`)
	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/linear/v1/webhooks",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("Linear-Signature", "deadbeef")
	req.Header.Set("Linear-Delivery", factory.SafeName("bad-sig"))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
