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
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestGitHubAppConnector_Initiate(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	initiateURL := githubAppInitiateURL(t, owner.GetOrganizationID().String())

	t.Run("unauthenticated caller is rejected", func(t *testing.T) {
		t.Parallel()

		guest := testutil.NewUnauthenticatedClient(t)
		resp := guest.GetNoRedirect(initiateURL)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("viewer cannot initiate", func(t *testing.T) {
		t.Parallel()

		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		resp := viewer.GetNoRedirect(
			githubAppInitiateURL(t, viewer.GetOrganizationID().String()),
		)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("owner is rejected when github app is not configured", func(t *testing.T) {
		t.Parallel()

		resp := owner.ForTest(t).GetNoRedirect(initiateURL)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "github app is not configured")
	})
}

func githubAppInitiateURL(t *testing.T, organizationID string) string {
	t.Helper()

	u, err := url.Parse(testutil.GetBaseURL())
	require.NoError(t, err)

	u.Path = "/api/console/v1/connectors/github-app/initiate"
	q := u.Query()
	q.Set("organization_id", organizationID)
	u.RawQuery = q.Encode()

	return u.String()
}
