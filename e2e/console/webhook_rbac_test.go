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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestWebhookSubscription_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const readWebhookSubscriptionQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on WebhookSubscription {
					id
					endpointUrl
					selectedEvents
					events(first: 1) {
						totalCount
					}
				}
			}
		}
	`

	t.Run(
		"create",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can create",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to create webhook subscription",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create webhook subscription",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create webhook subscription",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							createWebhookSubscriptionMutation,
							map[string]any{
								"input": map[string]any{
									"organizationId": client.GetOrganizationID().String(),
									"endpointUrl":    unroutableWebhookEndpoint(t),
									"selectedEvents": thirdPartyWebhookEventTypes(),
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"update",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can update",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to update webhook subscription",
				},
				{
					name:     "admin can update",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to update webhook subscription",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to update webhook subscription",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						subscriptionID := createWebhookSubscription(
							t,
							ownerClient,
							unroutableWebhookEndpoint(t),
							thirdPartyWebhookEventTypes(),
						).ID

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateWebhookSubscriptionMutation,
							map[string]any{
								"input": map[string]any{
									"id":          subscriptionID,
									"endpointUrl": unroutableWebhookEndpoint(t),
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"delete",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can delete",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to delete webhook subscription",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete webhook subscription",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete webhook subscription",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						subscriptionID := createWebhookSubscription(
							t,
							ownerClient,
							unroutableWebhookEndpoint(t),
							thirdPartyWebhookEventTypes(),
						).ID

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteWebhookSubscriptionMutation,
							map[string]any{
								"input": map[string]any{
									"webhookSubscriptionId": subscriptionID,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"read",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				allowMsg    string
				nodePresent string
			}{
				{
					name:        "owner can read",
					role:        testutil.RoleOwner,
					allowMsg:    "owner should be able to read webhook subscription",
					nodePresent: "owner should receive webhook subscription data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read webhook subscription",
					nodePresent: "admin should receive webhook subscription data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read webhook subscription",
					nodePresent: "viewer should receive webhook subscription data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						subscriptionID := createWebhookSubscription(
							t,
							ownerClient,
							unroutableWebhookEndpoint(t),
							thirdPartyWebhookEventTypes(),
						).ID

						client := org.Client(t, tt.role)

						var result struct {
							Node *struct {
								ID             string   `json:"id"`
								EndpointURL    string   `json:"endpointUrl"`
								SelectedEvents []string `json:"selectedEvents"`
								Events         struct {
									TotalCount int `json:"totalCount"`
								} `json:"events"`
							} `json:"node"`
						}

						err := client.Execute(
							readWebhookSubscriptionQuery,
							map[string]any{"id": subscriptionID},
							&result,
						)
						require.NoError(t, err, tt.allowMsg)
						require.NotNil(t, result.Node, tt.nodePresent)
						assert.Equal(t, subscriptionID, result.Node.ID)
					},
				)
			}
		},
	)

	t.Run(
		"signingSecret",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)

			subscriptionID := createWebhookSubscription(
				t,
				ownerClient,
				unroutableWebhookEndpoint(t),
				thirdPartyWebhookEventTypes(),
			).ID

			t.Run(
				"owner can read signing secret",
				func(t *testing.T) {
					t.Parallel()

					const query = `
						query($id: ID!) {
							node(id: $id) {
								... on WebhookSubscription {
									signingSecret
								}
							}
						}
					`

					var result struct {
						Node struct {
							SigningSecret string `json:"signingSecret"`
						} `json:"node"`
					}

					err := ownerClient.Execute(
						query,
						map[string]any{"id": subscriptionID},
						&result,
					)
					require.NoError(t, err)
					assert.NotEmpty(t, result.Node.SigningSecret)
				},
			)

			t.Run(
				"admin can read signing secret",
				func(t *testing.T) {
					t.Parallel()

					adminClient := org.Client(t, testutil.RoleAdmin)

					const query = `
						query($id: ID!) {
							node(id: $id) {
								... on WebhookSubscription {
									signingSecret
								}
							}
						}
					`

					var result struct {
						Node struct {
							SigningSecret string `json:"signingSecret"`
						} `json:"node"`
					}

					err := adminClient.Execute(
						query,
						map[string]any{"id": subscriptionID},
						&result,
					)
					require.NoError(t, err)
					assert.NotEmpty(t, result.Node.SigningSecret)
				},
			)

			t.Run(
				"viewer cannot read signing secret",
				func(t *testing.T) {
					t.Parallel()

					viewerClient := org.Client(t, testutil.RoleViewer)

					const query = `
						query($id: ID!) {
							node(id: $id) {
								... on WebhookSubscription {
									signingSecret
								}
							}
						}
					`

					_, err := viewerClient.Do(
						query,
						map[string]any{"id": subscriptionID},
					)
					testutil.RequireForbiddenError(t, err, "viewer must not read signing secret")
				},
			)
		},
	)
}
