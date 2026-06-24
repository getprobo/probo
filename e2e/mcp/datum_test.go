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

func TestMCP_Datum_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	// Create
	var addResult struct {
		Datum struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"datum"`
	}
	mc.CallToolInto("addDatum", map[string]any{
		"organizationId":     orgID,
		"name":               factory.SafeName("Datum"),
		"ownerId":            profileID,
		"dataClassification": "PUBLIC",
	}, &addResult)
	require.NotEmpty(t, addResult.Datum.ID)

	// Get
	var getResult struct {
		Datum struct {
			ID string `json:"id"`
		} `json:"datum"`
	}
	mc.CallToolInto("getDatum", map[string]any{
		"id": addResult.Datum.ID,
	}, &getResult)
	assert.Equal(t, addResult.Datum.ID, getResult.Datum.ID)

	// Update
	var updateResult struct {
		Datum struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"datum"`
	}
	mc.CallToolInto("updateDatum", map[string]any{
		"id":   addResult.Datum.ID,
		"name": "Updated Datum",
	}, &updateResult)
	assert.Equal(t, "Updated Datum", updateResult.Datum.Name)

	// List
	var listResult struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	mc.CallToolInto("listData", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Data)

	// Delete
	var deleteResult struct {
		DeletedDatumID string `json:"deletedDatumId"`
	}
	mc.CallToolInto("deleteDatum", map[string]any{
		"id": addResult.Datum.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Datum.ID, deleteResult.DeletedDatumID)
}
