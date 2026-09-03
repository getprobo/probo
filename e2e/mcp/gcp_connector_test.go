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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
)

const (
	gcpFixtureProviderResource = "projects/123456789012/locations/global/workloadIdentityPools/probo-pool/providers/probo"
	gcpFixtureServiceAccount   = "probo-audit@example-project.iam.gserviceaccount.com"
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

	tr := mc.CallTool("createWorkloadIdentityConnector", map[string]any{
		"organization_id":                orgID,
		"provider":                       "GCP",
		"gcp_workload_identity_provider": gcpFixtureProviderResource,
		"gcp_service_account_email":      gcpFixtureServiceAccount,
	})
	require.False(t, tr.IsError, "createWorkloadIdentityConnector returned error: %v", tr.Content)
	require.NotEmpty(t, tr.Content)

	var payload string

	require.NoError(t, json.Unmarshal(tr.Content[0].Text, &payload))

	var createResult struct {
		Connector struct {
			ID               string `json:"id"`
			Provider         string `json:"provider"`
			Protocol         string `json:"protocol"`
			ConnectionStatus string `json:"connection_status"`
		} `json:"connector"`
	}

	require.NoError(t, json.Unmarshal([]byte(payload), &createResult))
	require.NotEmpty(t, createResult.Connector.ID)
	assert.Equal(t, "GCP", createResult.Connector.Provider)
	assert.Equal(t, "WORKLOAD_IDENTITY", createResult.Connector.Protocol)
	assert.Equal(t, "DISCONNECTED", createResult.Connector.ConnectionStatus)

	assert.NotContains(t, payload, gcpFixtureProviderResource)
	assert.NotContains(t, payload, gcpFixtureServiceAccount)
	assert.NotContains(t, strings.ToLower(payload), "permissiondenied")
	assert.NotContains(t, strings.ToLower(payload), "accessdenied")
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

	msg = viewerMC.CallToolExpectToolError("createWorkloadIdentityConnector", map[string]any{
		"organization_id":                orgID,
		"provider":                       "GCP",
		"gcp_workload_identity_provider": gcpFixtureProviderResource,
		"gcp_service_account_email":      gcpFixtureServiceAccount,
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

	msg = org2MC.CallToolExpectToolError("createWorkloadIdentityConnector", map[string]any{
		"organization_id":                org1ID,
		"provider":                       "GCP",
		"gcp_workload_identity_provider": gcpFixtureProviderResource,
		"gcp_service_account_email":      gcpFixtureServiceAccount,
	})
	assert.NotEmpty(t, msg)
}
