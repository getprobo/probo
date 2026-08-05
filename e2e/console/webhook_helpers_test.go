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
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	createWebhookSubscriptionMutation = `
		mutation CreateWebhookSubscription($input: CreateWebhookSubscriptionInput!) {
			createWebhookSubscription(input: $input) {
				webhookSubscriptionEdge {
					node {
						id
						endpointUrl
						selectedEvents
						createdAt
						updatedAt
					}
				}
			}
		}
	`

	updateWebhookSubscriptionMutation = `
		mutation UpdateWebhookSubscription($input: UpdateWebhookSubscriptionInput!) {
			updateWebhookSubscription(input: $input) {
				webhookSubscription {
					id
					endpointUrl
					selectedEvents
					createdAt
					updatedAt
				}
			}
		}
	`

	deleteWebhookSubscriptionMutation = `
		mutation DeleteWebhookSubscription($input: DeleteWebhookSubscriptionInput!) {
			deleteWebhookSubscription(input: $input) {
				deletedWebhookSubscriptionId
			}
		}
	`

	webhookSubscriptionNodeQuery = `
		query WebhookSubscriptionNode($id: ID!) {
			node(id: $id) {
				... on WebhookSubscription {
					id
					endpointUrl
					selectedEvents
					createdAt
					updatedAt
					signingSecret
					organization {
						id
					}
					permission(action: "core:webhook-subscription:update")
					events(first: 10) {
						totalCount
						edges {
							node {
								id
								webhookSubscriptionId
								status
								createdAt
							}
						}
					}
				}
			}
		}
	`

	organizationWebhookSubscriptionsQuery = `
		query OrganizationWebhookSubscriptions($id: ID!) {
			node(id: $id) {
				... on Organization {
					webhookSubscriptions(first: 50) {
						totalCount
						edges {
							node {
								id
								endpointUrl
							}
						}
					}
				}
			}
		}
	`
)

type (
	webhookSubscriptionNode struct {
		ID             string    `json:"id"`
		EndpointURL    string    `json:"endpointUrl"`
		SelectedEvents []string  `json:"selectedEvents"`
		CreatedAt      time.Time `json:"createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"`
	}

	webhookSubscriptionDetail struct {
		webhookSubscriptionNode
		SigningSecret string `json:"signingSecret"`
		Organization  struct {
			ID string `json:"id"`
		} `json:"organization"`
		Permission bool `json:"permission"`
		Events     struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID                    string    `json:"id"`
					WebhookSubscriptionID string    `json:"webhookSubscriptionId"`
					Status                string    `json:"status"`
					CreatedAt             time.Time `json:"createdAt"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"events"`
	}

	webhookSubscriptionNodeResponse struct {
		Node *webhookSubscriptionDetail `json:"node"`
	}

	organizationWebhookSubscriptionsResponse struct {
		Node struct {
			WebhookSubscriptions struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID          string `json:"id"`
						EndpointURL string `json:"endpointUrl"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"webhookSubscriptions"`
		} `json:"node"`
	}
)

// unroutableWebhookEndpoint returns a syntactically valid HTTPS URL that should
// not accept deliveries, keeping webhook worker attempts bounded.
func unroutableWebhookEndpoint(t *testing.T) string {
	t.Helper()

	endpoint := url.URL{
		Scheme: "https",
		Host:   "example.invalid",
		Path:   "/" + factory.SafeName("webhook"),
	}

	return endpoint.String()
}

func thirdPartyWebhookEventTypes() []string {
	return []string{
		"THIRD_PARTY_CREATED",
		"THIRD_PARTY_UPDATED",
		"THIRD_PARTY_DELETED",
	}
}

func createWebhookSubscription(
	t *testing.T,
	client *testutil.Client,
	endpointURL string,
	selectedEvents []string,
) webhookSubscriptionNode {
	t.Helper()

	var result struct {
		CreateWebhookSubscription struct {
			WebhookSubscriptionEdge struct {
				Node webhookSubscriptionNode `json:"node"`
			} `json:"webhookSubscriptionEdge"`
		} `json:"createWebhookSubscription"`
	}

	err := client.Execute(
		createWebhookSubscriptionMutation,
		map[string]any{
			"input": map[string]any{
				"organizationId": client.GetOrganizationID().String(),
				"endpointUrl":    endpointURL,
				"selectedEvents": selectedEvents,
			},
		},
		&result,
	)
	require.NoError(t, err, "create webhook subscription")

	node := result.CreateWebhookSubscription.WebhookSubscriptionEdge.Node
	require.NotEmpty(t, node.ID)

	subscriptionID := node.ID

	t.Cleanup(func() {
		bestEffortDeleteWebhookSubscription(client, subscriptionID)
	})

	return node
}

func bestEffortDeleteWebhookSubscription(
	client *testutil.Client,
	subscriptionID string,
) {
	_, err := client.Do(
		deleteWebhookSubscriptionMutation,
		map[string]any{
			"input": map[string]any{
				"webhookSubscriptionId": subscriptionID,
			},
		},
	)
	if err == nil {
		return
	}

	var gqlErrors testutil.GraphQLErrors
	if errors.As(err, &gqlErrors) {
		for _, gqlErr := range gqlErrors {
			if gqlErr.Code() == "NOT_FOUND" {
				return
			}
		}
	}
}

func requireWebhookEventsEventually(
	t *testing.T,
	client *testutil.Client,
	subscriptionID string,
	minCount int,
) webhookSubscriptionNodeResponse {
	t.Helper()

	var (
		last    webhookSubscriptionNodeResponse
		lastErr error
	)

	ok := testutil.Poll(
		t,
		90*time.Second,
		500*time.Millisecond,
		func() bool {
			lastErr = client.Execute(
				webhookSubscriptionNodeQuery,
				map[string]any{"id": subscriptionID},
				&last,
			)
			if lastErr != nil || last.Node == nil || last.Node.Events.TotalCount < minCount {
				return false
			}

			for _, edge := range last.Node.Events.Edges {
				if edge.Node.Status != "PENDING" {
					return true
				}
			}

			return false
		},
	)
	if !ok {
		require.NoError(t, lastErr, "last webhook event query failed")
		require.FailNow(t, "webhook event did not leave PENDING state")
	}

	return last
}

func organizationContainsWebhookSubscription(
	t *testing.T,
	client *testutil.Client,
	orgID string,
	subscriptionID string,
) bool {
	t.Helper()

	var result organizationWebhookSubscriptionsResponse

	err := client.Execute(
		organizationWebhookSubscriptionsQuery,
		map[string]any{"id": orgID},
		&result,
	)
	require.NoError(t, err)

	for _, edge := range result.Node.WebhookSubscriptions.Edges {
		if edge.Node.ID == subscriptionID {
			return true
		}
	}

	return false
}
