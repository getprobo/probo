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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_Obligation_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Obligation struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"obligation"`
	}
	mc.CallToolInto("addObligation", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("Obligation"),
		"description":    "Test obligation",
	}, &addResult)
	require.NotEmpty(t, addResult.Obligation.ID)

	// Get
	var getResult struct {
		Obligation struct {
			ID string `json:"id"`
		} `json:"obligation"`
	}
	mc.CallToolInto("getObligation", map[string]any{
		"id": addResult.Obligation.ID,
	}, &getResult)
	assert.Equal(t, addResult.Obligation.ID, getResult.Obligation.ID)

	// Update
	var updateResult struct {
		Obligation struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"obligation"`
	}
	mc.CallToolInto("updateObligation", map[string]any{
		"id":   addResult.Obligation.ID,
		"name": "Updated Obligation",
	}, &updateResult)
	assert.Equal(t, "Updated Obligation", updateResult.Obligation.Name)

	// List
	var listResult struct {
		Obligations []struct {
			ID string `json:"id"`
		} `json:"obligations"`
	}
	mc.CallToolInto("listObligations", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Obligations)

	// Delete
	var deleteResult struct {
		DeletedObligationID string `json:"deletedObligationId"`
	}
	mc.CallToolInto("deleteObligation", map[string]any{
		"id": addResult.Obligation.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Obligation.ID, deleteResult.DeletedObligationID)
}
