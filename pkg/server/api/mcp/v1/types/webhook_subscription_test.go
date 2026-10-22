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

package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func TestWebhookSubscriptionSchemas_AcceptDocumentEvents(t *testing.T) {
	t.Parallel()

	events := documentWebhookEventTypes()
	now := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	tenantID := gid.NewTenantID()
	subscriptionID := gid.New(tenantID, coredata.WebhookSubscriptionEntityType)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	subscription := &types.WebhookSubscription{
		ID:             subscriptionID,
		OrganizationID: organizationID,
		EndpointURL:    "https://example.com/hooks",
		SelectedEvents: events,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	outputSchemas := []struct {
		name   string
		schema *jsonschema.Schema
		value  any
	}{
		{
			name:   "list subscriptions",
			schema: types.ListWebhookSubscriptionsToolOutputSchema,
			value: types.ListWebhookSubscriptionsOutput{
				WebhookSubscriptions: []*types.WebhookSubscription{subscription},
			},
		},
		{
			name:   "get subscription",
			schema: types.GetWebhookSubscriptionToolOutputSchema,
			value: types.GetWebhookSubscriptionOutput{
				WebhookSubscription: subscription,
			},
		},
		{
			name:   "create subscription output",
			schema: types.CreateWebhookSubscriptionToolOutputSchema,
			value: types.CreateWebhookSubscriptionOutput{
				WebhookSubscription: subscription,
			},
		},
		{
			name:   "update subscription output",
			schema: types.UpdateWebhookSubscriptionToolOutputSchema,
			value: types.UpdateWebhookSubscriptionOutput{
				WebhookSubscription: subscription,
			},
		},
		{
			name:   "create subscription input",
			schema: types.CreateWebhookSubscriptionToolInputSchema,
			value: types.CreateWebhookSubscriptionInput{
				OrganizationID: organizationID,
				EndpointURL:    "https://example.com/hooks",
				SelectedEvents: events,
			},
		},
		{
			name:   "update subscription input",
			schema: types.UpdateWebhookSubscriptionToolInputSchema,
			value: types.UpdateWebhookSubscriptionInput{
				ID:             subscriptionID,
				SelectedEvents: events,
			},
		},
	}

	for _, tc := range outputSchemas {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				enum := selectedEventEnum(t, tc.schema)
				assert.Contains(t, enum, string(coredata.WebhookEventTypeDocumentArchived))

				for _, event := range events {
					assert.Contains(t, enum, string(event))
				}

				requireValidSchema(t, tc.schema, tc.value)
			},
		)
	}
}

func documentWebhookEventTypes() []coredata.WebhookEventType {
	return []coredata.WebhookEventType{
		coredata.WebhookEventTypeDocumentCreated,
		coredata.WebhookEventTypeDocumentUpdated,
		coredata.WebhookEventTypeDocumentArchived,
		coredata.WebhookEventTypeDocumentUnarchived,
		coredata.WebhookEventTypeDocumentDeleted,
		coredata.WebhookEventTypeDocumentVersionCreated,
		coredata.WebhookEventTypeDocumentVersionUpdated,
		coredata.WebhookEventTypeDocumentVersionPublished,
		coredata.WebhookEventTypeDocumentVersionRejected,
		coredata.WebhookEventTypeDocumentVersionDeleted,
		coredata.WebhookEventTypeDocumentVersionSignatureRequested,
		coredata.WebhookEventTypeDocumentVersionSignatureSigned,
		coredata.WebhookEventTypeDocumentVersionSignatureCancelled,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumRequested,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumUpdated,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumApproved,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumRejected,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumVoided,
	}
}

func selectedEventEnum(t *testing.T, schema *jsonschema.Schema) []any {
	t.Helper()

	selected := findSchemaProperty(schema, "selected_events")
	require.NotNil(t, selected)
	require.NotNil(t, selected.Items)

	return selected.Items.Enum
}

func findSchemaProperty(schema *jsonschema.Schema, name string) *jsonschema.Schema {
	if schema == nil {
		return nil
	}

	if property, ok := schema.Properties[name]; ok {
		return property
	}

	for _, child := range schema.Properties {
		if found := findSchemaProperty(child, name); found != nil {
			return found
		}
	}

	return findSchemaProperty(schema.Items, name)
}

func requireValidSchema(t *testing.T, schema *jsonschema.Schema, value any) {
	t.Helper()

	raw, err := json.Marshal(value)
	require.NoError(t, err)

	var instance any
	require.NoError(t, json.Unmarshal(raw, &instance))

	resolved, err := schema.Resolve(nil)
	require.NoError(t, err)
	require.NoError(t, resolved.Validate(instance))
}
