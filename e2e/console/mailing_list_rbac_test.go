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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_MailingList_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	t.Run(
		"create subscriber",
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
					name:     "owner can create subscriber",
					role:     testutil.RoleOwner,
					allowMsg: "owner should create mailing list subscribers",
				},
				{
					name:     "admin can create subscriber",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should create mailing list subscribers",
				},
				{
					name:         "viewer cannot create subscriber",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not create mailing list subscribers",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						portalID := createCompliancePortalMailingListPortal(t, owner)
						mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID

						client := org.Client(t, tt.role)
						err := tryCreateMailingListSubscriber(
							client,
							mailingListID,
							factory.SafeName("RBAC subscriber"),
							mailingListSyntheticEmail(),
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
		"send update",
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
					name:     "owner can send update",
					role:     testutil.RoleOwner,
					allowMsg: "owner should send mailing list updates",
				},
				{
					name:     "admin can send update",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should send mailing list updates",
				},
				{
					name:         "viewer cannot send update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not send mailing list updates",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						portalID := createCompliancePortalMailingListPortal(t, owner)
						mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID
						update := createMailingListUpdate(
							t,
							owner,
							mailingListID,
							factory.SafeName("RBAC send title"),
							factory.SafeName("RBAC send body"),
						)

						client := org.Client(t, tt.role)
						err := trySendMailingListUpdate(client, update.ID)

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
		"delete update",
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
					name:     "owner can delete update",
					role:     testutil.RoleOwner,
					allowMsg: "owner should delete mailing list updates",
				},
				{
					name:     "admin can delete update",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should delete mailing list updates",
				},
				{
					name:         "viewer cannot delete update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not delete mailing list updates",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						portalID := createCompliancePortalMailingListPortal(t, owner)
						mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID
						update := createMailingListUpdate(
							t,
							owner,
							mailingListID,
							factory.SafeName("RBAC delete title"),
							factory.SafeName("RBAC delete body"),
						)

						client := org.Client(t, tt.role)
						err := tryDeleteMailingListUpdate(client, update.ID)

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
		"delete subscriber",
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
					name:     "owner can delete subscriber",
					role:     testutil.RoleOwner,
					allowMsg: "owner should delete mailing list subscribers",
				},
				{
					name:     "admin can delete subscriber",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should delete mailing list subscribers",
				},
				{
					name:         "viewer cannot delete subscriber",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not delete mailing list subscribers",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						portalID := createCompliancePortalMailingListPortal(t, owner)
						mailingListID := queryCompliancePortalMailingList(t, owner, portalID).MailingList.ID
						subscriber := createMailingListSubscriber(
							t,
							owner,
							mailingListID,
							factory.SafeName("RBAC delete subscriber"),
							mailingListSyntheticEmail(),
							true,
						)

						client := org.Client(t, tt.role)
						err := tryDeleteMailingListSubscriber(client, subscriber.ID)

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
}
