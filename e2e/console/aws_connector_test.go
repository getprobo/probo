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

package console_test

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

const (
	awsFixtureRoleARN = "arn:aws:iam::123456789012:role/ProboAudit"
	awsSecondRoleARN  = "arn:aws:iam::111111111111:role/ProboAudit"
)

const awsConnectorSetupQuery = `
	query($organizationId: ID!) {
		awsConnectorSetup(organizationId: $organizationId) {
			issuer
			audience
			subject
			suggestedRoleName
			terraformSnippet
			cloudFormationQuickCreateURL
		}
	}
`

const createWorkloadIdentityConnectorMutation = `
	mutation($input: CreateWorkloadIdentityConnectorInput!) {
		createWorkloadIdentityConnector(input: $input) {
			connector {
				id
				provider
				protocol
			}
		}
	}
`

// Selected apart from createWorkloadIdentityConnectorMutation: connectionStatus
// probes the provider live, so only the tests that assert on it should pay for
// the round trip.
const createWorkloadIdentityConnectorWithStatusMutation = `
	mutation($input: CreateWorkloadIdentityConnectorInput!) {
		createWorkloadIdentityConnector(input: $input) {
			connector {
				id
				connectionStatus
			}
		}
	}
`

const organizationConnectorStatusQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Organization {
				connectors {
					id
					connectionStatus
				}
			}
		}
	}
`

type awsConnectorSetupResult struct {
	AWSConnectorSetup struct {
		Issuer                       string `json:"issuer"`
		Audience                     string `json:"audience"`
		Subject                      string `json:"subject"`
		SuggestedRoleName            string `json:"suggestedRoleName"`
		TerraformSnippet             string `json:"terraformSnippet"`
		CloudFormationQuickCreateURL string `json:"cloudFormationQuickCreateURL"`
	} `json:"awsConnectorSetup"`
}

type createWorkloadIdentityConnectorResult struct {
	CreateWorkloadIdentityConnector struct {
		Connector struct {
			ID       string `json:"id"`
			Provider string `json:"provider"`
			Protocol string `json:"protocol"`
		} `json:"connector"`
	} `json:"createWorkloadIdentityConnector"`
}

type createWorkloadIdentityConnectorWithStatusResult struct {
	CreateWorkloadIdentityConnector struct {
		Connector struct {
			ID               string `json:"id"`
			ConnectionStatus string `json:"connectionStatus"`
		} `json:"connector"`
	} `json:"createWorkloadIdentityConnector"`
}

type organizationConnectorStatusResult struct {
	Node struct {
		Connectors []struct {
			ID               string `json:"id"`
			ConnectionStatus string `json:"connectionStatus"`
		} `json:"connectors"`
	} `json:"node"`
}

func TestAWSConnectorSetup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result awsConnectorSetupResult

	err := owner.Execute(awsConnectorSetupQuery, map[string]any{
		"organizationId": orgID,
	}, &result)
	require.NoError(t, err)

	setup := result.AWSConnectorSetup
	assert.Contains(t, setup.Issuer, orgID)
	assert.Equal(t, identityfederation.AudienceAWS, setup.Audience)
	assert.Equal(t, orgID, setup.Subject)
	assert.Equal(t, coredata.DefaultAWSRoleName, setup.SuggestedRoleName)
	assert.Contains(t, setup.TerraformSnippet, setup.Issuer)
	assert.Contains(t, setup.TerraformSnippet, setup.Subject)
	assert.Contains(t, setup.TerraformSnippet, cloudaws.DefaultTerraformModuleSource)
	assert.Contains(t, setup.TerraformSnippet, coredata.DefaultAWSRoleName)
	assert.Contains(t, setup.CloudFormationQuickCreateURL, url.QueryEscape(setup.Issuer))
	assert.Contains(t, setup.CloudFormationQuickCreateURL, url.QueryEscape(coredata.DefaultAWSRoleName))
}

func TestCreateWorkloadIdentityConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result createWorkloadIdentityConnectorResult

	err := owner.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"provider":       "AWS",
			"awsRoleArn":     awsFixtureRoleARN,
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateWorkloadIdentityConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "AWS", connector.Provider)
	assert.Equal(t, "WORKLOAD_IDENTITY", connector.Protocol)

	t.Run("allows a second connector for the same provider", func(t *testing.T) {
		t.Parallel()

		var second createWorkloadIdentityConnectorResult

		err := owner.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
				"provider":       "AWS",
				"awsRoleArn":     awsSecondRoleARN,
			},
		}, &second)
		require.NoError(t, err)

		secondID := second.CreateWorkloadIdentityConnector.Connector.ID
		assert.NotEmpty(t, secondID)
		assert.NotEqual(t, connector.ID, secondID)
	})
}

func TestCreateWorkloadIdentityConnector_InvalidRoleARN(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "arn:aws:iam::123456789012:user/alice"

	err := owner.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"provider":       "AWS",
			"awsRoleArn":     bogus,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
	assert.NotContains(t, err.Error(), "alice")
	assert.NotContains(t, strings.ToLower(err.Error()), "arn:aws")
}

// A role Probo cannot assume is reported as DISCONNECTED rather than as an
// error, so the response now travels to the client on the failure path too:
// assert the redaction against the whole payload, not just an error message.
func TestConnectorConnectionStatus_AssumeRoleFailure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	resp, err := owner.Do(createWorkloadIdentityConnectorWithStatusMutation, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"provider":       "AWS",
			"awsRoleArn":     awsFixtureRoleARN,
		},
	})
	require.NoError(t, err)

	var created createWorkloadIdentityConnectorWithStatusResult
	require.NoError(t, json.Unmarshal(resp.Data, &created))

	connector := created.CreateWorkloadIdentityConnector.Connector
	require.NotEmpty(t, connector.ID)
	assert.Equal(t, "DISCONNECTED", connector.ConnectionStatus)

	payload := resp.DataString()
	assert.NotContains(t, payload, awsFixtureRoleARN)
	assert.NotContains(t, payload, "arn:aws")
	assert.NotContains(t, strings.ToLower(payload), "accessdenied")
}

func TestAWSConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID := owner.GetOrganizationID().String()

	t.Run("viewer cannot read setup", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(awsConnectorSetupQuery, map[string]any{
			"organizationId": orgID,
		}, &awsConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read aws connector setup")
	})

	t.Run("viewer cannot create connector", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
				"provider":       "AWS",
				"awsRoleArn":     awsFixtureRoleARN,
			},
		}, &createWorkloadIdentityConnectorResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create aws connector")
	})

	// A viewer may list connectors but not read one, so the status field is
	// what it is refused, not the listing around it.
	t.Run("viewer cannot read connection status", func(t *testing.T) {
		t.Parallel()

		var created createWorkloadIdentityConnectorResult

		err := owner.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
				"provider":       "AWS",
				"awsRoleArn":     awsFixtureRoleARN,
			},
		}, &created)
		require.NoError(t, err)

		err = viewer.Execute(organizationConnectorStatusQuery, map[string]any{
			"id": orgID,
		}, &organizationConnectorStatusResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read aws connector status")
	})
}

func TestAWSConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org1ID := org1.GetOrganizationID().String()

	err := org1.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": org1ID,
			"provider":       "AWS",
			"awsRoleArn":     awsFixtureRoleARN,
		},
	}, &createWorkloadIdentityConnectorResult{})
	require.NoError(t, err)

	t.Run("cannot read setup for another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(awsConnectorSetupQuery, map[string]any{
			"organizationId": org1ID,
		}, &awsConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "org B should not read org A's aws connector setup")
	})

	t.Run("cannot create connector in another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(createWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId": org1ID,
				"provider":       "AWS",
				"awsRoleArn":     awsFixtureRoleARN,
			},
		}, &createWorkloadIdentityConnectorResult{})
		testutil.RequireForbiddenError(t, err, "org B should not create a connector in org A")
	})

	t.Run("cannot read connection status from another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(organizationConnectorStatusQuery, map[string]any{
			"id": org1ID,
		}, &organizationConnectorStatusResult{})
		testutil.RequireForbiddenError(t, err, "org B should not read org A's connector status")
	})
}
