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

func TestClient_ForTest(t *testing.T) {
	t.Parallel()

	sharedHTTP := &http.Client{}
	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, 1)
	userID := gid.New(tenantID, 2)
	profileID := gid.New(tenantID, 3)

	stored := &Client{
		T:              t,
		httpClient:     sharedHTTP,
		role:           RoleOwner,
		userID:         userID,
		profileID:      profileID,
		organizationID: orgID,
		email:          "owner@example.com",
	}

	t.Run(
		"returns shallow copy bound to leaf testing.TB",
		func(t *testing.T) {
			t.Parallel()

			leafClient := stored.ForTest(t)

			require.NotSame(t, stored, leafClient)
			assert.Same(t, t, leafClient.T)
			assert.Same(t, sharedHTTP, leafClient.httpClient)
			assert.Equal(t, userID, leafClient.GetUserID())
			assert.Equal(t, profileID, leafClient.GetProfileID())
			assert.Equal(t, orgID, leafClient.GetOrganizationID())
			assert.Equal(t, RoleOwner, leafClient.GetRole())
			assert.Equal(t, "owner@example.com", leafClient.GetEmail())
		},
	)

	t.Run(
		"each call returns a distinct pointer",
		func(t *testing.T) {
			t.Parallel()

			first := stored.ForTest(t)
			second := stored.ForTest(t)

			assert.NotSame(t, first, second)
			assert.NotSame(t, stored, first)
		},
	)
}
