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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestWebhook_SubscriptionLifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	initialEndpoint := unroutableWebhookEndpoint(t)
	beforeCreate := time.Now().Add(-time.Second)

	created := createWebhookSubscription(
		t,
		owner,
		initialEndpoint,
		thirdPartyWebhookEventTypes(),
	)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, initialEndpoint, created.EndpointURL)
	assert.Equal(t, thirdPartyWebhookEventTypes(), created.SelectedEvents)
	testutil.AssertTimestampsOnCreate(t, created.CreatedAt, created.UpdatedAt, beforeCreate)

	var nodeAfterCreate webhookSubscriptionNodeResponse

	err := owner.Execute(
		webhookSubscriptionNodeQuery,
		map[string]any{"id": created.ID},
		&nodeAfterCreate,
	)
	require.NoError(t, err)
	require.NotNil(t, nodeAfterCreate.Node)
	assert.Equal(t, created.ID, nodeAfterCreate.Node.ID)
	assert.NotEmpty(t, nodeAfterCreate.Node.SigningSecret)
	assert.Equal(t, orgID, nodeAfterCreate.Node.Organization.ID)
	assert.True(t, nodeAfterCreate.Node.Permission)

	var orgList organizationWebhookSubscriptionsResponse

	err = owner.Execute(
		organizationWebhookSubscriptionsQuery,
		map[string]any{"id": orgID},
		&orgList,
	)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, orgList.Node.WebhookSubscriptions.TotalCount, 1)
	assert.True(t, organizationContainsWebhookSubscription(t, owner, orgID, created.ID))

	updatedEndpoint := unroutableWebhookEndpoint(t)
	updatedEvents := []string{"THIRD_PARTY_CREATED", "THIRD_PARTY_UPDATED"}

	var updateResult struct {
		UpdateWebhookSubscription struct {
			WebhookSubscription webhookSubscriptionNode `json:"webhookSubscription"`
		} `json:"updateWebhookSubscription"`
	}

	err = owner.Execute(
		updateWebhookSubscriptionMutation,
		map[string]any{
			"input": map[string]any{
				"id":             created.ID,
				"endpointUrl":    updatedEndpoint,
				"selectedEvents": updatedEvents,
			},
		},
		&updateResult,
	)
	require.NoError(t, err)

	updated := updateResult.UpdateWebhookSubscription.WebhookSubscription
	assert.Equal(t, updatedEndpoint, updated.EndpointURL)
	assert.Equal(t, updatedEvents, updated.SelectedEvents)
	testutil.AssertTimestampsOnUpdate(
		t,
		updated.CreatedAt,
		updated.UpdatedAt,
		created.CreatedAt,
		created.UpdatedAt,
	)

	var deleteResult struct {
		DeleteWebhookSubscription struct {
			DeletedWebhookSubscriptionID string `json:"deletedWebhookSubscriptionId"`
		} `json:"deleteWebhookSubscription"`
	}

	err = owner.Execute(
		deleteWebhookSubscriptionMutation,
		map[string]any{
			"input": map[string]any{
				"webhookSubscriptionId": created.ID,
			},
		},
		&deleteResult,
	)
	require.NoError(t, err)
	assert.Equal(t, created.ID, deleteResult.DeleteWebhookSubscription.DeletedWebhookSubscriptionID)

	_, err = owner.Do(
		webhookSubscriptionNodeQuery,
		map[string]any{"id": created.ID},
	)

	var gqlErrors testutil.GraphQLErrors
	require.ErrorAs(t, err, &gqlErrors)
	require.NotEmpty(t, gqlErrors)
	assert.Equal(t, "NOT_FOUND", gqlErrors[0].Code())

	assert.False(
		t,
		organizationContainsWebhookSubscription(t, owner, orgID, created.ID),
		"deleted subscription must not appear in organization connection",
	)
}

func TestWebhook_Create_Validation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("rejects insecure HTTP endpoint", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(
			createWebhookSubscriptionMutation,
			map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"endpointUrl":    "http://example.com/webhook",
					"selectedEvents": thirdPartyWebhookEventTypes(),
				},
			},
		)
		testutil.RequireErrorCode(t, err, "INVALID", "HTTP endpoint must be rejected")
	})
}

func TestWebhook_ThirdPartyCreatedEvent(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	endpoint := unroutableWebhookEndpoint(t)

	subscription := createWebhookSubscription(
		t,
		owner,
		endpoint,
		[]string{"THIRD_PARTY_CREATED"},
	)

	factory.NewThirdParty(owner).
		WithName(factory.SafeName("Webhook Third Party")).
		Create()

	eventResult := requireWebhookEventsEventually(t, owner, subscription.ID, 1)
	require.NotNil(t, eventResult.Node)
	assert.Equal(t, 1, eventResult.Node.Events.TotalCount)
	require.NotEmpty(t, eventResult.Node.Events.Edges)

	event := eventResult.Node.Events.Edges[0].Node
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, subscription.ID, event.WebhookSubscriptionID)
	assert.Equal(t, "PENDING", event.Status)
	assert.False(t, event.CreatedAt.IsZero())
}
