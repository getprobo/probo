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

package testutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestOrganizationRoles_Client(t *testing.T) {
	t.Parallel()

	sharedHTTP := &http.Client{}
	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, 1)
	ownerUserID := gid.New(tenantID, 2)
	adminUserID := gid.New(tenantID, 3)
	viewerUserID := gid.New(tenantID, 4)

	org := OrganizationRoles{
		owner: &Client{
			T:              t,
			httpClient:     sharedHTTP,
			role:           RoleOwner,
			userID:         ownerUserID,
			organizationID: orgID,
			email:          "owner@example.com",
		},
		admin: &Client{
			T:              t,
			httpClient:     sharedHTTP,
			role:           RoleAdmin,
			userID:         adminUserID,
			organizationID: orgID,
			email:          "admin@example.com",
		},
		viewer: &Client{
			T:              t,
			httpClient:     sharedHTTP,
			role:           RoleViewer,
			userID:         viewerUserID,
			organizationID: orgID,
			email:          "viewer@example.com",
		},
	}

	t.Run(
		"returns shallow copy bound to leaf testing.TB",
		func(t *testing.T) {
			t.Parallel()

			leafClient := org.Client(t, RoleOwner)

			require.NotSame(t, org.owner, leafClient)
			assert.Same(t, t, leafClient.T)
			assert.Same(t, sharedHTTP, leafClient.httpClient)
			assert.Equal(t, ownerUserID, leafClient.GetUserID())
			assert.Equal(t, orgID, leafClient.GetOrganizationID())
			assert.Equal(t, RoleOwner, leafClient.GetRole())
			assert.Equal(t, "owner@example.com", leafClient.GetEmail())
		},
	)

	t.Run(
		"each call returns a distinct pointer",
		func(t *testing.T) {
			t.Parallel()

			first := org.Client(t, RoleOwner)
			second := org.Client(t, RoleOwner)

			assert.NotSame(t, first, second)
			assert.NotSame(t, org.owner, first)
		},
	)

	t.Run(
		"selects role clients correctly",
		func(t *testing.T) {
			t.Parallel()

			adminClient := org.Client(t, RoleAdmin)
			viewerClient := org.Client(t, RoleViewer)

			assert.Equal(t, RoleAdmin, adminClient.GetRole())
			assert.Equal(t, adminUserID, adminClient.GetUserID())
			assert.NotSame(t, org.admin, adminClient)

			assert.Equal(t, RoleViewer, viewerClient.GetRole())
			assert.Equal(t, viewerUserID, viewerClient.GetUserID())
			assert.NotSame(t, org.viewer, viewerClient)
		},
	)
}
