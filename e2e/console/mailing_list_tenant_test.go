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

	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_MailingList_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1PortalID := createCompliancePortalMailingListPortal(t, org1Owner)
	org1MailingListID := queryCompliancePortalMailingList(t, org1Owner, org1PortalID).MailingList.ID

	org1Subscriber := createMailingListSubscriber(
		t,
		org1Owner,
		org1MailingListID,
		factory.SafeName("Org1 subscriber"),
		mailingListSyntheticEmail(),
		true,
	)
	org1Update := createMailingListUpdate(
		t,
		org1Owner,
		org1MailingListID,
		factory.SafeName("Org1 update title"),
		factory.SafeName("Org1 update body"),
	)

	t.Run(
		"cannot create subscriber on foreign mailing list",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := tryCreateMailingListSubscriber(
				org2,
				org1MailingListID,
				factory.SafeName("Cross-tenant subscriber"),
				mailingListSyntheticEmail(),
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not create subscribers on another organization's mailing list",
			)
		},
	)

	t.Run(
		"cannot create update on foreign mailing list",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := tryCreateMailingListUpdate(
				org2,
				org1MailingListID,
				factory.SafeName("Cross-tenant title"),
				factory.SafeName("Cross-tenant body"),
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not create updates on another organization's mailing list",
			)
		},
	)

	t.Run(
		"cannot delete foreign subscriber",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := tryDeleteMailingListSubscriber(org2, org1Subscriber.ID)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete subscribers from another organization's mailing list",
			)
		},
	)

	t.Run(
		"cannot send foreign update",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := trySendMailingListUpdate(org2, org1Update.ID)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not send updates from another organization's mailing list",
			)
		},
	)

	t.Run(
		"cannot delete foreign update",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := tryDeleteMailingListUpdate(org2, org1Update.ID)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete updates from another organization's mailing list",
			)
		},
	)
}
