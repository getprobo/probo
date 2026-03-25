// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestSCIMEvent_Export(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const mutation = `
		mutation($input: RequestSCIMEventExportInput!) {
			requestSCIMEventExport(input: $input) {
				logExportId
			}
		}
	`

	t.Run("owner can request export", func(t *testing.T) {
		t.Parallel()

		var result struct {
			RequestSCIMEventExport struct {
				LogExportID string `json:"logExportId"`
			} `json:"requestSCIMEventExport"`
		}

		err := owner.Execute(mutation, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"fromTime":       "2026-01-01T00:00:00Z",
				"toTime":         "2026-03-24T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.RequestSCIMEventExport.LogExportID)
	})

	t.Run("admin can request export", func(t *testing.T) {
		t.Parallel()
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

		var result struct {
			RequestSCIMEventExport struct {
				LogExportID string `json:"logExportId"`
			} `json:"requestSCIMEventExport"`
		}

		err := admin.Execute(mutation, map[string]any{
			"input": map[string]any{
				"organizationId": admin.GetOrganizationID().String(),
				"fromTime":       "2026-01-01T00:00:00Z",
				"toTime":         "2026-03-24T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.RequestSCIMEventExport.LogExportID)
	})

	t.Run("viewer cannot request export", func(t *testing.T) {
		t.Parallel()
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.Do(mutation, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"fromTime":       "2026-01-01T00:00:00Z",
				"toTime":         "2026-03-24T00:00:00Z",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot request SCIM event export")
	})
}
