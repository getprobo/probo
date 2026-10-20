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
	"go.probo.inc/probo/e2e/internal/testutil"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
)

func TestMCP_GCPConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var setupResult struct {
		Setup struct {
			Issuer                      string `json:"issuer"`
			Audience                    string `json:"audience"`
			Subject                     string `json:"subject"`
			SuggestedServiceAccountName string `json:"suggested_service_account_name"`
			TerraformSnippet            string `json:"terraform_snippet"`
		} `json:"setup"`
	}
	mc.CallToolInto("gcpConnectorSetup", map[string]any{
		"organization_id": orgID,
	}, &setupResult)
	assert.Contains(t, setupResult.Setup.Issuer, orgID)
	assert.Equal(t, cloudgcp.AudienceTemplate, setupResult.Setup.Audience)
	assert.Equal(t, orgID, setupResult.Setup.Subject)
	assert.Equal(t, cloudgcp.DefaultServiceAccountName, setupResult.Setup.SuggestedServiceAccountName)
	assert.Contains(t, setupResult.Setup.TerraformSnippet, setupResult.Setup.Issuer)
	assert.Contains(t, setupResult.Setup.TerraformSnippet, cloudgcp.DefaultTerraformModuleSource)
	assert.Contains(t, setupResult.Setup.TerraformSnippet, cloudgcp.DefaultServiceAccountName)
}

func TestMCP_GCPConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)
	orgID := owner.GetOrganizationID().String()

	msg := viewerMC.CallToolExpectToolError("gcpConnectorSetup", map[string]any{
		"organization_id": orgID,
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_GCPConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org2MC := testutil.NewMCPClient(t, org2)
	org1ID := org1.GetOrganizationID().String()

	msg := org2MC.CallToolExpectToolError("gcpConnectorSetup", map[string]any{
		"organization_id": org1ID,
	})
	assert.NotEmpty(t, msg)
}
