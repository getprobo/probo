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

package webhook_test

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
	"go.probo.inc/probo/pkg/webhook"
)

func TestInsertUpdateData_PersistsUpdatedFromSnapshot(t *testing.T) {
	client := test.PGClient(t)
	orgID := insertTestOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())

	insertTestSubscription(t, client, orgID, coredata.WebhookEventTypeUserUpdated)

	current := map[string]any{"role": "OWNER"}
	updatedFrom := map[string]any{"role": "ADMIN"}

	err := client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
		return webhook.InsertUpdateData(
			ctx,
			tx,
			scope,
			orgID,
			coredata.WebhookEventTypeUserUpdated,
			current,
			updatedFrom,
		)
	})
	require.NoError(t, err)

	data, updatedFromData := loadWebhookData(t, client, orgID)

	assert.JSONEq(t, `{"role":"OWNER"}`, string(data))
	require.NotNil(t, updatedFromData, "updated_from must be persisted for update events")
	assert.JSONEq(t, `{"role":"ADMIN"}`, string(updatedFromData))
}

func TestInsertData_StoresNullUpdatedFrom(t *testing.T) {
	client := test.PGClient(t)
	orgID := insertTestOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())

	insertTestSubscription(t, client, orgID, coredata.WebhookEventTypeObligationUpdated)

	err := client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
		return webhook.InsertData(
			ctx,
			tx,
			scope,
			orgID,
			coredata.WebhookEventTypeObligationUpdated,
			map[string]any{"status": "OPEN"},
		)
	})
	require.NoError(t, err)

	data, updatedFromData := loadWebhookData(t, client, orgID)

	assert.JSONEq(t, `{"status":"OPEN"}`, string(data))
	assert.Nil(t, updatedFromData, "updated_from must be SQL NULL when no previous snapshot is provided")
}

func TestInsertUpdateData_NoSubscriptionIsNoop(t *testing.T) {
	client := test.PGClient(t)
	orgID := insertTestOrganization(t, client)
	scope := coredata.NewScope(orgID.TenantID())

	// Subscribe to a different event than the one we emit.
	insertTestSubscription(t, client, orgID, coredata.WebhookEventTypeObligationUpdated)

	err := client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
		return webhook.InsertUpdateData(
			ctx,
			tx,
			scope,
			orgID,
			coredata.WebhookEventTypeUserUpdated,
			map[string]any{"role": "OWNER"},
			map[string]any{"role": "ADMIN"},
		)
	})
	require.NoError(t, err)

	assert.Equal(t, 0, countWebhookData(t, client, orgID), "no webhook_data row should be enqueued without a matching subscription")
}

func TestPayload_UpdatedFromOmittedWhenAbsent(t *testing.T) {
	withUpdatedFrom, err := json.Marshal(webhook.Payload{
		EventType:   "user:updated",
		Data:        json.RawMessage(`{"role":"OWNER"}`),
		UpdatedFrom: json.RawMessage(`{"role":"ADMIN"}`),
	})
	require.NoError(t, err)
	assert.Contains(t, string(withUpdatedFrom), `"updatedFrom":{"role":"ADMIN"}`)

	withoutUpdatedFrom, err := json.Marshal(webhook.Payload{
		EventType: "user:created",
		Data:      json.RawMessage(`{"role":"OWNER"}`),
	})
	require.NoError(t, err)
	assert.NotContains(t, string(withoutUpdatedFrom), "updatedFrom")
}

func insertTestOrganization(t *testing.T, client *pg.Client) gid.GID {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				orgID.String(),
				tenantID.String(),
				"test-org-"+orgID.String(),
				now,
				now,
			)

			return err
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", orgID.String())
			return err
		})
	})

	return orgID
}

func insertTestSubscription(
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
		EndpointURL:            "https://example.test/webhook",
		SelectedEvents:         coredata.WebhookEventTypes(events),
		EncryptedSigningSecret: []byte("test-signing-secret"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err := client.WithTx(
		context.Background(),
		func(ctx context.Context, tx pg.Tx) error {
			return subscription.Insert(ctx, tx, coredata.NewScope(orgID.TenantID()))
		},
	)
	require.NoError(t, err)
}

func loadWebhookData(t *testing.T, client *pg.Client, orgID gid.GID) (json.RawMessage, json.RawMessage) {
	t.Helper()

	var (
		data        []byte
		updatedFrom []byte
	)

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				"SELECT data, updated_from FROM webhook_data WHERE organization_id = $1",
				orgID.String(),
			).Scan(&data, &updatedFrom)
		},
	)
	require.NoError(t, err)

	return data, updatedFrom
}

func countWebhookData(t *testing.T, client *pg.Client, orgID gid.GID) int {
	t.Helper()

	var count int

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				"SELECT COUNT(*) FROM webhook_data WHERE organization_id = $1",
				orgID.String(),
			).Scan(&count)
		},
	)
	require.NoError(t, err)

	return count
}
