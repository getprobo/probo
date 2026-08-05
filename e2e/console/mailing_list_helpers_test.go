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
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	mailingListUpdateTitleMaxLength = 200

	compliancePortalMailingListQuery = `
		query CompliancePortalMailingList($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					id
					canListSubscribers: permission(action: "compliance-portal:mailing-list-subscriber:list")
					canCreateSubscriber: permission(action: "compliance-portal:mailing-list-subscriber:create")
					canDeleteSubscriber: permission(action: "compliance-portal:mailing-list-subscriber:delete")
					canListUpdates: permission(action: "compliance-portal:mailing-list-update:list")
					canCreateUpdate: permission(action: "compliance-portal:mailing-list-update:create")
					canUpdateUpdate: permission(action: "compliance-portal:mailing-list-update:update")
					canSendUpdate: permission(action: "compliance-portal:mailing-list-update:send")
					canDeleteUpdate: permission(action: "compliance-portal:mailing-list-update:delete")
					canUpdateMailingList: permission(action: "compliance-portal:mailing-list:update")
					mailingList {
						id
						replyTo
						subscribers(first: 20) {
							totalCount
							edges {
								node {
									id
									fullName
									email
									status
								}
							}
						}
						updates(first: 20) {
							totalCount
							edges {
								node {
									id
									title
									body
									status
								}
							}
						}
					}
				}
			}
		}
	`

	mailingListSubscriberNodeQuery = `
		query MailingListSubscriberNode($id: ID!) {
			node(id: $id) {
				... on MailingListSubscriber {
					id
					fullName
					email
					status
					createdAt
					updatedAt
				}
			}
		}
	`

	mailingListUpdateNodeQuery = `
		query MailingListUpdateNode($id: ID!) {
			node(id: $id) {
				... on MailingListUpdate {
					id
					title
					body
					status
					createdAt
					updatedAt
				}
			}
		}
	`

	createMailingListSubscriberMutation = `
		mutation CreateMailingListSubscriber($input: CreateMailingListSubscriberInput!) {
			createMailingListSubscriber(input: $input) {
				mailingListSubscriberEdge {
					cursor
					node {
						id
						fullName
						email
						status
						createdAt
						updatedAt
					}
				}
			}
		}
	`

	deleteMailingListSubscriberMutation = `
		mutation DeleteMailingListSubscriber($input: DeleteMailingListSubscriberInput!) {
			deleteMailingListSubscriber(input: $input) {
				deletedMailingListSubscriberId
			}
		}
	`

	updateMailingListMutation = `
		mutation UpdateMailingList($input: UpdateMailingListInput!) {
			updateMailingList(input: $input) {
				mailingList {
					id
					replyTo
				}
			}
		}
	`

	createMailingListUpdateMutation = `
		mutation CreateMailingListUpdate($input: CreateMailingListUpdateInput!) {
			createMailingListUpdate(input: $input) {
				mailingListUpdate {
					id
					title
					body
					status
					createdAt
					updatedAt
				}
			}
		}
	`

	updateMailingListUpdateMutation = `
		mutation UpdateMailingListUpdate($input: UpdateMailingListUpdateInput!) {
			updateMailingListUpdate(input: $input) {
				mailingListUpdate {
					id
					title
					body
					status
					updatedAt
				}
			}
		}
	`

	sendMailingListUpdateMutation = `
		mutation SendMailingListUpdate($input: SendMailingListUpdateInput!) {
			sendMailingListUpdate(input: $input) {
				mailingListUpdate {
					id
					status
				}
			}
		}
	`

	deleteMailingListUpdateMutation = `
		mutation DeleteMailingListUpdate($input: DeleteMailingListUpdateInput!) {
			deleteMailingListUpdate(input: $input) {
				deletedMailingListUpdateId
			}
		}
	`
)

type (
	mailingListSubscriberWire struct {
		ID        string    `json:"id"`
		FullName  string    `json:"fullName"`
		Email     string    `json:"email"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	mailingListUpdateWire struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	mailingListWire struct {
		ID          string  `json:"id"`
		ReplyTo     *string `json:"replyTo"`
		Subscribers struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node mailingListSubscriberWire `json:"node"`
			} `json:"edges"`
		} `json:"subscribers"`
		Updates struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node mailingListUpdateWire `json:"node"`
			} `json:"edges"`
		} `json:"updates"`
	}

	compliancePortalMailingListWire struct {
		ID                   string           `json:"id"`
		CanListSubscribers   bool             `json:"canListSubscribers"`
		CanCreateSubscriber  bool             `json:"canCreateSubscriber"`
		CanDeleteSubscriber  bool             `json:"canDeleteSubscriber"`
		CanListUpdates       bool             `json:"canListUpdates"`
		CanCreateUpdate      bool             `json:"canCreateUpdate"`
		CanUpdateUpdate      bool             `json:"canUpdateUpdate"`
		CanSendUpdate        bool             `json:"canSendUpdate"`
		CanDeleteUpdate      bool             `json:"canDeleteUpdate"`
		CanUpdateMailingList bool             `json:"canUpdateMailingList"`
		MailingList          *mailingListWire `json:"mailingList"`
	}

	compliancePortalMailingListResponse struct {
		Node *compliancePortalMailingListWire `json:"node"`
	}
)

func mailingListUpdateExpectedSubject(organizationName, updateTitle string) string {
	return fmt.Sprintf("%s – %s", organizationName, updateTitle)
}

func queryOrganizationName(t *testing.T, client *testutil.Client) string {
	t.Helper()

	const query = `
		query OrganizationName($id: ID!) {
			node(id: $id) {
				... on Organization {
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

	err := client.Execute(
		query,
		map[string]any{
			"id": client.GetOrganizationID().String(),
		},
		&result,
	)
	require.NoError(t, err, "query organization name")
	require.NotNil(t, result.Node)
	require.NotEmpty(t, result.Node.Name)

	return result.Node.Name
}

func postMailingListConfirmSubscription(t *testing.T, confirmURL string) {
	t.Helper()

	resp, err := http.Post(confirmURL, "", nil)
	require.NoError(t, err, "post mailing list confirm subscription")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm subscription should succeed")
}

func mailingListSyntheticEmail() string {
	localPart := strings.TrimSuffix(factory.SafeEmail(), "@example.com")

	return fmt.Sprintf("%s@e2e.probo.test", localPart)
}

func queryCompliancePortalMailingList(
	t *testing.T,
	client *testutil.Client,
	portalID string,
) compliancePortalMailingListWire {
	t.Helper()

	var result compliancePortalMailingListResponse

	err := client.Execute(
		compliancePortalMailingListQuery,
		map[string]any{
			"id": portalID,
		},
		&result,
	)
	require.NoError(t, err, "query compliance portal mailing list")
	require.NotNil(t, result.Node, "compliance portal node must be present")

	return *result.Node
}

func createCompliancePortalMailingListPortal(t *testing.T, client *testutil.Client) string {
	t.Helper()

	return factory.CreateCompliancePortal(
		client,
		factory.Attrs{"entityName": factory.SafeName("Mailing list portal")},
	)
}

func createMailingListSubscriber(
	t *testing.T,
	client *testutil.Client,
	mailingListID string,
	fullName string,
	email string,
	confirmed bool,
) mailingListSubscriberWire {
	t.Helper()

	var result struct {
		CreateMailingListSubscriber struct {
			MailingListSubscriberEdge struct {
				Node mailingListSubscriberWire `json:"node"`
			} `json:"mailingListSubscriberEdge"`
		} `json:"createMailingListSubscriber"`
	}

	input := map[string]any{
		"mailingListId": mailingListID,
		"fullName":      fullName,
		"email":         email,
		"confirmed":     confirmed,
	}

	err := client.Execute(
		createMailingListSubscriberMutation,
		map[string]any{
			"input": input,
		},
		&result,
	)
	require.NoError(t, err, "create mailing list subscriber")

	return result.CreateMailingListSubscriber.MailingListSubscriberEdge.Node
}

func tryCreateMailingListSubscriber(
	client *testutil.Client,
	mailingListID string,
	fullName string,
	email string,
) error {
	_, err := client.Do(
		createMailingListSubscriberMutation,
		map[string]any{
			"input": map[string]any{
				"mailingListId": mailingListID,
				"fullName":      fullName,
				"email":         email,
				"confirmed":     true,
			},
		},
	)

	return err
}

func deleteMailingListSubscriber(t *testing.T, client *testutil.Client, subscriberID string) {
	t.Helper()

	var result struct {
		DeleteMailingListSubscriber struct {
			DeletedMailingListSubscriberID string `json:"deletedMailingListSubscriberId"`
		} `json:"deleteMailingListSubscriber"`
	}

	err := client.Execute(
		deleteMailingListSubscriberMutation,
		map[string]any{
			"input": map[string]any{
				"id": subscriberID,
			},
		},
		&result,
	)
	require.NoError(t, err, "delete mailing list subscriber")
	assert.Equal(t, subscriberID, result.DeleteMailingListSubscriber.DeletedMailingListSubscriberID)
}

func tryDeleteMailingListSubscriber(client *testutil.Client, subscriberID string) error {
	_, err := client.Do(
		deleteMailingListSubscriberMutation,
		map[string]any{
			"input": map[string]any{
				"id": subscriberID,
			},
		},
	)

	return err
}

func updateMailingListReplyTo(
	t *testing.T,
	client *testutil.Client,
	mailingListID string,
	replyTo string,
) {
	t.Helper()

	var result struct {
		UpdateMailingList struct {
			MailingList struct {
				ID      string  `json:"id"`
				ReplyTo *string `json:"replyTo"`
			} `json:"mailingList"`
		} `json:"updateMailingList"`
	}

	err := client.Execute(
		updateMailingListMutation,
		map[string]any{
			"input": map[string]any{
				"id":      mailingListID,
				"replyTo": replyTo,
			},
		},
		&result,
	)
	require.NoError(t, err, "update mailing list")
	require.NotNil(t, result.UpdateMailingList.MailingList.ReplyTo)
	assert.Equal(t, replyTo, *result.UpdateMailingList.MailingList.ReplyTo)
}

func createMailingListUpdate(
	t *testing.T,
	client *testutil.Client,
	mailingListID string,
	title string,
	body string,
) mailingListUpdateWire {
	t.Helper()

	var result struct {
		CreateMailingListUpdate struct {
			MailingListUpdate mailingListUpdateWire `json:"mailingListUpdate"`
		} `json:"createMailingListUpdate"`
	}

	err := client.Execute(
		createMailingListUpdateMutation,
		map[string]any{
			"input": map[string]any{
				"mailingListId": mailingListID,
				"title":         title,
				"body":          body,
			},
		},
		&result,
	)
	require.NoError(t, err, "create mailing list update")

	return result.CreateMailingListUpdate.MailingListUpdate
}

func tryCreateMailingListUpdate(
	client *testutil.Client,
	mailingListID string,
	title string,
	body string,
) error {
	_, err := client.Do(
		createMailingListUpdateMutation,
		map[string]any{
			"input": map[string]any{
				"mailingListId": mailingListID,
				"title":         title,
				"body":          body,
			},
		},
	)

	return err
}

func updateMailingListUpdateFields(
	t *testing.T,
	client *testutil.Client,
	updateID string,
	title string,
	body string,
) mailingListUpdateWire {
	t.Helper()

	var result struct {
		UpdateMailingListUpdate struct {
			MailingListUpdate mailingListUpdateWire `json:"mailingListUpdate"`
		} `json:"updateMailingListUpdate"`
	}

	input := map[string]any{"id": updateID}
	if title != "" {
		input["title"] = title
	}

	if body != "" {
		input["body"] = body
	}

	err := client.Execute(
		updateMailingListUpdateMutation,
		map[string]any{
			"input": input,
		},
		&result,
	)
	require.NoError(t, err, "update mailing list update")

	return result.UpdateMailingListUpdate.MailingListUpdate
}

func sendMailingListUpdate(t *testing.T, client *testutil.Client, updateID string) string {
	t.Helper()

	var result struct {
		SendMailingListUpdate struct {
			MailingListUpdate struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"mailingListUpdate"`
		} `json:"sendMailingListUpdate"`
	}

	err := client.Execute(
		sendMailingListUpdateMutation,
		map[string]any{
			"input": map[string]any{
				"id": updateID,
			},
		},
		&result,
	)
	require.NoError(t, err, "send mailing list update")
	assert.Equal(t, "ENQUEUED", result.SendMailingListUpdate.MailingListUpdate.Status)

	return result.SendMailingListUpdate.MailingListUpdate.Status
}

func trySendMailingListUpdate(client *testutil.Client, updateID string) error {
	_, err := client.Do(
		sendMailingListUpdateMutation,
		map[string]any{
			"input": map[string]any{
				"id": updateID,
			},
		},
	)

	return err
}

func deleteMailingListUpdate(t *testing.T, client *testutil.Client, updateID string) {
	t.Helper()

	var result struct {
		DeleteMailingListUpdate struct {
			DeletedMailingListUpdateID string `json:"deletedMailingListUpdateId"`
		} `json:"deleteMailingListUpdate"`
	}

	err := client.Execute(
		deleteMailingListUpdateMutation,
		map[string]any{
			"input": map[string]any{
				"id": updateID,
			},
		},
		&result,
	)
	require.NoError(t, err, "delete mailing list update")
	assert.Equal(t, updateID, result.DeleteMailingListUpdate.DeletedMailingListUpdateID)
}

func tryDeleteMailingListUpdate(client *testutil.Client, updateID string) error {
	_, err := client.Do(
		deleteMailingListUpdateMutation,
		map[string]any{
			"input": map[string]any{
				"id": updateID,
			},
		},
	)

	return err
}

func queryMailingListSubscriberNode(
	t *testing.T,
	client *testutil.Client,
	subscriberID string,
) mailingListSubscriberWire {
	t.Helper()

	var result struct {
		Node *mailingListSubscriberWire `json:"node"`
	}

	err := client.Execute(
		mailingListSubscriberNodeQuery,
		map[string]any{
			"id": subscriberID,
		},
		&result,
	)
	require.NoError(t, err, "query mailing list subscriber node")
	require.NotNil(t, result.Node)

	return *result.Node
}

func queryMailingListUpdateNode(
	t *testing.T,
	client *testutil.Client,
	updateID string,
) mailingListUpdateWire {
	t.Helper()

	var result struct {
		Node *mailingListUpdateWire `json:"node"`
	}

	err := client.Execute(
		mailingListUpdateNodeQuery,
		map[string]any{
			"id": updateID,
		},
		&result,
	)
	require.NoError(t, err, "query mailing list update node")
	require.NotNil(t, result.Node)

	return *result.Node
}

func requireMailingListUpdateStatusEventually(
	t *testing.T,
	client *testutil.Client,
	updateID string,
	expectedStatus string,
) {
	t.Helper()

	var (
		last    mailingListUpdateWire
		lastErr error
	)

	ok := assert.Eventually(
		t,
		func() bool {
			var result struct {
				Node *mailingListUpdateWire `json:"node"`
			}

			lastErr = client.Execute(
				mailingListUpdateNodeQuery,
				map[string]any{
					"id": updateID,
				},
				&result,
			)
			if lastErr != nil || result.Node == nil {
				return false
			}

			last = *result.Node

			return last.Status == expectedStatus
		},
		90*time.Second,
		500*time.Millisecond,
		"mailing list update should reach status %s",
		expectedStatus,
	)
	if !ok {
		require.NoError(t, lastErr, "last mailing list update query failed")
		require.Equal(t, expectedStatus, last.Status, "mailing list update status mismatch")
	}
}

func requireMailpitConfirmationURLEventually(
	t *testing.T,
	client *testutil.Client,
	searchQuery string,
) string {
	t.Helper()

	var (
		linkURL string
		lastErr error
	)

	ok := assert.Eventually(
		t,
		func() bool {
			linkURL, lastErr = client.FindLinkFromMailpitSearch(searchQuery, "/mail-actions/confirm")
			return lastErr == nil && linkURL != ""
		},
		90*time.Second,
		500*time.Millisecond,
		"mailpit should contain a confirmation link for query %s",
		searchQuery,
	)
	if !ok {
		require.NoError(t, lastErr, "last mailpit confirmation link search failed")
		require.FailNow(t, "mailpit confirmation link not found")
	}

	return linkURL
}

func requireMailingListUpdateEmailEventually(
	t *testing.T,
	client *testutil.Client,
	recipientEmail string,
	subjectSearchKeyword string,
	expectedSubject string,
	expectedBody string,
) *testutil.MailpitMessageDetail {
	t.Helper()

	searchQuery := fmt.Sprintf(
		"to:%s subject:\"%s\"",
		recipientEmail,
		subjectSearchKeyword,
	)

	var (
		detail  *testutil.MailpitMessageDetail
		lastErr error
	)

	ok := assert.Eventually(
		t,
		func() bool {
			match, err := client.FindMailpitMessage(
				searchQuery,
				func(message *testutil.MailpitMessageDetail) bool {
					if message.Subject != expectedSubject {
						return false
					}

					if !strings.Contains(message.Text, expectedBody) &&
						!strings.Contains(message.HTML, expectedBody) {
						return false
					}

					for _, to := range message.To {
						if to.Address == recipientEmail {
							return true
						}
					}

					return false
				},
			)
			if err != nil {
				lastErr = err

				return false
			}

			detail = match
			lastErr = nil

			return true
		},
		90*time.Second,
		500*time.Millisecond,
		"subscriber should receive mailing list update email",
	)
	if !ok {
		if lastErr != nil {
			require.NoError(t, lastErr, "last mailpit poll failed")
		}

		require.FailNow(t, "mailing list update email did not match expected subject/body/recipient")
	}

	require.Equal(t, expectedSubject, detail.Subject)
	require.True(
		t,
		strings.Contains(detail.Text, expectedBody) || strings.Contains(detail.HTML, expectedBody),
		"mail body must contain update text",
	)

	return detail
}
