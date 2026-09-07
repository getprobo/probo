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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
)

const (
	gcpFixtureProviderResource = "projects/123456789012/locations/global/workloadIdentityPools/probo-pool/providers/probo"
	gcpSecondProviderResource  = "projects/111111111111/locations/global/workloadIdentityPools/probo-pool/providers/probo"
	gcpFixtureServiceAccount   = "probo-audit@example-project.iam.gserviceaccount.com"
)

const gcpConnectorSetupQuery = `
	query($organizationId: ID!) {
		gcpConnectorSetup(organizationId: $organizationId) {
			issuer
			audience
			subject
			suggestedServiceAccountName
			terraformSnippet
		}
	}
`

const createGCPWorkloadIdentityConnectorMutation = `
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

const createGCPWorkloadIdentityConnectorWithStatusMutation = `
	mutation($input: CreateWorkloadIdentityConnectorInput!) {
		createWorkloadIdentityConnector(input: $input) {
			connector {
				id
				connectionStatus
			}
		}
	}
`

type gcpConnectorSetupResult struct {
	GCPConnectorSetup struct {
		Issuer                      string `json:"issuer"`
		Audience                    string `json:"audience"`
		Subject                     string `json:"subject"`
		SuggestedServiceAccountName string `json:"suggestedServiceAccountName"`
		TerraformSnippet            string `json:"terraformSnippet"`
	} `json:"gcpConnectorSetup"`
}

func TestGCPConnectorSetup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result gcpConnectorSetupResult

	err := owner.Execute(gcpConnectorSetupQuery, map[string]any{
		"organizationId": orgID,
	}, &result)
	require.NoError(t, err)

	setup := result.GCPConnectorSetup
	assert.Contains(t, setup.Issuer, orgID)
	assert.Equal(t, cloudgcp.AudienceTemplate, setup.Audience)
	assert.Equal(t, orgID, setup.Subject)
	assert.Equal(t, cloudgcp.DefaultServiceAccountName, setup.SuggestedServiceAccountName)
	assert.Contains(t, setup.TerraformSnippet, setup.Issuer)
	assert.Contains(t, setup.TerraformSnippet, setup.Subject)
	assert.Contains(t, setup.TerraformSnippet, cloudgcp.DefaultTerraformModuleSource)
	assert.Contains(t, setup.TerraformSnippet, cloudgcp.DefaultServiceAccountName)
}

func TestCreateGCPWorkloadIdentityConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result createWorkloadIdentityConnectorResult

	err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              orgID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateWorkloadIdentityConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "GCP", connector.Provider)
	assert.Equal(t, "WORKLOAD_IDENTITY", connector.Protocol)

	t.Run("allows a second connector for the same provider", func(t *testing.T) {
		t.Parallel()

		var second createWorkloadIdentityConnectorResult

		err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              orgID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpSecondProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
			},
		}, &second)
		require.NoError(t, err)

		secondID := second.CreateWorkloadIdentityConnector.Connector.ID
		assert.NotEmpty(t, secondID)
		assert.NotEqual(t, connector.ID, secondID)
	})
}

func TestCreateGCPWorkloadIdentityConnector_InvalidProviderResource(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "projects/alice/locations/global/workloadIdentityPools/probo-pool/providers/probo"

	err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              owner.GetOrganizationID().String(),
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": bogus,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
	assert.NotContains(t, err.Error(), "alice")
}

func TestCreateGCPWorkloadIdentityConnector_InvalidServiceAccountEmail(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "alice@example.com"

	err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              owner.GetOrganizationID().String(),
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      bogus,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
	assert.NotContains(t, err.Error(), "alice")
}

func TestGCPConnectorConnectionStatus_ImpersonationFailure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	resp, err := owner.Do(createGCPWorkloadIdentityConnectorWithStatusMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              orgID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	})
	require.NoError(t, err)

	var created createWorkloadIdentityConnectorWithStatusResult
	require.NoError(t, json.Unmarshal(resp.Data, &created))

	connector := created.CreateWorkloadIdentityConnector.Connector
	require.NotEmpty(t, connector.ID)
	assert.Equal(t, "DISCONNECTED", connector.ConnectionStatus)

	payload := resp.DataString()
	assert.NotContains(t, payload, gcpFixtureProviderResource)
	assert.NotContains(t, payload, gcpFixtureServiceAccount)
	assert.NotContains(t, strings.ToLower(payload), "permissiondenied")
	assert.NotContains(t, strings.ToLower(payload), "accessdenied")
}

func TestGCPConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID := owner.GetOrganizationID().String()

	t.Run("viewer cannot read setup", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(gcpConnectorSetupQuery, map[string]any{
			"organizationId": orgID,
		}, &gcpConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read gcp connector setup")
	})

	t.Run("viewer cannot create connector", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              orgID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
			},
		}, &createWorkloadIdentityConnectorResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create gcp connector")
	})

	t.Run("viewer cannot read connection status", func(t *testing.T) {
		t.Parallel()

		var created createWorkloadIdentityConnectorResult

		err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              orgID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
			},
		}, &created)
		require.NoError(t, err)

		err = viewer.Execute(organizationConnectorStatusQuery, map[string]any{
			"id": orgID,
		}, &organizationConnectorStatusResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read gcp connector status")
	})
}

func TestGCPConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org1ID := org1.GetOrganizationID().String()

	err := org1.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              org1ID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	}, &createWorkloadIdentityConnectorResult{})
	require.NoError(t, err)

	t.Run("cannot read setup for another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(gcpConnectorSetupQuery, map[string]any{
			"organizationId": org1ID,
		}, &gcpConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "org B should not read org A's gcp connector setup")
	})

	t.Run("cannot create connector in another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              org1ID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
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
