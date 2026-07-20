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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type commitmentGroup struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
}

type commitment struct {
	ID          string `json:"id"`
	GroupID     string `json:"group_id"`
	Icon        string `json:"icon"`
	Eyebrow     string `json:"eyebrow"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
}

func mcpTrustCenterID(t *testing.T, mc *testutil.MCPClient, orgID string) string {
	t.Helper()

	var getResult struct {
		TrustCenter struct {
			ID string `json:"id"`
		} `json:"trust_center"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	require.NotEmpty(t, getResult.TrustCenter.ID)

	return getResult.TrustCenter.ID
}

func TestMCP_AddCommitmentGroup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var result struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "Security Practices",
		"description":     "How we protect customer data",
	}, &result)

	assert.NotEmpty(t, result.CommitmentGroup.ID)
	assert.Equal(t, "Security Practices", result.CommitmentGroup.Title)
	assert.Equal(t, "How we protect customer data", result.CommitmentGroup.Description)
}

func TestMCP_UpdateCommitmentGroup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var addResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "Original Group",
		"description":     "Original description",
	}, &addResult)
	require.NotEmpty(t, addResult.CommitmentGroup.ID)

	var updateResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("updateCommitmentGroup", map[string]any{
		"id":    addResult.CommitmentGroup.ID,
		"title": "Updated Group",
	}, &updateResult)

	assert.Equal(t, addResult.CommitmentGroup.ID, updateResult.CommitmentGroup.ID)
	assert.Equal(t, "Updated Group", updateResult.CommitmentGroup.Title)
}

func TestMCP_DeleteCommitmentGroup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var addResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "Group to delete",
		"description":     "Temporary",
	}, &addResult)
	require.NotEmpty(t, addResult.CommitmentGroup.ID)

	var deleteResult struct {
		DeletedCommitmentGroupID string `json:"deleted_commitment_group_id"`
	}
	mc.CallToolInto("deleteCommitmentGroup", map[string]any{
		"id": addResult.CommitmentGroup.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.CommitmentGroup.ID, deleteResult.DeletedCommitmentGroupID)
}

func TestMCP_ListCommitmentGroups(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	for range 2 {
		var result struct {
			CommitmentGroup commitmentGroup `json:"commitment_group"`
		}
		mc.CallToolInto("addCommitmentGroup", map[string]any{
			"trust_center_id": tcID,
			"title":           factory.SafeName("Group"),
			"description":     factory.SafeName("Desc"),
		}, &result)
		require.NotEmpty(t, result.CommitmentGroup.ID)
	}

	var listResult struct {
		CommitmentGroups []commitmentGroup `json:"commitment_groups"`
	}
	mc.CallToolInto("listCommitmentGroups", map[string]any{
		"trust_center_id": tcID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.CommitmentGroups), 2)
}

func TestMCP_AddCommitment(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var groupResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "Encryption",
		"description":     "Encryption commitments",
	}, &groupResult)
	require.NotEmpty(t, groupResult.CommitmentGroup.ID)

	var result struct {
		Commitment commitment `json:"commitment"`
	}
	mc.CallToolInto("addCommitment", map[string]any{
		"group_id":    groupResult.CommitmentGroup.ID,
		"icon":        "LOCK_KEY",
		"eyebrow":     "Data at rest",
		"title":       "AES-256 encryption",
		"description": "All customer data is encrypted at rest",
	}, &result)

	assert.NotEmpty(t, result.Commitment.ID)
	assert.Equal(t, groupResult.CommitmentGroup.ID, result.Commitment.GroupID)
	assert.Equal(t, "LOCK_KEY", result.Commitment.Icon)
	assert.Equal(t, "AES-256 encryption", result.Commitment.Title)
}

func TestMCP_UpdateCommitment(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var groupResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "Access control",
		"description":     "Access commitments",
	}, &groupResult)
	require.NotEmpty(t, groupResult.CommitmentGroup.ID)

	var addResult struct {
		Commitment commitment `json:"commitment"`
	}
	mc.CallToolInto("addCommitment", map[string]any{
		"group_id":    groupResult.CommitmentGroup.ID,
		"icon":        "KEY",
		"eyebrow":     "SSO",
		"title":       "Original title",
		"description": "Original description",
	}, &addResult)
	require.NotEmpty(t, addResult.Commitment.ID)

	var updateResult struct {
		Commitment commitment `json:"commitment"`
	}
	mc.CallToolInto("updateCommitment", map[string]any{
		"id":    addResult.Commitment.ID,
		"title": "Updated title",
		"icon":  "FINGERPRINT",
	}, &updateResult)

	assert.Equal(t, addResult.Commitment.ID, updateResult.Commitment.ID)
	assert.Equal(t, "Updated title", updateResult.Commitment.Title)
	assert.Equal(t, "FINGERPRINT", updateResult.Commitment.Icon)
}

func TestMCP_DeleteCommitment(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var groupResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "Temporary group",
		"description":     "For delete test",
	}, &groupResult)
	require.NotEmpty(t, groupResult.CommitmentGroup.ID)

	var addResult struct {
		Commitment commitment `json:"commitment"`
	}
	mc.CallToolInto("addCommitment", map[string]any{
		"group_id":    groupResult.CommitmentGroup.ID,
		"icon":        "SHIELD_CHECK",
		"eyebrow":     "Compliance",
		"title":       "Commitment to delete",
		"description": "Temporary",
	}, &addResult)
	require.NotEmpty(t, addResult.Commitment.ID)

	var deleteResult struct {
		DeletedCommitmentID string `json:"deleted_commitment_id"`
	}
	mc.CallToolInto("deleteCommitment", map[string]any{
		"id": addResult.Commitment.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.Commitment.ID, deleteResult.DeletedCommitmentID)
}

func TestMCP_ListCommitments(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	tcID := mcpTrustCenterID(t, mc, owner.GetOrganizationID().String())

	var groupResult struct {
		CommitmentGroup commitmentGroup `json:"commitment_group"`
	}
	mc.CallToolInto("addCommitmentGroup", map[string]any{
		"trust_center_id": tcID,
		"title":           "List group",
		"description":     "For list test",
	}, &groupResult)
	require.NotEmpty(t, groupResult.CommitmentGroup.ID)

	for range 2 {
		var result struct {
			Commitment commitment `json:"commitment"`
		}
		mc.CallToolInto("addCommitment", map[string]any{
			"group_id":    groupResult.CommitmentGroup.ID,
			"icon":        "LOCK",
			"eyebrow":     factory.SafeName("Eye"),
			"title":       factory.SafeName("Title"),
			"description": factory.SafeName("Desc"),
		}, &result)
		require.NotEmpty(t, result.Commitment.ID)
	}

	var listResult struct {
		Commitments []commitment `json:"commitments"`
	}
	mc.CallToolInto("listCommitments", map[string]any{
		"group_id": groupResult.CommitmentGroup.ID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.Commitments), 2)
}
