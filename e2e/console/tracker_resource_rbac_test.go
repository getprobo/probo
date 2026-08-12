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

func TestTrackerResource_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		_, err := viewer.Do(`
			mutation CreateTrackerResource($input: CreateTrackerResourceInput!) {
				createTrackerResource(input: $input) {
					trackerResourceEdge { node { id } }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"type":             "SCRIPT",
				"origin":           "https://example.com",
				"path":             "/viewer.js",
				"displayName":      "Test Viewer Resource",
				"description":      "Should fail",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create tracker resource")
	})

	t.Run("viewer cannot update resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		resourceID := factory.CreateTrackerResource(owner, categoryID)

		_, err := viewer.Do(`
			mutation UpdateTrackerResource($input: UpdateTrackerResourceInput!) {
				updateTrackerResource(input: $input) {
					trackerResource { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerResourceId": resourceID,
				"description":       "Updated by Viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update tracker resource")
	})

	t.Run("viewer cannot delete resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		resourceID := factory.CreateTrackerResource(owner, categoryID)

		_, err := viewer.Do(`
			mutation DeleteTrackerResource($input: DeleteTrackerResourceInput!) {
				deleteTrackerResource(input: $input) {
					deletedTrackerResourceId
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"trackerResourceId": resourceID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete tracker resource")
	})

	t.Run("viewer cannot move resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-res-move-a"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-res-move-b"})
		resourceID := factory.CreateTrackerResource(owner, categoryA)

		_, err := viewer.Do(`
			mutation MoveTrackerResourceToCategory($input: MoveTrackerResourceToCategoryInput!) {
				moveTrackerResourceToCategory(input: $input) {
					trackerResource { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerResourceId":      resourceID,
				"targetCookieCategoryId": categoryB,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to move tracker resource")
	})
}
