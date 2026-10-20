// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_ListOrganizations(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	var result struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	mc.CallToolInto("listOrganizations", map[string]any{}, &result)

	assert.NotEmpty(t, result.Organizations)
}

func TestMCP_ListOrganizations_LegalName(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	legalName := "Listed Org Legal Name Inc."
	err := owner.ExecuteConnect(`
		mutation UpdateOrganization($input: UpdateOrganizationInput!) {
			updateOrganization(input: $input) {
				organization { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"legalName":      legalName,
		},
	}, nil)
	require.NoError(t, err)

	var result struct {
		Organizations []struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			LegalName *string `json:"legal_name"`
		} `json:"organizations"`
	}
	mc.CallToolInto("listOrganizations", map[string]any{}, &result)

	var (
		found        bool
		gotName      string
		gotLegalName *string
	)

	for _, organization := range result.Organizations {
		if organization.ID != owner.GetOrganizationID().String() {
			continue
		}

		found = true
		gotName = organization.Name
		gotLegalName = organization.LegalName

		break
	}

	require.True(t, found)
	assert.NotEmpty(t, gotName)
	require.NotNil(t, gotLegalName)
	assert.Equal(t, legalName, *gotLegalName)
}
