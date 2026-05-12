// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type rightsRequest struct {
	ID           string  `json:"id"`
	RequestType  string  `json:"requestType"`
	RequestState string  `json:"requestState"`
	DataSubject  string  `json:"dataSubject"`
	Contact      *string `json:"contact"`
	Details      *string `json:"details"`
}

func TestMCP_AddRightsRequest(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var result struct {
		RightsRequest rightsRequest `json:"rightsRequest"`
	}
	mc.CallToolInto("addRightsRequest", map[string]any{
		"organizationId": orgID,
		"requestType":    "ACCESS",
		"requestState":   "TODO",
		"dataSubject":    "John Doe",
		"contact":        "john@example.com",
		"details":        "Request for data access",
	}, &result)

	assert.NotEmpty(t, result.RightsRequest.ID)
	assert.Equal(t, "ACCESS", result.RightsRequest.RequestType)
	assert.Equal(t, "TODO", result.RightsRequest.RequestState)
	assert.Equal(t, "John Doe", result.RightsRequest.DataSubject)
}

func TestMCP_GetRightsRequest(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		RightsRequest rightsRequest `json:"rightsRequest"`
	}
	mc.CallToolInto("addRightsRequest", map[string]any{
		"organizationId": orgID,
		"requestType":    "DELETION",
		"requestState":   "TODO",
		"dataSubject":    "Jane Doe",
	}, &addResult)
	require.NotEmpty(t, addResult.RightsRequest.ID)

	// Get
	var getResult struct {
		RightsRequest rightsRequest `json:"rightsRequest"`
	}
	mc.CallToolInto("getRightsRequest", map[string]any{
		"id": addResult.RightsRequest.ID,
	}, &getResult)

	assert.Equal(t, addResult.RightsRequest.ID, getResult.RightsRequest.ID)
	assert.Equal(t, "DELETION", getResult.RightsRequest.RequestType)
	assert.Equal(t, "Jane Doe", getResult.RightsRequest.DataSubject)
}

func TestMCP_UpdateRightsRequest(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		RightsRequest rightsRequest `json:"rightsRequest"`
	}
	mc.CallToolInto("addRightsRequest", map[string]any{
		"organizationId": orgID,
		"requestType":    "ACCESS",
		"requestState":   "TODO",
		"dataSubject":    "Test Subject",
	}, &addResult)
	require.NotEmpty(t, addResult.RightsRequest.ID)

	// Update
	var updateResult struct {
		RightsRequest rightsRequest `json:"rightsRequest"`
	}
	mc.CallToolInto("updateRightsRequest", map[string]any{
		"id":           addResult.RightsRequest.ID,
		"requestState": "IN_PROGRESS",
		"dataSubject":  "Updated Subject",
	}, &updateResult)

	assert.Equal(t, addResult.RightsRequest.ID, updateResult.RightsRequest.ID)
	assert.Equal(t, "IN_PROGRESS", updateResult.RightsRequest.RequestState)
	assert.Equal(t, "Updated Subject", updateResult.RightsRequest.DataSubject)
}

func TestMCP_DeleteRightsRequest(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		RightsRequest rightsRequest `json:"rightsRequest"`
	}
	mc.CallToolInto("addRightsRequest", map[string]any{
		"organizationId": orgID,
		"requestType":    "PORTABILITY",
		"requestState":   "TODO",
		"dataSubject":    "Delete Subject",
	}, &addResult)
	require.NotEmpty(t, addResult.RightsRequest.ID)

	// Delete
	var deleteResult struct {
		DeletedRightsRequestID string `json:"deletedRightsRequestId"`
	}
	mc.CallToolInto("deleteRightsRequest", map[string]any{
		"id": addResult.RightsRequest.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.RightsRequest.ID, deleteResult.DeletedRightsRequestID)
}

func TestMCP_ListRightsRequests(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create multiple rights requests
	for _, reqType := range []string{"ACCESS", "DELETION", "PORTABILITY"} {
		var result struct {
			RightsRequest rightsRequest `json:"rightsRequest"`
		}
		mc.CallToolInto("addRightsRequest", map[string]any{
			"organizationId": orgID,
			"requestType":    reqType,
			"requestState":   "TODO",
			"dataSubject":    factory.SafeName("Subject"),
		}, &result)
		require.NotEmpty(t, result.RightsRequest.ID)
	}

	// List
	var listResult struct {
		RightsRequests []rightsRequest `json:"rightsRequests"`
	}
	mc.CallToolInto("listRightsRequests", map[string]any{
		"organizationId": orgID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.RightsRequests), 3)
}

func TestMCP_RightsRequest_Types(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	for _, reqType := range []string{"ACCESS", "DELETION", "PORTABILITY"} {
		t.Run(reqType, func(t *testing.T) {
			var result struct {
				RightsRequest rightsRequest `json:"rightsRequest"`
			}
			mc.CallToolInto("addRightsRequest", map[string]any{
				"organizationId": orgID,
				"requestType":    reqType,
				"requestState":   "TODO",
				"dataSubject":    factory.SafeName("Subject"),
			}, &result)

			assert.Equal(t, reqType, result.RightsRequest.RequestType)
		})
	}
}
