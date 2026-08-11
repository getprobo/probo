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

	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestWebhookSubscription_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	subscriptionID := createWebhookSubscription(
		t,
		org1Owner,
		unroutableWebhookEndpoint(t),
		thirdPartyWebhookEventTypes(),
	).ID

	const readQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on WebhookSubscription {
					id
					endpointUrl
				}
			}
		}
	`

	t.Run("cannot read webhook subscription from another organization", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node *struct {
				ID          string `json:"id"`
				EndpointURL string `json:"endpointUrl"`
			} `json:"node"`
		}

		err := org2Owner.Execute(
			readQuery,
			map[string]any{"id": subscriptionID},
			&result,
		)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "webhook subscription")
	})

	t.Run("cannot update webhook subscription from another organization", func(t *testing.T) {
		t.Parallel()

		_, err := org2Owner.Do(
			updateWebhookSubscriptionMutation,
			map[string]any{
				"input": map[string]any{
					"id":          subscriptionID,
					"endpointUrl": unroutableWebhookEndpoint(t),
				},
			},
		)
		testutil.RequireForbiddenError(
			t,
			err,
			"must not update webhook subscription from another organization",
		)
	})

	t.Run("cannot delete webhook subscription from another organization", func(t *testing.T) {
		t.Parallel()

		_, err := org2Owner.Do(
			deleteWebhookSubscriptionMutation,
			map[string]any{
				"input": map[string]any{
					"webhookSubscriptionId": subscriptionID,
				},
			},
		)
		testutil.RequireForbiddenError(
			t,
			err,
			"must not delete webhook subscription from another organization",
		)
	})

	t.Run("cannot create webhook subscription using another organization id", func(t *testing.T) {
		t.Parallel()

		_, err := org2Owner.Do(
			createWebhookSubscriptionMutation,
			map[string]any{
				"input": map[string]any{
					"organizationId": org1Owner.GetOrganizationID().String(),
					"endpointUrl":    unroutableWebhookEndpoint(t),
					"selectedEvents": thirdPartyWebhookEventTypes(),
				},
			},
		)
		testutil.RequireForbiddenError(
			t,
			err,
			"must not create webhook subscription in another organization",
		)
	})
}
