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

package trust

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

func TestRightsRequestService_CreateEnqueuesWebhook(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	organizationID := insertPortalRequestWebhookOrganization(t, client)
	scope := coredata.NewScope(organizationID.TenantID())
	insertPortalRequestWebhookSubscription(t, client, scope, organizationID)

	service := RightsRequestService{svc: &Service{pg: client}}
	dataSubject := "Jane Doe"
	contact := "jane@example.com"
	rightsRequest, err := service.Create(
		t.Context(),
		scope,
		&CreateRightsRequest{
			OrganizationID: organizationID,
			RequestType:    coredata.RightsRequestTypeAccess,
			DataSubject:    &dataSubject,
			Contact:        contact,
		},
	)
	require.NoError(t, err)

	var data []byte

	err = client.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				`SELECT data FROM webhook_data WHERE organization_id = $1 AND event_type = $2`,
				organizationID.String(),
				coredata.WebhookEventTypeRightRequestCreated.String(),
			).Scan(&data)
		},
	)
	require.NoError(t, err)

	payload := &webhooktypes.RightsRequest{}
	require.NoError(t, json.Unmarshal(data, payload))
	assert.Equal(t, rightsRequest.ID, payload.ID)
	assert.Equal(t, coredata.RightsRequestStateTodo, payload.RequestState)
	assert.Equal(t, dataSubject, *payload.DataSubject)
	assert.Equal(t, contact, *payload.Contact)
	assert.NotNil(t, payload.Deadline)
}

func insertPortalRequestWebhookOrganization(t *testing.T, client *pg.Client) gid.GID {
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
				"portal-request-webhook-"+organizationID.String(),
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

func insertPortalRequestWebhookSubscription(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
) {
	t.Helper()

	now := time.Now()
	subscription := coredata.WebhookSubscription{
		ID:                     gid.New(organizationID.TenantID(), coredata.WebhookSubscriptionEntityType),
		OrganizationID:         organizationID,
		EndpointURL:            "https://example.test/webhook",
		SelectedEvents:         coredata.WebhookEventTypes{coredata.WebhookEventTypeRightRequestCreated},
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
