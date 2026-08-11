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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_MailingList_OwnerLifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	portalID := createCompliancePortalMailingListPortal(t, owner)

	portal := queryCompliancePortalMailingList(t, owner, portalID)
	require.NotNil(t, portal.MailingList)
	assert.NotEmpty(t, portal.MailingList.ID)
	assert.True(t, portal.CanListSubscribers)
	assert.True(t, portal.CanCreateSubscriber)
	assert.True(t, portal.CanDeleteSubscriber)
	assert.True(t, portal.CanListUpdates)
	assert.True(t, portal.CanCreateUpdate)
	assert.True(t, portal.CanUpdateUpdate)
	assert.True(t, portal.CanSendUpdate)
	assert.True(t, portal.CanDeleteUpdate)
	assert.True(t, portal.CanUpdateMailingList)
	assert.Equal(t, 0, portal.MailingList.Subscribers.TotalCount)
	assert.Equal(t, 0, portal.MailingList.Updates.TotalCount)

	mailingListID := portal.MailingList.ID
	subscriberName := factory.SafeName("Mailing list subscriber")
	subscriberEmail := mailingListSyntheticEmail()

	subscriber := createMailingListSubscriber(
		t,
		owner,
		mailingListID,
		subscriberName,
		subscriberEmail,
		true,
	)
	assert.Equal(t, subscriberName, subscriber.FullName)
	assert.Equal(t, subscriberEmail, subscriber.Email)
	assert.Equal(t, "CONFIRMED", subscriber.Status)

	nodeSubscriber := queryMailingListSubscriberNode(t, owner, subscriber.ID)
	assert.Equal(t, subscriber.ID, nodeSubscriber.ID)
	assert.Equal(t, subscriberName, nodeSubscriber.FullName)
	assert.Equal(t, subscriberEmail, nodeSubscriber.Email)

	replyTo := "mailing-list-reply@e2e.probo.test"
	updateMailingListReplyTo(t, owner, mailingListID, replyTo)

	portalAfterSubscriber := queryCompliancePortalMailingList(t, owner, portalID)
	assert.Equal(t, 1, portalAfterSubscriber.MailingList.Subscribers.TotalCount)
	require.Len(t, portalAfterSubscriber.MailingList.Subscribers.Edges, 1)
	assert.Equal(t, subscriber.ID, portalAfterSubscriber.MailingList.Subscribers.Edges[0].Node.ID)

	draftTitle := factory.SafeName("Mailing list draft")
	draftBody := factory.SafeName("Draft body")
	draftUpdate := createMailingListUpdate(t, owner, mailingListID, draftTitle, draftBody)
	assert.Equal(t, "DRAFT", draftUpdate.Status)
	assert.Equal(t, draftTitle, draftUpdate.Title)
	assert.Equal(t, draftBody, draftUpdate.Body)

	updatedTitle := factory.SafeName("Mailing list updated title")
	updatedBody := factory.SafeName("Updated body content")
	updated := updateMailingListUpdateFields(t, owner, draftUpdate.ID, updatedTitle, updatedBody)
	assert.Equal(t, updatedTitle, updated.Title)
	assert.Equal(t, updatedBody, updated.Body)
	assert.Equal(t, "DRAFT", updated.Status)

	nodeUpdate := queryMailingListUpdateNode(t, owner, draftUpdate.ID)
	assert.Equal(t, updatedTitle, nodeUpdate.Title)
	assert.Equal(t, updatedBody, nodeUpdate.Body)

	deleteMailingListUpdate(t, owner, draftUpdate.ID)

	_, err := owner.Do(
		mailingListUpdateNodeQuery,
		map[string]any{"id": draftUpdate.ID},
	)

	var deletedNodeErrors testutil.GraphQLErrors
	require.ErrorAs(t, err, &deletedNodeErrors)
	require.Len(t, deletedNodeErrors, 1)
	assert.Equal(t, "NOT_FOUND", deletedNodeErrors[0].Code())

	portalAfterDelete := queryCompliancePortalMailingList(t, owner, portalID)
	assert.Equal(t, 0, portalAfterDelete.MailingList.Updates.TotalCount)

	sendTitle := factory.SafeName("Mailing list send title")
	sendBody := factory.SafeName("Mailing list send body unique")
	sendUpdate := createMailingListUpdate(t, owner, mailingListID, sendTitle, sendBody)
	sendMailingListUpdate(t, owner, sendUpdate.ID)

	requireMailingListUpdateStatusEventually(t, owner, sendUpdate.ID, "SENT")

	orgName := queryOrganizationName(t, owner)
	expectedSubject := mailingListUpdateExpectedSubject(orgName, sendTitle)
	mail := requireMailingListUpdateEmailEventually(
		t,
		owner,
		subscriberEmail,
		sendTitle,
		expectedSubject,
		sendBody,
	)
	assert.NotEmpty(t, mail.ID)

	deleteMailingListSubscriber(t, owner, subscriber.ID)

	portalAfterRemoval := queryCompliancePortalMailingList(t, owner, portalID)
	assert.Equal(t, 0, portalAfterRemoval.MailingList.Subscribers.TotalCount)
}

func TestCompliancePortal_MailingList_Validation(t *testing.T) {
	t.Parallel()

	t.Run(
		"create update requires title",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner).ForTest(t)
			portalID := createCompliancePortalMailingListPortal(t, owner)
			mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID

			err := tryCreateMailingListUpdate(
				owner,
				mailingListID,
				"",
				factory.SafeName("body"),
			)
			testutil.RequireErrorCode(t, err, "INVALID", "empty title must be rejected")
		},
	)

	t.Run(
		"create update rejects title longer than service limit",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner).ForTest(t)
			portalID := createCompliancePortalMailingListPortal(t, owner)
			mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID

			err := tryCreateMailingListUpdate(
				owner,
				mailingListID,
				strings.Repeat("a", mailingListUpdateTitleMaxLength+1),
				factory.SafeName("body"),
			)
			testutil.RequireErrorCode(t, err, "INVALID", "overlong title must be rejected")
		},
	)

	t.Run(
		"create subscriber requires full name",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner).ForTest(t)
			portalID := createCompliancePortalMailingListPortal(t, owner)
			mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID

			err := tryCreateMailingListSubscriber(
				owner,
				mailingListID,
				"",
				mailingListSyntheticEmail(),
			)
			testutil.RequireErrorCode(t, err, "INVALID", "empty full name must be rejected")
		},
	)
}

func TestCompliancePortal_MailingList_PendingSubscriberConfirmation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	portalID := createCompliancePortalMailingListPortal(t, owner)
	mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID

	subscriberName := factory.SafeName("Pending subscriber")
	subscriberEmail := mailingListSyntheticEmail()
	subscriber := createMailingListSubscriber(
		t,
		owner,
		mailingListID,
		subscriberName,
		subscriberEmail,
		false,
	)
	assert.Equal(t, "PENDING", subscriber.Status)

	searchQuery := fmt.Sprintf(
		"to:%s subject:\"Confirm Your Compliance Updates Subscription\"",
		subscriberEmail,
	)
	confirmURL := requireMailpitConfirmationURLEventually(t, owner, searchQuery)
	postMailingListConfirmSubscription(t, confirmURL)

	confirmed := queryMailingListSubscriberNode(t, owner, subscriber.ID)
	assert.Equal(t, "CONFIRMED", confirmed.Status)
	assert.Equal(t, subscriberEmail, confirmed.Email)
}
