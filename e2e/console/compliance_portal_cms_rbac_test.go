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

func TestCompliancePortal_CMS_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)
	viewer := org.Client(t, testutil.RoleViewer)

	tests := []struct {
		name    string
		mutate  func(portalID string) error
		message string
	}{
		{
			name: "create reference",
			mutate: func(portalID string) error {
				return tryCreateCompliancePortalCMSReference(
					viewer,
					portalID,
					factory.SafeName("RBAC reference"),
					"https://example.com/rbac",
				)
			},
			message: "viewer must not create compliance portal references",
		},
		{
			name: "create commitment group",
			mutate: func(portalID string) error {
				return tryCreateCompliancePortalCMSCommitmentGroup(
					viewer,
					portalID,
					factory.SafeName("RBAC group"),
				)
			},
			message: "viewer must not create commitment groups",
		},
		{
			name: "create custom link",
			mutate: func(portalID string) error {
				return tryCreateCompliancePortalCMSCustomLink(
					viewer,
					portalID,
					factory.SafeName("RBAC link"),
					"https://example.com/rbac",
				)
			},
			message: "viewer must not create custom links",
		},
		{
			name: "create portal file",
			mutate: func(portalID string) error {
				return tryCreateCompliancePortalCMSPortalFile(viewer, portalID)
			},
			message: "viewer must not upload compliance portal files",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				owner := org.Client(t, testutil.RoleOwner)
				portalID := factory.CreateCompliancePortal(
					owner,
					factory.Attrs{"entityName": factory.SafeName("RBAC portal " + tt.name)},
				)

				err := tt.mutate(portalID)
				testutil.RequireForbiddenError(t, err, tt.message)
			},
		)
	}
}
