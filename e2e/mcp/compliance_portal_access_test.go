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

type mcpCompliancePortalAccess struct {
	ID    string  `json:"id"`
	State string  `json:"state"`
	Auth  *string `json:"authenticated_at"`
}

func TestMCP_CreateCompliancePortalAccess(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	compliancePortalID := mcpCompliancePortalID(t, owner)
	email := factory.SafeEmail()

	var result struct {
		CompliancePortalAccess mcpCompliancePortalAccess `json:"compliance_portal_access"`
	}
	mc.CallToolInto("createCompliancePortalAccess", map[string]any{
		"compliance_portal_id": compliancePortalID,
		"email":                email,
	}, &result)

	assert.NotEmpty(t, result.CompliancePortalAccess.ID)
	assert.Equal(t, "ACTIVE", result.CompliancePortalAccess.State)
	assert.Nil(t, result.CompliancePortalAccess.Auth)

	errText := mc.CallToolExpectToolError("createCompliancePortalAccess", map[string]any{
		"compliance_portal_id": compliancePortalID,
		"email":                email,
	})
	require.Contains(t, errText, "already has access")
}

func TestMCP_CreateCompliancePortalAccessByProfileID(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	compliancePortalID := mcpCompliancePortalID(t, owner)
	profileID := factory.CreateUser(owner)

	var result struct {
		CompliancePortalAccess mcpCompliancePortalAccess `json:"compliance_portal_access"`
	}
	mc.CallToolInto("createCompliancePortalAccess", map[string]any{
		"compliance_portal_id": compliancePortalID,
		"profile_id":           profileID,
	}, &result)

	assert.NotEmpty(t, result.CompliancePortalAccess.ID)
	assert.Equal(t, "ACTIVE", result.CompliancePortalAccess.State)
}

func TestMCP_CreateCompliancePortalAccessTenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, org2Owner)
	org1CompliancePortalID := mcpCompliancePortalID(t, org1Owner)

	errText := mc.CallToolExpectToolError("createCompliancePortalAccess", map[string]any{
		"compliance_portal_id": org1CompliancePortalID,
		"email":                factory.SafeEmail(),
	})
	require.NotEmpty(t, errText)
}

func TestMCP_DeactivateActivateCompliancePortalAccess(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	compliancePortalID := mcpCompliancePortalID(t, owner)

	var created struct {
		CompliancePortalAccess mcpCompliancePortalAccess `json:"compliance_portal_access"`
	}
	mc.CallToolInto("createCompliancePortalAccess", map[string]any{
		"compliance_portal_id": compliancePortalID,
		"email":                factory.SafeEmail(),
	}, &created)
	require.NotEmpty(t, created.CompliancePortalAccess.ID)

	var deactivated struct {
		CompliancePortalAccess mcpCompliancePortalAccess `json:"compliance_portal_access"`
	}
	mc.CallToolInto("deactivateCompliancePortalAccess", map[string]any{
		"id": created.CompliancePortalAccess.ID,
	}, &deactivated)
	assert.Equal(t, "DEACTIVATED", deactivated.CompliancePortalAccess.State)

	mc.CallToolInto("deactivateCompliancePortalAccess", map[string]any{
		"id": created.CompliancePortalAccess.ID,
	}, &deactivated)
	assert.Equal(t, "DEACTIVATED", deactivated.CompliancePortalAccess.State)

	var activated struct {
		CompliancePortalAccess mcpCompliancePortalAccess `json:"compliance_portal_access"`
	}
	mc.CallToolInto("activateCompliancePortalAccess", map[string]any{
		"id": created.CompliancePortalAccess.ID,
	}, &activated)
	assert.Equal(t, "ACTIVE", activated.CompliancePortalAccess.State)

	mc.CallToolInto("activateCompliancePortalAccess", map[string]any{
		"id": created.CompliancePortalAccess.ID,
	}, &activated)
	assert.Equal(t, "ACTIVE", activated.CompliancePortalAccess.State)
}

func TestMCP_DeactivateCompliancePortalAccessTenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, org2Owner)
	org1CompliancePortalID := mcpCompliancePortalID(t, org1Owner)

	org1MCP := testutil.NewMCPClient(t, org1Owner)
	var created struct {
		CompliancePortalAccess mcpCompliancePortalAccess `json:"compliance_portal_access"`
	}
	org1MCP.CallToolInto("createCompliancePortalAccess", map[string]any{
		"compliance_portal_id": org1CompliancePortalID,
		"email":                factory.SafeEmail(),
	}, &created)

	errText := mc.CallToolExpectToolError("deactivateCompliancePortalAccess", map[string]any{
		"id": created.CompliancePortalAccess.ID,
	})
	require.NotEmpty(t, errText)
}
