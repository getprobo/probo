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

func TestMCP_AuditProgram_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	frameworkID := factory.CreateFramework(owner)

	var addResult struct {
		AuditProgram struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"auditProgram"`
	}
	mc.CallToolInto("addAuditProgram", map[string]any{
		"organizationId": orgID,
		"frameworkId":    frameworkID,
		"name":           factory.SafeName("AuditProgram"),
	}, &addResult)
	require.NotEmpty(t, addResult.AuditProgram.ID)

	var getResult struct {
		AuditProgram struct {
			ID string `json:"id"`
		} `json:"auditProgram"`
	}
	mc.CallToolInto("getAuditProgram", map[string]any{
		"id": addResult.AuditProgram.ID,
	}, &getResult)
	assert.Equal(t, addResult.AuditProgram.ID, getResult.AuditProgram.ID)

	var updateResult struct {
		AuditProgram struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"auditProgram"`
	}
	mc.CallToolInto("updateAuditProgram", map[string]any{
		"id":   addResult.AuditProgram.ID,
		"name": "Updated Audit Program",
	}, &updateResult)
	assert.Equal(t, "Updated Audit Program", updateResult.AuditProgram.Name)

	var listResult struct {
		AuditPrograms []struct {
			ID string `json:"id"`
		} `json:"auditPrograms"`
	}
	mc.CallToolInto("listAuditPrograms", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.AuditPrograms)

	var addAuditResult struct {
		Audit struct {
			ID             string  `json:"id"`
			AuditProgramID *string `json:"auditProgramId"`
		} `json:"audit"`
	}
	mc.CallToolInto("addAudit", map[string]any{
		"organizationId": orgID,
		"frameworkId":    frameworkID,
		"auditProgramId": addResult.AuditProgram.ID,
		"name":           factory.SafeName("LinkedAudit"),
	}, &addAuditResult)
	require.NotEmpty(t, addAuditResult.Audit.ID)
	require.NotNil(t, addAuditResult.Audit.AuditProgramID)
	assert.Equal(t, addResult.AuditProgram.ID, *addAuditResult.Audit.AuditProgramID)

	var deleteResult struct {
		DeletedAuditProgramID string `json:"deletedAuditProgramId"`
	}
	mc.CallToolInto("deleteAuditProgram", map[string]any{
		"id": addResult.AuditProgram.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.AuditProgram.ID, deleteResult.DeletedAuditProgramID)

	var getAuditResult struct {
		Audit struct {
			ID             string  `json:"id"`
			AuditProgramID *string `json:"auditProgramId"`
		} `json:"audit"`
	}
	mc.CallToolInto("getAudit", map[string]any{
		"id": addAuditResult.Audit.ID,
	}, &getAuditResult)
	assert.Equal(t, addAuditResult.Audit.ID, getAuditResult.Audit.ID)
	assert.Nil(t, getAuditResult.Audit.AuditProgramID)
}
