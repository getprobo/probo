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

package probo

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

func TestRightsRequestService_WebhookLifecycle(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	organizationID := insertRightsRequestWebhookOrganization(t, client)
	scope := coredata.NewScope(organizationID.TenantID())
	insertRightsRequestWebhookSubscription(t, client, scope, organizationID)

	service := RightsRequestService{svc: &Service{pg: client}}
	requestType := coredata.RightsRequestTypeAccess
	requestState := coredata.RightsRequestStateTodo
	dataSubject := "Jane Doe"
	contact := "jane@example.com"

	rightsRequest, err := service.Create(
		t.Context(),
		scope,
		&CreateRightsRequestRequest{
			OrganizationID: organizationID,
			RequestType:    &requestType,
			RequestState:   &requestState,
			DataSubject:    &dataSubject,
			Contact:        &contact,
		},
	)
	require.NoError(t, err)

	created, createdFrom := loadRightsRequestWebhookData(
		t,
		client,
		organizationID,
		coredata.WebhookEventTypeRightRequestCreated,
	)
	assert.Equal(t, rightsRequest.ID, created.ID)
	assert.Equal(t, dataSubject, *created.DataSubject)
	assert.Nil(t, createdFrom)

	updatedState := coredata.RightsRequestStateInProgress
	updatedSubject := "Jane Smith"
	rightsRequest, err = service.Update(
		t.Context(),
		scope,
		&UpdateRightsRequestRequest{
			ID:           rightsRequest.ID,
			RequestState: &updatedState,
			DataSubject:  new(&updatedSubject),
		},
	)
	require.NoError(t, err)

	updated, updatedFrom := loadRightsRequestWebhookData(
		t,
		client,
		organizationID,
		coredata.WebhookEventTypeRightRequestUpdated,
	)
	require.NotNil(t, updatedFrom)
	assert.Equal(t, updatedState, updated.RequestState)
	assert.Equal(t, updatedSubject, *updated.DataSubject)
	assert.Equal(t, requestState, updatedFrom.RequestState)
	assert.Equal(t, dataSubject, *updatedFrom.DataSubject)

	err = service.Delete(t.Context(), scope, rightsRequest.ID)
	require.NoError(t, err)

	deleted, deletedFrom := loadRightsRequestWebhookData(
		t,
		client,
		organizationID,
		coredata.WebhookEventTypeRightRequestDeleted,
	)
	assert.Equal(t, rightsRequest.ID, deleted.ID)
	assert.Equal(t, updatedState, deleted.RequestState)
	assert.Nil(t, deletedFrom)
}

func insertRightsRequestWebhookOrganization(t *testing.T, client *pg.Client) gid.GID {
	t.Helper()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	err := client.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				organizationID.String(),
				tenantID.String(),
				"right-request-webhook-"+organizationID.String(),
				now,
				now,
			)

			return err
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		_ = client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", organizationID.String())
				return err
			},
		)
	})

	return organizationID
}

func insertRightsRequestWebhookSubscription(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
) {
	t.Helper()

	now := time.Now()
	subscription := coredata.WebhookSubscription{
		ID:             gid.New(organizationID.TenantID(), coredata.WebhookSubscriptionEntityType),
		OrganizationID: organizationID,
		EndpointURL:    "https://example.test/webhook",
		SelectedEvents: coredata.WebhookEventTypes{
			coredata.WebhookEventTypeRightRequestCreated,
			coredata.WebhookEventTypeRightRequestUpdated,
			coredata.WebhookEventTypeRightRequestDeleted,
		},
		EncryptedSigningSecret: []byte("test-signing-secret"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err := client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return subscription.Insert(ctx, tx, scope)
		},
	)
	require.NoError(t, err)
}

func loadRightsRequestWebhookData(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
	eventType coredata.WebhookEventType,
) (*webhooktypes.RightsRequest, *webhooktypes.RightsRequest) {
	t.Helper()

	var (
		data        []byte
		updatedFrom []byte
	)

	err := client.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				`SELECT data, updated_from FROM webhook_data WHERE organization_id = $1 AND event_type = $2`,
				organizationID.String(),
				eventType.String(),
			).Scan(&data, &updatedFrom)
		},
	)
	require.NoError(t, err)

	current := &webhooktypes.RightsRequest{}
	require.NoError(t, json.Unmarshal(data, current))

	if updatedFrom == nil {
		return current, nil
	}

	previous := &webhooktypes.RightsRequest{}
	require.NoError(t, json.Unmarshal(updatedFrom, previous))

	return current, previous
}
