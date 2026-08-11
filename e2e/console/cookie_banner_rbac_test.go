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

func TestCookieBanner_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.Do(`
			mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBannerEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":    viewer.GetOrganizationID().String(),
				"name":              factory.SafeName("Banner"),
				"origin":            factory.SafeOrigin(),
				"cookiePolicyUrl":   "https://example.com/cookies",
				"consentExpiryDays": 365,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create cookie banner")
	})

	t.Run("viewer cannot update", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)

		_, err := viewer.Do(`
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) {
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           "Updated",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update cookie banner")
	})

	t.Run("viewer cannot delete", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)

		_, err := viewer.Do(`
			mutation DeleteCookieBanner($input: DeleteCookieBannerInput!) {
				deleteCookieBanner(input: $input) {
					deletedCookieBannerId
				}
			}
		`, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete cookie banner")
	})
}
