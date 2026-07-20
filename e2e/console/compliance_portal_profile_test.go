// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

func TestCompliancePortal_UpdateProfile(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const compliancePortalQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					compliancePortal {
						id
					}
				}
			}
		}
	`

	var compliancePortalLookup struct {
		Node struct {
			CompliancePortal struct {
				ID string `json:"id"`
			} `json:"compliancePortal"`
		} `json:"node"`
	}

	err := owner.Execute(compliancePortalQuery, map[string]any{
		"organizationId": organizationID,
	}, &compliancePortalLookup)
	require.NoError(t, err)
	require.NotEmpty(t, compliancePortalLookup.Node.CompliancePortal.ID)

	compliancePortalID := compliancePortalLookup.Node.CompliancePortal.ID

	const updateMutation = `
		mutation UpdateCompliancePortal($input: UpdateCompliancePortalInput!) {
			updateCompliancePortal(input: $input) {
				compliancePortal {
					id
					title
					description
					websiteUrl
					email
					headquarterAddress
				}
			}
		}
	`

	var result struct {
		UpdateCompliancePortal struct {
			CompliancePortal struct {
				ID                 string  `json:"id"`
				Title              string  `json:"title"`
				Description        *string `json:"description"`
				WebsiteURL         *string `json:"websiteUrl"`
				Email              *string `json:"email"`
				HeadquarterAddress *string `json:"headquarterAddress"`
			} `json:"compliancePortal"`
		} `json:"updateCompliancePortal"`
	}

	err = owner.Execute(updateMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"title":              "Acme Security",
			"description":        "We keep your data safe.",
			"websiteUrl":         "https://example.com",
			"email":              "security@example.com",
			"headquarterAddress": "123 Main St, San Francisco, CA 94102",
		},
	}, &result)
	require.NoError(t, err)

	tc := result.UpdateCompliancePortal.CompliancePortal
	assert.Equal(t, compliancePortalID, tc.ID)
	assert.Equal(t, "Acme Security", tc.Title)
	require.NotNil(t, tc.Description)
	assert.Equal(t, "We keep your data safe.", *tc.Description)
	require.NotNil(t, tc.WebsiteURL)
	assert.Equal(t, "https://example.com", *tc.WebsiteURL)
	require.NotNil(t, tc.Email)
	assert.Equal(t, "security@example.com", *tc.Email)
	require.NotNil(t, tc.HeadquarterAddress)
	assert.Equal(t, "123 Main St, San Francisco, CA 94102", *tc.HeadquarterAddress)
}
