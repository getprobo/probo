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
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

const awsFixtureRoleARN = "arn:aws:iam::123456789012:role/ProboAudit"

func TestMCP_AWSConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var setupResult struct {
		Setup struct {
			Issuer                       string `json:"issuer"`
			Audience                     string `json:"audience"`
			Subject                      string `json:"subject"`
			SuggestedRoleName            string `json:"suggested_role_name"`
			TerraformSnippet             string `json:"terraform_snippet"`
			CloudFormationQuickCreateURL string `json:"cloud_formation_quick_create_url"`
		} `json:"setup"`
	}
	mc.CallToolInto("awsConnectorSetup", map[string]any{
		"organization_id": orgID,
	}, &setupResult)
	assert.Contains(t, setupResult.Setup.Issuer, orgID)
	assert.Equal(t, identityfederation.AudienceAWS, setupResult.Setup.Audience)
	assert.Equal(t, orgID, setupResult.Setup.Subject)
	assert.Equal(t, coredata.DefaultAWSRoleName, setupResult.Setup.SuggestedRoleName)
	assert.Contains(t, setupResult.Setup.TerraformSnippet, setupResult.Setup.Issuer)
	assert.Contains(t, setupResult.Setup.TerraformSnippet, cloudaws.DefaultTerraformModuleSource)
	assert.Contains(t, setupResult.Setup.TerraformSnippet, coredata.DefaultAWSRoleName)
	assert.Contains(t, setupResult.Setup.CloudFormationQuickCreateURL, url.QueryEscape(setupResult.Setup.Issuer))
	assert.Contains(t, setupResult.Setup.CloudFormationQuickCreateURL, url.QueryEscape(coredata.DefaultAWSRoleName))

	// MCP has no lazy field resolution, so the create tool probes the role and
	// reports the verdict on the connector it returns. A role Probo cannot
	// assume is DISCONNECTED, not a tool error, so the payload reaches the
	// client on the failure path and has to stay free of AWS detail.
	tr := mc.CallTool("createWorkloadIdentityConnector", map[string]any{
		"organization_id": orgID,
		"provider":        "AWS",
		"aws_role_arn":    awsFixtureRoleARN,
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
	assert.Equal(t, "AWS", createResult.Connector.Provider)
	assert.Equal(t, "WORKLOAD_IDENTITY", createResult.Connector.Protocol)
	assert.Equal(t, "DISCONNECTED", createResult.Connector.ConnectionStatus)

	assert.NotContains(t, payload, awsFixtureRoleARN)
	assert.NotContains(t, payload, "arn:aws")
	assert.NotContains(t, strings.ToLower(payload), "accessdenied")
}

func TestMCP_AWSConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)
	orgID := owner.GetOrganizationID().String()

	msg := viewerMC.CallToolExpectToolError("awsConnectorSetup", map[string]any{
		"organization_id": orgID,
	})
	assert.Contains(t, msg, "permission denied")

	// Connection status rides on the connector the create tool returns, and MCP
	// exposes no other way to reach a connector, so refusing create is what
	// keeps a viewer away from the status too.
	msg = viewerMC.CallToolExpectToolError("createWorkloadIdentityConnector", map[string]any{
		"organization_id": orgID,
		"provider":        "AWS",
		"aws_role_arn":    awsFixtureRoleARN,
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_AWSConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org2MC := testutil.NewMCPClient(t, org2)
	org1ID := org1.GetOrganizationID().String()

	msg := org2MC.CallToolExpectToolError("awsConnectorSetup", map[string]any{
		"organization_id": org1ID,
	})
	assert.NotEmpty(t, msg)

	msg = org2MC.CallToolExpectToolError("createWorkloadIdentityConnector", map[string]any{
		"organization_id": org1ID,
		"provider":        "AWS",
		"aws_role_arn":    awsFixtureRoleARN,
	})
	assert.NotEmpty(t, msg)
}
