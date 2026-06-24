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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_OrganizationContext(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get
	var getResult struct {
		OrganizationContext struct {
			ID string `json:"id"`
		} `json:"organizationContext"`
	}
	mc.CallToolInto("getOrganizationContext", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	assert.NotEmpty(t, getResult.OrganizationContext.ID)

	// Update
	var updateResult struct {
		OrganizationContext struct {
			ID string `json:"id"`
		} `json:"organizationContext"`
	}
	mc.CallToolInto("updateOrganizationContext", map[string]any{
		"id":               getResult.OrganizationContext.ID,
		"companyLegalName": "Test Company LLC",
	}, &updateResult)
	assert.Equal(t, getResult.OrganizationContext.ID, updateResult.OrganizationContext.ID)
}
